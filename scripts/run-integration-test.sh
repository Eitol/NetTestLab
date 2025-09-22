#!/bin/bash

# NetTestLab Full Integration Test
# This script automatically:
# 0. Detects router architecture
# 1. Compiles OpenWRT package for the target architecture
# 2. Copies package to target device
# 3. Installs the package
# 4. Runs integration tests

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
VERBOSE=0
SKIP_BUILD=0
SKIP_DEPLOY=0
PROJECT_ROOT=""

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    cat << EOF
NetTestLab Full Integration Test

Usage: $0 --router-ip IP [OPTIONS]

Required:
  --router-ip IP          Target router IP address

Optional:
  --user USER             SSH username (default: root)
  --password PASS         SSH password
  --key-path PATH         SSH private key path
  --sdk-path PATH         OpenWrt SDK path (will auto-download if not provided)
  --skip-build           Skip package building
  --skip-deploy          Skip package deployment
  --verbose              Enable verbose output
  --help                 Show this help

Examples:
  $0 --router-ip 192.168.1.4
  $0 --router-ip 192.168.1.4 --password admin123
  $0 --router-ip 192.168.1.4 --key-path ~/.ssh/id_rsa

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --router-ip)
            ROUTER_IP="$2"
            shift 2
            ;;
        --user)
            ROUTER_USER="$2"
            shift 2
            ;;
        --password)
            ROUTER_PASSWORD="$2"
            shift 2
            ;;
        --key-path)
            SSH_KEY_PATH="$2"
            shift 2
            ;;
        --sdk-path)
            OPENWRT_SDK_PATH="$2"
            shift 2
            ;;
        --skip-build)
            SKIP_BUILD=1
            shift
            ;;
        --skip-deploy)
            SKIP_DEPLOY=1
            shift
            ;;
        --verbose)
            VERBOSE=1
            shift
            ;;
        --help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Validate required parameters
if [[ -z "$ROUTER_IP" ]]; then
    print_error "Router IP is required"
    show_usage
    exit 1
fi

# Find project root
find_project_root() {
    local dir="$(pwd)"
    while [[ "$dir" != "/" ]]; do
        if [[ -f "$dir/go.mod" ]] && grep -q "NetTestLab" "$dir/go.mod"; then
            echo "$dir"
            return 0
        fi
        dir="$(dirname "$dir")"
    done
    return 1
}

PROJECT_ROOT=$(find_project_root)
if [[ -z "$PROJECT_ROOT" ]]; then
    print_error "Could not find NetTestLab project root"
    exit 1
fi

print_status "Using project root: $PROJECT_ROOT"
cd "$PROJECT_ROOT"

# SSH command builder
build_ssh_cmd() {
    local cmd="$1"
    local ssh_opts="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
    
    if [[ -n "$SSH_KEY_PATH" ]]; then
        if [[ $VERBOSE -eq 1 ]]; then
            echo "ssh $ssh_opts -i \"$SSH_KEY_PATH\" \"$ROUTER_USER@$ROUTER_IP\" \"$cmd\""
        else
            echo "ssh $ssh_opts -i \"$SSH_KEY_PATH\" \"$ROUTER_USER@$ROUTER_IP\" \"$cmd\" 2>/dev/null"
        fi
    elif [[ -n "$ROUTER_PASSWORD" ]]; then
        if [[ $VERBOSE -eq 1 ]]; then
            echo "sshpass -p \"$ROUTER_PASSWORD\" ssh $ssh_opts \"$ROUTER_USER@$ROUTER_IP\" \"$cmd\""
        else
            echo "sshpass -p \"$ROUTER_PASSWORD\" ssh $ssh_opts \"$ROUTER_USER@$ROUTER_IP\" \"$cmd\" 2>/dev/null"
        fi
    else
        if [[ $VERBOSE -eq 1 ]]; then
            echo "ssh $ssh_opts \"$ROUTER_USER@$ROUTER_IP\" \"$cmd\""
        else
            echo "ssh $ssh_opts \"$ROUTER_USER@$ROUTER_IP\" \"$cmd\" 2>/dev/null"
        fi
    fi
}

