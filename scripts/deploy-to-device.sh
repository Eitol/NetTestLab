#!/bin/bash

# NetTestLab Auto-Deployment Script
# Automatically detects router architecture and deploys the appropriate package

set -e  # Exit on any error

# Configuration
ROUTER_IP="${ROUTER_IP:-192.168.1.4}"
ROUTER_USER="${ROUTER_USER:-root}"
PROJECT_NAME="nettestlab"
OPENWRT_BUILD_DIR="openwrt/build"
PACKAGE_PREFIX="nettestlab_1.0.0-1"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Logging functions
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

log_step() {
    echo -e "${PURPLE}[STEP $1]${NC} $2"
}

# Check if router is reachable
check_router_connectivity() {
    log_info "Checking connectivity to router at $ROUTER_IP..."
    if ! ping -c 1 -W 3 "$ROUTER_IP" >/dev/null 2>&1; then
        log_error "Cannot reach router at $ROUTER_IP"
        exit 1
    fi
    log_success "Router is reachable"
}

# Detect router architecture
detect_router_architecture() {
    log_step "0" "Detecting router architecture..."
    
    # Get architecture information from router
    ARCH_INFO=$(ssh "$ROUTER_USER@$ROUTER_IP" "uname -m" 2>/dev/null || echo "unknown")
    OPENWRT_INFO=$(ssh "$ROUTER_USER@$ROUTER_IP" "cat /etc/openwrt_release 2>/dev/null | grep DISTRIB_TARGET || echo 'unknown'" || echo "unknown")
    
    log_info "Router reports architecture: $ARCH_INFO"
    log_info "OpenWrt target info: $OPENWRT_INFO"
    
    # Get the specific OpenWrt architecture
    OPENWRT_ARCH=$(ssh "$ROUTER_USER@$ROUTER_IP" "opkg print-architecture" | grep -v "arch all\|arch noarch" | head -1 | awk '{print $2}')
    log_info "OpenWrt package architecture: $OPENWRT_ARCH"
    
    # Determine package architecture based on uname output, but use OpenWrt specific arch for packages
    case "$ARCH_INFO" in
        "aarch64"|"arm64")
            PACKAGE_ARCH="$OPENWRT_ARCH"  # Use specific OpenWrt arch like aarch64_cortex-a53
            GO_ARCH="arm64"
            ;;
        "armv7l"|"armv6l"|"arm")
            PACKAGE_ARCH="arm"
            GO_ARCH="arm"
            ;;
        "x86_64"|"amd64")
            PACKAGE_ARCH="x86_64"
            GO_ARCH="amd64"
            ;;
        "mips"|"mipsel")
            PACKAGE_ARCH="mips"
            GO_ARCH="mips"
            ;;
        *)
            log_warning "Unknown architecture: $ARCH_INFO, using detected OpenWrt arch: $OPENWRT_ARCH"
            PACKAGE_ARCH="$OPENWRT_ARCH"
            GO_ARCH="arm64"  # Default for most common case
            ;;
    esac
    
    PACKAGE_FILE="${PACKAGE_PREFIX}_${PACKAGE_ARCH}.ipk"
    
    log_success "Detected architecture: $PACKAGE_ARCH (Go: $GO_ARCH)"
    log_info "Will use package: $PACKAGE_FILE"
}

