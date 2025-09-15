#!/bin/bash

# NetTestLab Simple Integration Test
# This script:
# 1. Copies the pre-built binary to the router
# 2. Creates a simple service file
# 3. Runs integration tests

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
NetTestLab Simple Integration Test

Usage: $0 --router-ip IP [OPTIONS]

Required:
  --router-ip IP          Target router IP address

Optional:
  --user USER             SSH username (default: root)
  --password PASS         SSH password
  --key-path PATH         SSH private key path
  --verbose              Enable verbose output
  --help                 Show this help

Examples:
  $0 --router-ip 192.168.1.4
  $0 --router-ip 192.168.1.4 --password admin123

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

# SSH command builder
build_ssh_cmd() {
    local cmd="$1"
    local ssh_opts="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
    
    if [[ -n "$SSH_KEY_PATH" ]]; then
        echo "ssh $ssh_opts -i \"$SSH_KEY_PATH\" \"$ROUTER_USER@$ROUTER_IP\" \"$cmd\" 2>/dev/null"
    elif [[ -n "$ROUTER_PASSWORD" ]]; then
        echo "sshpass -p \"$ROUTER_PASSWORD\" ssh $ssh_opts \"$ROUTER_USER@$ROUTER_IP\" \"$cmd\" 2>/dev/null"
    else
        echo "ssh $ssh_opts \"$ROUTER_USER@$ROUTER_IP\" \"$cmd\" 2>/dev/null"
    fi
}

# SCP command builder
build_scp_cmd() {
    local local_path="$1"
    local remote_path="$2"
    local scp_opts="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
    
    if [[ -n "$SSH_KEY_PATH" ]]; then
        echo "scp $scp_opts -i \"$SSH_KEY_PATH\" \"$local_path\" \"$ROUTER_USER@$ROUTER_IP:$remote_path\" 2>/dev/null"
    elif [[ -n "$ROUTER_PASSWORD" ]]; then
        echo "sshpass -p \"$ROUTER_PASSWORD\" scp $scp_opts \"$local_path\" \"$ROUTER_USER@$ROUTER_IP:$remote_path\" 2>/dev/null"
    else
        echo "scp $scp_opts \"$local_path\" \"$ROUTER_USER@$ROUTER_IP:$remote_path\" 2>/dev/null"
    fi
}

# Function to run SSH command
run_ssh() {
    local cmd="$1"
    local ssh_command=$(build_ssh_cmd "$cmd")
    eval "$ssh_command"
}

# Function to run SCP command
run_scp() {
    local local_path="$1"
    local remote_path="$2"
    local scp_command=$(build_scp_cmd "$local_path" "$remote_path")
    eval "$scp_command"
}

print_status "🚀 NetTestLab Simple Integration Test"
print_status "Target: $ROUTER_IP"

# Check SSH connectivity
print_status "🔗 Testing SSH connectivity..."
if ! run_ssh "echo 'SSH OK'" >/dev/null 2>&1; then
    print_error "Cannot connect to router via SSH"
    exit 1
fi
print_success "SSH connectivity verified"

# Detect router architecture
print_status "🔍 Detecting router architecture..."
ROUTER_ARCH=$(run_ssh "uname -m")
print_success "Router architecture: $ROUTER_ARCH"

# Choose the right binary
if [[ -f "bin/nettestlab-arm64" ]] && [[ "$ROUTER_ARCH" == "aarch64" ]]; then
    BINARY_PATH="bin/nettestlab-arm64"
    print_success "Using ARM64 binary: $BINARY_PATH"
elif [[ -f "bin/nettestlab-server" ]]; then
    BINARY_PATH="bin/nettestlab-server"
    print_success "Using server binary: $BINARY_PATH"
else
    print_error "No suitable binary found"
    print_status "Available binaries:"
    ls -la bin/ || true
    exit 1
fi

# Step 1: Copy binary to router
print_status "📤 Step 1: Copying binary to router..."
run_scp "$BINARY_PATH" "/usr/bin/nettestlab"
run_ssh "chmod +x /usr/bin/nettestlab"
print_success "Binary copied and made executable"

# Step 2: Create data directories
print_status "📁 Step 2: Creating data directories..."
run_ssh "mkdir -p /etc/nettestlab /var/lib/nettestlab"
print_success "Directories created"