# SCP command builder
build_scp_cmd() {
    local local_path="$1"
    local remote_path="$2"
    local scp_opts="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
    
    if [[ -n "$SSH_KEY_PATH" ]]; then
        if [[ $VERBOSE -eq 1 ]]; then
            echo "scp $scp_opts -i \"$SSH_KEY_PATH\" \"$local_path\" \"$ROUTER_USER@$ROUTER_IP:$remote_path\""
        else
            echo "scp $scp_opts -i \"$SSH_KEY_PATH\" \"$local_path\" \"$ROUTER_USER@$ROUTER_IP:$remote_path\" 2>/dev/null"
        fi
    elif [[ -n "$ROUTER_PASSWORD" ]]; then
        if [[ $VERBOSE -eq 1 ]]; then
            echo "sshpass -p \"$ROUTER_PASSWORD\" scp $scp_opts \"$local_path\" \"$ROUTER_USER@$ROUTER_IP:$remote_path\""
        else
            echo "sshpass -p \"$ROUTER_PASSWORD\" scp $scp_opts \"$local_path\" \"$ROUTER_USER@$ROUTER_IP:$remote_path\" 2>/dev/null"
        fi
    else
        if [[ $VERBOSE -eq 1 ]]; then
            echo "scp $scp_opts \"$local_path\" \"$ROUTER_USER@$ROUTER_IP:$remote_path\""
        else
            echo "scp $scp_opts \"$local_path\" \"$ROUTER_USER@$ROUTER_IP:$remote_path\" 2>/dev/null"
        fi
    fi
}

# Function to run SSH command
run_ssh() {
    local cmd="$1"
    local ssh_command=$(build_ssh_cmd "$cmd")
    if [[ $VERBOSE -eq 1 ]]; then
        print_status "Running: $ssh_command"
    fi
    eval "$ssh_command"
}

# Function to run SCP command
run_scp() {
    local local_path="$1"
    local remote_path="$2"
    local scp_command=$(build_scp_cmd "$local_path" "$remote_path")
    if [[ $VERBOSE -eq 1 ]]; then
        print_status "Running: $scp_command"
    fi
    eval "$scp_command"
}

# Check SSH connectivity
print_status "🔗 Testing SSH connectivity to $ROUTER_IP..."
if ! run_ssh "echo 'SSH OK'" >/dev/null 2>&1; then
    print_error "Cannot connect to router via SSH"
    print_error "Please check:"
    print_error "1. Router IP is correct: $ROUTER_IP"
    print_error "2. SSH is enabled on the router"
    print_error "3. Username is correct: $ROUTER_USER"
    print_error "4. Password or SSH key is correct"
    exit 1
fi
print_success "SSH connectivity verified"

# Step 0: Detect router architecture
print_status "🔍 Step 0: Detecting router architecture..."
ROUTER_ARCH=$(run_ssh "uname -m")
ROUTER_INFO=$(run_ssh "cat /etc/openwrt_release" | grep DISTRIB_RELEASE | cut -d'=' -f2 | tr -d '"')
ROUTER_TARGET=$(run_ssh "cat /etc/openwrt_release" | grep DISTRIB_TARGET | cut -d'=' -f2 | tr -d '"')

print_success "Router detected:"
echo "  Architecture: $ROUTER_ARCH"
echo "  OpenWrt version: $ROUTER_INFO"
echo "  Target: $ROUTER_TARGET"

# Map architecture to OpenWrt SDK architecture
case "$ROUTER_ARCH" in
    "aarch64")
        # Check the target to determine correct SDK
        if [[ "$ROUTER_TARGET" == *"mediatek"* ]]; then
            SDK_ARCH="aarch64_cortex-a53"
            SDK_TARGET="mediatek"
            SDK_SUBTARGET="mt7622"
        elif [[ "$ROUTER_TARGET" == *"bcm27xx"* ]]; then
            SDK_ARCH="aarch64_cortex-a72"
            SDK_TARGET="bcm27xx"
            SDK_SUBTARGET="bcm2711"
        else
            SDK_ARCH="aarch64_generic"
            SDK_TARGET="generic"
            SDK_SUBTARGET="generic"
        fi
        ;;
    "x86_64")
        SDK_ARCH="x86_64"
        SDK_TARGET="x86"
        SDK_SUBTARGET="64"
        ;;
    "mips")
        SDK_ARCH="mips_24kc"
        SDK_TARGET="ath79"
        SDK_SUBTARGET="generic"
        ;;
    "armv7l")
        SDK_ARCH="arm_cortex-a7_neon-vfpv4"
        SDK_TARGET="bcm27xx"
        SDK_SUBTARGET="bcm2709"
        ;;
    *)
        print_warning "Unknown architecture: $ROUTER_ARCH, defaulting to generic"
        SDK_ARCH="generic"
        SDK_TARGET="generic"
        SDK_SUBTARGET="generic"
        ;;