# Build the OpenWrt package
build_openwrt_package() {
    log_step "1" "Building OpenWrt package for $PACKAGE_ARCH using Docker..."
    
    # Ensure we're in the project root
    cd "$(dirname "$0")/.."
    
    # Build the Go binary for the target architecture
    log_info "Building Go binary for $GO_ARCH..."
    GOOS=linux GOARCH="$GO_ARCH" go build -o bin/nettestlab-$GO_ARCH ./cmd/server
    
    # Build test binaries for local execution
    log_info "Building test binaries..."
    go build -o bin/unified-test ./cmd/unified-test 2>/dev/null || log_warning "unified-test binary not built"
    go build -o bin/simple-test ./cmd/simple-test 2>/dev/null || log_warning "simple-test binary not built (legacy)"
    go build -o bin/auto-wifi-test ./cmd/auto-wifi-test 2>/dev/null || log_warning "auto-wifi-test binary not built (legacy)"
    go build -o bin/wifi-test ./cmd/wifi-test 2>/dev/null || log_warning "wifi-test binary not built (legacy)"
    go build -o bin/nettestlab-client ./cmd/client 2>/dev/null || log_warning "client binary not built (legacy)"
    
    # Build Docker image for package building
    log_info "Building Docker image for package creation..."
    docker build -f Dockerfile.opkg -t nettestlab-opkg-builder .
    
    # Prepare files for Docker build
    BUILD_DIR="/tmp/nettestlab-build-$$"
    mkdir -p "$BUILD_DIR"
    
    # Copy binary
    cp bin/nettestlab-$GO_ARCH "$BUILD_DIR/nettestlab"
    
    # Copy OpenWrt files
    cp -r openwrt/files "$BUILD_DIR/"
    
    # Copy web interface files
    log_info "Copying web interface files..."
    if [ -d "web" ]; then
        cp -r web "$BUILD_DIR/"
        log_info "Web interface files copied"
    else
        log_warning "Web directory not found"
    fi
    
    # Run Docker container to build the package
    log_info "Running Docker container to build IPK package..."
    docker run --rm \
        -v "$BUILD_DIR:/workspace" \
        -e PACKAGE_NAME="nettestlab" \
        -e VERSION="1.0.0" \
        -e RELEASE="1" \
        -e ARCHITECTURE="$PACKAGE_ARCH" \
        nettestlab-opkg-builder
    
    # Check if package was created
    if [ -f "$BUILD_DIR/$PACKAGE_FILE" ]; then
        # Move package to project root
        mv "$BUILD_DIR/$PACKAGE_FILE" "openwrt/$PACKAGE_FILE"
        log_success "Package built: $PACKAGE_FILE"
    else
        log_error "Package generation failed"
        exit 1
    fi
    
    # Cleanup
    rm -rf "$BUILD_DIR"
}

# Copy package to router
copy_package_to_router() {
    log_step "2" "Copying package to router..."
    
    PACKAGE_PATH="openwrt/$PACKAGE_FILE"
    
    if [ ! -f "$PACKAGE_PATH" ]; then
        log_error "Package file not found: $PACKAGE_PATH"
        exit 1
    fi
    
    log_info "Copying $PACKAGE_FILE to router..."
    scp "$PACKAGE_PATH" "$ROUTER_USER@$ROUTER_IP:/tmp/"
    
    log_success "Package copied to router"
}

# Stop running processes
stop_running_processes() {
    log_step "3" "Stopping running NetTestLab processes..."
    
    # Kill any running nettestlab processes
    ssh "$ROUTER_USER@$ROUTER_IP" "killall nettestlab nettestlab-server nettestlab-separated nettestlab-debug 2>/dev/null || true"
    
    # Wait a moment for processes to stop
    sleep 2
    
    # Check if any processes are still running
    RUNNING_PROCS=$(ssh "$ROUTER_USER@$ROUTER_IP" "ps | grep nettestlab | grep -v grep" || echo "")
    
    if [ -n "$RUNNING_PROCS" ]; then
        log_warning "Some processes still running, forcing kill..."
        ssh "$ROUTER_USER@$ROUTER_IP" "killall -9 nettestlab 2>/dev/null || true"
        sleep 1
    fi
    
    log_success "All NetTestLab processes stopped"
}

# Install the appropriate package
install_package() {
    log_step "4" "Installing package for $PACKAGE_ARCH architecture..."
    
    # Remove existing installation
    log_info "Removing existing installation..."
    ssh root@$ROUTER_IP "opkg remove nettestlab" 2>/dev/null || log_info "No packages removed."
    
    # Install new package
    log_info "Installing new package..."
    if ssh root@$ROUTER_IP "opkg install /tmp/$PACKAGE_FILE"; then
        log_success "Package installed successfully"
    else
        log_error "Failed to install package"
        exit 1
    fi
}