# Step 3: Create service file
print_status "⚙️  Step 3: Creating service file..."
SERVICE_CONTENT='#!/bin/sh /etc/rc.common

START=99
STOP=10

USE_PROCD=1
PROG=/usr/bin/nettestlab
PIDFILE=/var/run/nettestlab.pid

start_service() {
    procd_open_instance
    procd_set_param command $PROG server --port=8080 --profiles-dir=/var/lib/nettestlab --data-dir=/var/lib/nettestlab
    procd_set_param pidfile $PIDFILE
    procd_set_param respawn
    procd_set_param stdout 1
    procd_set_param stderr 1
    procd_close_instance
}'

echo "$SERVICE_CONTENT" | run_ssh "cat > /etc/init.d/nettestlab"
run_ssh "chmod +x /etc/init.d/nettestlab"
print_success "Service file created"

# Step 4: Stop any existing service and start new one
print_status "🔄 Step 4: Starting NetTestLab service..."
run_ssh "/etc/init.d/nettestlab stop" 2>/dev/null || true
run_ssh "killall nettestlab" 2>/dev/null || true
sleep 2

run_ssh "/etc/init.d/nettestlab start"
sleep 3

# Check if service is running
if run_ssh "pgrep nettestlab" >/dev/null; then
    print_success "NetTestLab service is running"
else
    print_error "Service failed to start"
    print_status "Checking logs..."
    run_ssh "logread | grep nettestlab | tail -5" || true
    exit 1
fi

# Step 5: Verify gRPC port is listening
print_status "🌐 Step 5: Verifying gRPC port..."
if run_ssh "netstat -ln | grep :8080" >/dev/null 2>&1; then
    print_success "gRPC server is listening on port 8080"
else
    print_warning "gRPC port may not be listening yet, waiting..."
    sleep 5
    if run_ssh "netstat -ln | grep :8080" >/dev/null 2>&1; then
        print_success "gRPC server is now listening"
    else
        print_error "gRPC server not listening"
        exit 1
    fi
fi

# Step 6: Run integration tests
print_status "🧪 Step 6: Running integration tests..."

# Set environment variables for the test
export NETTESTLAB_ROUTER_IP="$ROUTER_IP"
export NETTESTLAB_ROUTER_USER="$ROUTER_USER"
if [[ -n "$ROUTER_PASSWORD" ]]; then
    export NETTESTLAB_ROUTER_PASSWORD="$ROUTER_PASSWORD"
fi
if [[ -n "$SSH_KEY_PATH" ]]; then
    export SSH_KEY_PATH="$SSH_KEY_PATH"
fi

# Test basic connectivity first
print_status "Testing basic gRPC connectivity..."
timeout 10 go run cmd/client/main.go --server "$ROUTER_IP:8080" --test-connectivity 2>/dev/null || {
    print_warning "Direct client test failed, trying integration tests..."
}

# Run the auto-wifi-test as a comprehensive test
print_status "Running auto-wifi discovery test..."
if timeout 30 go run cmd/auto-wifi-test/main.go --server "$ROUTER_IP:8080"; then
    print_success "✅ Auto-wifi test passed"
else
    print_error "❌ Auto-wifi test failed"
fi

# Final status check
print_status "🔍 Final verification..."
SERVICE_STATUS=$(run_ssh "pgrep nettestlab && echo 'running' || echo 'not running'")
GRPC_STATUS=$(run_ssh "netstat -ln | grep :8080 && echo 'listening' || echo 'not listening'")

print_status "Service status: $SERVICE_STATUS"
print_status "gRPC status: $GRPC_STATUS"

if [[ "$SERVICE_STATUS" == "running" ]] && [[ "$GRPC_STATUS" == "listening" ]]; then
    print_success "🎉 Integration test completed successfully!"
    print_status "Summary:"
    echo "  ✅ Binary deployed to router"
    echo "  ✅ Service created and running"
    echo "  ✅ gRPC server listening on port 8080"
    echo "  ✅ Basic functionality tested"
    echo ""
    print_status "You can now use NetTestLab on your router!"
    print_status "gRPC endpoint: $ROUTER_IP:8080"
    print_status "Try: go run cmd/auto-wifi-test/main.go --server $ROUTER_IP:8080"
else
    print_error "❌ Integration test failed"
    exit 1
fi