#!/bin/bash

# NetTestLab OpenWrt Integration Test Runner
# This script helps set up and run integration tests for NetTestLab on OpenWrt

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ROUTER_IP=""
ROUTER_USER="root"
ROUTER_PASSWORD=""
SSH_KEY_PATH=""
OPENWRT_SDK_PATH=""
TEST_PATTERN=""
VERBOSE=false
DRY_RUN=false
SKIP_BUILD=false
SKIP_DEPLOY=false

# Help function
show_help() {
    cat << EOF
NetTestLab OpenWrt Integration Test Runner

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -h, --help              Show this help message
    -i, --router-ip IP      Router IP address (required)
    -u, --user USER         SSH username (default: root)
    -p, --password PASS     SSH password
    -k, --key-path PATH     SSH private key path
    -s, --sdk-path PATH     OpenWrt SDK path (required)
    -t, --test PATTERN      Run specific test pattern
    -v, --verbose           Verbose output
    -n, --dry-run           Show what would be done without executing
    --skip-build            Skip package build step
    --skip-deploy           Skip package deployment step

EXAMPLES:
    # Basic test run
    $0 -i 192.168.1.10 -p admin123 -s ~/openwrt-sdk

    # Test with SSH key
    $0 -i 192.168.1.10 -k ~/.ssh/id_rsa -s ~/openwrt-sdk

    # Run specific test
    $0 -i 192.168.1.10 -p admin123 -s ~/openwrt-sdk -t GRPCConnectivity

    # Verbose mode
    $0 -i 192.168.1.10 -p admin123 -s ~/openwrt-sdk -v

ENVIRONMENT VARIABLES:
    NETTESTLAB_ROUTER_IP        Router IP address
    NETTESTLAB_ROUTER_USER      SSH username
    NETTESTLAB_ROUTER_PASSWORD  SSH password
    SSH_KEY_PATH                SSH private key path
    OPENWRT_SDK_PATH            OpenWrt SDK path

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -i|--router-ip)
            ROUTER_IP="$2"
            shift 2
            ;;
        -u|--user)
            ROUTER_USER="$2"
            shift 2
            ;;
        -p|--password)
            ROUTER_PASSWORD="$2"
            shift 2
            ;;
        -k|--key-path)
            SSH_KEY_PATH="$2"
            shift 2
            ;;
        -s|--sdk-path)
            OPENWRT_SDK_PATH="$2"
            shift 2
            ;;
        -t|--test)
            TEST_PATTERN="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -n|--dry-run)
            DRY_RUN=true
            shift
            ;;
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --skip-deploy)
            SKIP_DEPLOY=true
            shift
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Load from environment if not set via command line
ROUTER_IP=${ROUTER_IP:-$NETTESTLAB_ROUTER_IP}
ROUTER_USER=${ROUTER_USER:-$NETTESTLAB_ROUTER_USER}
ROUTER_PASSWORD=${ROUTER_PASSWORD:-$NETTESTLAB_ROUTER_PASSWORD}
SSH_KEY_PATH=${SSH_KEY_PATH:-$SSH_KEY_PATH}
OPENWRT_SDK_PATH=${OPENWRT_SDK_PATH:-$OPENWRT_SDK_PATH}

# Validate required parameters
if [[ -z "$ROUTER_IP" ]]; then
    echo -e "${RED}Error: Router IP is required${NC}"
    echo "Use -i/--router-ip or set NETTESTLAB_ROUTER_IP environment variable"
    exit 1
fi

if [[ -z "$OPENWRT_SDK_PATH" ]]; then
    echo -e "${RED}Error: OpenWrt SDK path is required${NC}"
    echo "Use -s/--sdk-path or set OPENWRT_SDK_PATH environment variable"
    exit 1
fi

# Validate SDK path
if [[ ! -d "$OPENWRT_SDK_PATH" ]]; then
    echo -e "${RED}Error: OpenWrt SDK path does not exist: $OPENWRT_SDK_PATH${NC}"
    exit 1
fi

# Validate authentication method
if [[ -z "$ROUTER_PASSWORD" && -z "$SSH_KEY_PATH" ]]; then
    echo -e "${YELLOW}Warning: No authentication method specified${NC}"
    echo "Either provide -p/--password or -k/--key-path"
fi

if [[ -n "$SSH_KEY_PATH" && ! -f "$SSH_KEY_PATH" ]]; then
    echo -e "${RED}Error: SSH key file does not exist: $SSH_KEY_PATH${NC}"
    exit 1