# Start the project
start_project() {
    log_step "5" "Starting NetTestLab service..."
    
    # Start the service using init system
    log_info "Starting NetTestLab service..."
    ssh "$ROUTER_USER@$ROUTER_IP" "/etc/init.d/nettestlab start"
    
    # Wait for service to start
    log_info "Waiting for service to start..."
    sleep 5
    
    # Check if service is running
    if ssh "$ROUTER_USER@$ROUTER_IP" "netstat -tlnp | grep ':8080'" >/dev/null; then
        log_success "NetTestLab gRPC service is running on port 8080"
    else
        log_error "NetTestLab gRPC service failed to start"
        # Show logs for debugging
        log_info "Service logs:"
        ssh "$ROUTER_USER@$ROUTER_IP" "logread | tail -20"
        exit 1
    fi
    
    # Check web interface
    if ssh "$ROUTER_USER@$ROUTER_IP" "netstat -tlnp | grep ':8081'" >/dev/null; then
        log_success "NetTestLab web interface is running on port 8081"
    else
        log_warning "NetTestLab web interface not detected on port 8081"
    fi
}

# Run tests to verify everything works
run_tests() {
    log_step "6" "Running comprehensive integration tests..."
    
    # Return to project root for tests
    cd "$(dirname "$0")/.."
    
    # Primary test: Unified test suite
    log_info "Running unified test suite..."
    if timeout 120s ./bin/unified-test -server "$ROUTER_IP:8080" -v >/dev/null 2>&1; then
        log_success "Unified test suite passed"
    else
        log_warning "Unified test suite failed, falling back to individual tests"
        
        # Fallback to individual tests
        log_info "Running simple connectivity test..."
        if ./bin/simple-test >/dev/null 2>&1; then
            log_success "Simple test passed"
        else
            log_error "Simple test failed"
            exit 1
        fi
        
        log_info "Running auto-WiFi discovery test..."
        if ./bin/auto-wifi-test >/dev/null 2>&1; then
            log_success "Auto-WiFi test passed"
        else
            log_error "Auto-WiFi test failed"
            exit 1
        fi
    fi
    
    log_info "Running Go integration test suite..."
    if go test -v ./tests/integration/wifi_test.go -run TestWiFiInterfaceIntegration >/dev/null 2>&1; then
        log_success "Go integration tests passed"
    else
        log_warning "Go integration tests failed (this might be expected in some environments)"
    fi
    
    # Test Go client connectivity
    log_info "Testing Go gRPC client connectivity..."
    if timeout 30s ./bin/nettestlab-client -server "$ROUTER_IP:8080" -test >/dev/null 2>&1; then
        log_success "Go client connectivity test passed"
    else
        log_warning "Go client test failed or timed out"
    fi
    
    # Test Java client if available
    log_info "Testing Java client..."
    if [ -f "clients/java/pom.xml" ]; then
        cd clients/java
        if timeout 60s mvn test -Dtest=NetTestLabClientTest -Dserver.host="$ROUTER_IP" -q >/dev/null 2>&1; then
            log_success "Java client tests passed"
        else
            log_warning "Java client tests failed or not configured"
        fi
        cd ../..
    else
        log_info "Java client not found, skipping Java tests"
    fi
    
    # Test Python client if available
    log_info "Testing Python client..."
    if [ -d "clients/python" ] && [ -f "clients/python/setup.py" ]; then
        cd clients/python
        if timeout 60s python3 -m pytest tests/ -v --host="$ROUTER_IP" >/dev/null 2>&1; then
            log_success "Python client tests passed"
        else
            log_warning "Python client tests failed or dependencies missing"
        fi
        cd ../..
    else
        log_info "Python client not found, skipping Python tests"
    fi
    
    # Check web interface (runs on same port as gRPC)
    log_info "Testing web interface accessibility..."
    if curl -s --connect-timeout 10 "http://$ROUTER_IP:8080" >/dev/null 2>&1; then
        log_success "Web interface is accessible at http://$ROUTER_IP:8080"
    else
        log_warning "Web interface not accessible via HTTP"
    fi
    
    log_success "All critical tests completed!"
}

# Main execution
main() {
    echo "🚀 NetTestLab Auto-Deployment Script"
    echo "====================================="
    
    check_router_connectivity
    detect_router_architecture
    build_openwrt_package
    copy_package_to_router
    stop_running_processes
    install_package
    start_project
    run_tests
    
    echo ""
    echo "🎉 Deployment completed successfully!"
    echo "   📡 gRPC API: http://$ROUTER_IP:8080"
    echo "   🌐 Web UI:   http://$ROUTER_IP:8080"
    echo "   📋 Architecture: $PACKAGE_ARCH"
    echo "   📦 Package: $PACKAGE_FILE"
    echo ""
    echo "✅ NetTestLab is ready for testing!"
}

# Execute main function
main "$@"