esac

print_status "Using SDK architecture: $SDK_ARCH"

# Step 1: Setup/Download OpenWrt SDK if needed
if [[ $SKIP_BUILD -eq 0 ]]; then
    print_status "🛠️  Step 1: Setting up OpenWrt SDK..."
    
    if [[ -z "$OPENWRT_SDK_PATH" ]]; then
        # Auto-download SDK
        OPENWRT_VERSION=$(echo "$ROUTER_INFO" | grep -o '^[0-9]*\.[0-9]*')
        if [[ -z "$OPENWRT_VERSION" ]]; then
            OPENWRT_VERSION="23.05.4"
            print_warning "Could not detect OpenWrt version, using default: $OPENWRT_VERSION"
        fi
        
        SDK_NAME="openwrt-sdk-${OPENWRT_VERSION}-${ROUTER_TARGET}-${SDK_SUBTARGET}_gcc-12.3.0_musl.Linux-x86_64"
        SDK_DIR="/tmp/$SDK_NAME"
        
        if [[ ! -d "$SDK_DIR" ]]; then
            print_status "Downloading OpenWrt SDK..."
            mkdir -p /tmp
            cd /tmp
            
            SDK_URL="https://downloads.openwrt.org/releases/${OPENWRT_VERSION}/targets/${ROUTER_TARGET}/${SDK_SUBTARGET}/${SDK_NAME}.tar.xz"
            print_status "Downloading from: $SDK_URL"
            
            if ! wget -q --show-progress "$SDK_URL"; then
                print_error "Failed to download SDK from $SDK_URL"
                print_error "You may need to specify --sdk-path manually"
                exit 1
            fi
            
            print_status "Extracting SDK..."
            tar xf "${SDK_NAME}.tar.xz"
            rm "${SDK_NAME}.tar.xz"
        fi
        
        OPENWRT_SDK_PATH="$SDK_DIR"
        cd "$PROJECT_ROOT"
    fi
    
    if [[ ! -d "$OPENWRT_SDK_PATH" ]]; then
        print_error "OpenWrt SDK not found at: $OPENWRT_SDK_PATH"
        exit 1
    fi
    
    print_success "Using OpenWrt SDK: $OPENWRT_SDK_PATH"
    
    # Setup feeds in SDK
    print_status "Setting up SDK feeds..."
    cd "$OPENWRT_SDK_PATH"
    if [[ ! -f ".feeds_setup_done" ]]; then
        ./scripts/feeds update -a >/dev/null 2>&1 || true
        ./scripts/feeds install -a >/dev/null 2>&1 || true
        touch .feeds_setup_done
    fi
    cd "$PROJECT_ROOT"
    
    # Step 1.5: Build the package
    print_status "📦 Step 1.5: Building OpenWrt package..."
    
    # Copy package to SDK
    PACKAGE_DIR="$OPENWRT_SDK_PATH/package/nettestlab"
    rm -rf "$PACKAGE_DIR"
    mkdir -p "$PACKAGE_DIR"
    cp -r openwrt/* "$PACKAGE_DIR/"
    
    # Build package
    cd "$OPENWRT_SDK_PATH"
    print_status "Compiling package (this may take several minutes)..."
    
    if [[ $VERBOSE -eq 1 ]]; then
        make package/nettestlab/compile V=s
    else
        make package/nettestlab/compile >/dev/null 2>&1
    fi
    
    # Find built package
    PACKAGE_FILE=$(find "$OPENWRT_SDK_PATH/bin/packages" -name "nettestlab*.ipk" | head -1)
    if [[ -z "$PACKAGE_FILE" ]]; then
        print_error "Built package not found"
        exit 1
    fi
    
    print_success "Package built: $PACKAGE_FILE"
    cd "$PROJECT_ROOT"
else
    print_status "⏭️  Skipping package build"
    # Try to find existing package
    PACKAGE_FILE=$(find . -name "nettestlab*.ipk" | head -1)
    if [[ -z "$PACKAGE_FILE" ]]; then
        print_error "No existing package found and build is skipped"
        exit 1
    fi
fi

# Step 2: Copy package to router
if [[ $SKIP_DEPLOY -eq 0 ]]; then
    print_status "📤 Step 2: Copying package to router..."
    
    REMOTE_PACKAGE="/tmp/nettestlab.ipk"
    run_scp "$PACKAGE_FILE" "$REMOTE_PACKAGE"
    print_success "Package copied to router"
    
    # Step 3: Install package
    print_status "⚙️  Step 3: Installing package on router..."
    
    # Remove existing package if installed
    print_status "Removing existing package (if any)..."
    run_ssh "opkg remove nettestlab" >/dev/null 2>&1 || true
    
    # Install new package
    print_status "Installing new package..."
    if ! run_ssh "opkg install $REMOTE_PACKAGE"; then
        print_error "Package installation failed"
        # Show opkg logs for debugging
        print_status "Package installation logs:"
        run_ssh "logread | grep opkg | tail -10" || true
        exit 1
    fi
    
    print_success "Package installed successfully"
    
    # Verify service is running
    print_status "Starting NetTestLab service..."
    run_ssh "/etc/init.d/nettestlab start" || true
    sleep 2
    
    # Check service status
    if run_ssh "/etc/init.d/nettestlab status" | grep -q "running"; then
        print_success "NetTestLab service is running"
    else
        print_warning "Service may not be running, checking logs..."
        run_ssh "logread | grep nettestlab | tail -5" || true
    fi
else
    print_status "⏭️  Skipping package deployment"
fi

# Step 4: Run integration tests
print_status "🧪 Step 4: Running integration tests..."

# Set environment variables for the test
export NETTESTLAB_ROUTER_IP="$ROUTER_IP"
export NETTESTLAB_ROUTER_USER="$ROUTER_USER"
if [[ -n "$ROUTER_PASSWORD" ]]; then
    export NETTESTLAB_ROUTER_PASSWORD="$ROUTER_PASSWORD"
fi
if [[ -n "$SSH_KEY_PATH" ]]; then
    export SSH_KEY_PATH="$SSH_KEY_PATH"
fi
if [[ -n "$OPENWRT_SDK_PATH" ]]; then
    export OPENWRT_SDK_PATH="$OPENWRT_SDK_PATH"
fi

# Run specific integration tests
print_status "Running gRPC connectivity test..."
if go test -v ./tests/integration/ -run TestOpenWrtIntegration/GRPCConnectivity; then
    print_success "✅ gRPC connectivity test passed"
else
    print_error "❌ gRPC connectivity test failed"
fi

print_status "Running profile management test..."
if go test -v ./tests/integration/ -run TestOpenWrtIntegration/ProfileManagement; then
    print_success "✅ Profile management test passed"
else
    print_error "❌ Profile management test failed"
fi

print_status "Running network control test..."
if go test -v ./tests/integration/ -run TestOpenWrtIntegration/NetworkControl; then
    print_success "✅ Network control test passed"
else
    print_error "❌ Network control test failed"
fi

print_status "Running system monitoring test..."
if go test -v ./tests/integration/ -run TestOpenWrtIntegration/SystemMonitoring; then
    print_success "✅ System monitoring test passed"
else
    print_error "❌ System monitoring test failed"
fi

# Final verification
print_status "🔍 Final verification..."
print_status "Checking service status one more time..."
SERVICE_STATUS=$(run_ssh "/etc/init.d/nettestlab status" || echo "not running")
print_status "Service status: $SERVICE_STATUS"

print_status "Checking if gRPC port is listening..."
GRPC_STATUS=$(run_ssh "netstat -ln | grep :8080" || echo "not listening")
if [[ "$GRPC_STATUS" != "not listening" ]]; then
    print_success "✅ gRPC server is listening on port 8080"
else
    print_warning "⚠️  gRPC server may not be listening"
fi

print_success "🎉 Full integration test completed!"
print_status "Summary:"
echo "  ✅ Router architecture detected: $ROUTER_ARCH"
if [[ $SKIP_BUILD -eq 0 ]]; then
    echo "  ✅ OpenWrt package built successfully"
fi
if [[ $SKIP_DEPLOY -eq 0 ]]; then
    echo "  ✅ Package deployed and installed"
fi
echo "  ✅ Integration tests executed"
echo ""
print_status "You can now use NetTestLab on your router!"
print_status "gRPC endpoint: $ROUTER_IP:8080"
print_status "Try: go run cmd/auto-wifi-test/main.go --server $ROUTER_IP:8080"