fi

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check dependencies
check_dependencies() {
    log_info "Checking dependencies..."
    
    local missing_deps=()
    
    # Check Go
    if ! command -v go &> /dev/null; then
        missing_deps+=("go")
    fi
    
    # Check SSH
    if ! command -v ssh &> /dev/null; then
        missing_deps+=("ssh")
    fi
    
    # Check SCP
    if ! command -v scp &> /dev/null; then
        missing_deps+=("scp")
    fi
    
    # Check sshpass if using password auth
    if [[ -n "$ROUTER_PASSWORD" ]] && ! command -v sshpass &> /dev/null; then
        log_warning "sshpass not found - password authentication may not work"
        log_info "Install sshpass: apt-get install sshpass (Ubuntu) or brew install sshpass (macOS)"
    fi
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        exit 1
    fi
    
    log_success "All dependencies found"
}

# Test connectivity
test_connectivity() {
    log_info "Testing router connectivity..."
    
    # Test ping
    if ping -c 1 -W 3 "$ROUTER_IP" &> /dev/null; then
        log_success "Router is reachable at $ROUTER_IP"
    else
        log_error "Cannot reach router at $ROUTER_IP"
        exit 1
    fi
    
    # Test SSH
    local ssh_cmd="ssh -o ConnectTimeout=5 -o BatchMode=yes"
    if [[ -n "$SSH_KEY_PATH" ]]; then
        ssh_cmd="$ssh_cmd -i $SSH_KEY_PATH"
    fi
    
    if [[ -n "$ROUTER_PASSWORD" ]]; then
        if command -v sshpass &> /dev/null; then
            ssh_cmd="sshpass -p $ROUTER_PASSWORD $ssh_cmd"
        else
            log_warning "Cannot test SSH with password - sshpass not available"
            return
        fi
    fi
    
    if $ssh_cmd "$ROUTER_USER@$ROUTER_IP" "echo 'SSH OK'" &> /dev/null; then
        log_success "SSH connection successful"
    else
        log_error "SSH connection failed"
        log_info "Check credentials, network connectivity, and SSH service"
        exit 1
    fi
}

# Show configuration
show_config() {
    log_info "Test Configuration:"
    echo "  Router IP: $ROUTER_IP"
    echo "  Router User: $ROUTER_USER"
    if [[ -n "$ROUTER_PASSWORD" ]]; then
        echo "  Authentication: Password"
    elif [[ -n "$SSH_KEY_PATH" ]]; then
        echo "  Authentication: SSH Key ($SSH_KEY_PATH)"
    else
        echo "  Authentication: Default (no credentials specified)"
    fi
    echo "  OpenWrt SDK: $OPENWRT_SDK_PATH"
    if [[ -n "$TEST_PATTERN" ]]; then
        echo "  Test Pattern: $TEST_PATTERN"
    fi
    echo "  Skip Build: $SKIP_BUILD"
    echo "  Skip Deploy: $SKIP_DEPLOY"
    echo ""
}

# Set environment variables
set_environment() {
    export NETTESTLAB_ROUTER_IP="$ROUTER_IP"
    export NETTESTLAB_ROUTER_USER="$ROUTER_USER"
    
    if [[ -n "$ROUTER_PASSWORD" ]]; then
        export NETTESTLAB_ROUTER_PASSWORD="$ROUTER_PASSWORD"
    fi
    
    if [[ -n "$SSH_KEY_PATH" ]]; then
        export SSH_KEY_PATH="$SSH_KEY_PATH"
    fi
    
    export OPENWRT_SDK_PATH="$OPENWRT_SDK_PATH"
    
    # Set test-specific variables
    if [[ "$SKIP_BUILD" == "true" ]]; then
        export NETTESTLAB_SKIP_BUILD="true"
    fi
    
    if [[ "$SKIP_DEPLOY" == "true" ]]; then
        export NETTESTLAB_SKIP_DEPLOY="true"
    fi
}

# Run tests
run_tests() {
    log_info "Running integration tests..."
    
    local test_args="-v"
    
    if [[ -n "$TEST_PATTERN" ]]; then
        test_args="$test_args -run TestOpenWrtIntegration/$TEST_PATTERN"
    fi
    
    local go_test_cmd="go test $test_args ./tests/integration/"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "Dry run - would execute: $go_test_cmd"
        return
    fi
    
    if [[ "$VERBOSE" == "true" ]]; then
        echo ""
        log_info "Executing: $go_test_cmd"
        echo ""
    fi
    
    if $go_test_cmd; then
        echo ""
        log_success "All integration tests passed!"
    else
        echo ""
        log_error "Integration tests failed!"
        exit 1
    fi
}

# Main execution
main() {
    echo -e "${BLUE}NetTestLab OpenWrt Integration Test Runner${NC}"
    echo ""
    
    show_config
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "Dry run mode - showing what would be executed"
        echo ""
    fi
    
    check_dependencies
    test_connectivity
    set_environment
    run_tests
    
    echo ""
    log_success "Integration test execution completed!"
}

# Run main function
main "$@"