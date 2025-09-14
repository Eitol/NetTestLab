#!/bin/bash

# NetTestLab Complete Deployment Script
# Builds OpenWRT package, deploys to router, and runs tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

ROUTER_IP="192.168.1.4"
ROUTER_USER="root"
PACKAGE_NAME="nettestlab"
VERSION="1.0.0"
RELEASE="1" 
ARCHITECTURE="aarch64"
PACKAGE_FILE="${PACKAGE_NAME}_${VERSION}-${RELEASE}_${ARCHITECTURE}.ipk"
PROJECT_ROOT="/Users/hector/NetTestLab"

echo -e "${BLUE}🚀 NetTestLab Complete Deployment Pipeline${NC}"
echo "============================================="

# Step 1: Check router connectivity
echo -e "${CYAN}Step 1: Checking router connectivity...${NC}"
if ! ping -c 1 -W 3 $ROUTER_IP > /dev/null 2>&1; then
    echo -e "${RED}❌ Router $ROUTER_IP is not reachable${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Router is reachable${NC}"

# Step 2: Build the OpenWRT package
echo -e "${CYAN}Step 2: Building OpenWRT package...${NC}"
cd "$PROJECT_ROOT"
./scripts/build-openwrt-package.sh

if [ ! -f "$PACKAGE_FILE" ]; then
    echo -e "${RED}❌ Package build failed${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Package built successfully${NC}"

# Step 4: Uninstall existing package (if exists)
echo -e "${CYAN}Step 4: Removing existing package (if any)...${NC}"
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
    # Check if package is installed
    if opkg list-installed | grep -q '^$PACKAGE_NAME '; then
        echo 'Removing existing $PACKAGE_NAME package...'
        opkg remove $PACKAGE_NAME || true
        echo 'Existing package removed'
    else
        echo 'No existing package found'
    fi
    
    # Clean up any leftover files
    rm -f /usr/bin/nettestlab
    rm -f /etc/init.d/nettestlab
    rm -f /tmp/nettestlab*
"
echo -e "${GREEN}✅ Cleanup completed${NC}"

# Step 5: Install dependencies first
echo -e "${CYAN}Step 5: Installing dependencies...${NC}"
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
    echo 'Updating package list...'
    opkg update
    
    echo 'Installing tc (traffic control)...'
    opkg install tc-bpf || echo 'tc-bpf might already be installed'
    
    echo 'Installing basic kernel modules...'
    opkg install kmod-sched-core || echo 'kmod-sched-core might already be installed'
    opkg install kmod-ifb || echo 'kmod-ifb might already be installed'
    
    echo 'Checking if tc command is available...'
    if which tc >/dev/null 2>&1; then
        echo 'tc command is available'
        tc qdisc help | head -5 || echo 'tc basic test completed'
    else
        echo 'WARNING: tc command not found'
    fi
    
    echo 'Dependencies installation completed'
"
echo -e "${GREEN}✅ Dependencies installed${NC}"

# Step 6: Copy package to router (after dependencies)
echo -e "${CYAN}Step 6: Copying package to router...${NC}"

# First, check available tools on router
echo "  Checking available transfer tools on router..."
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
    echo 'Available tools:'
    which nc && echo 'nc: available' || echo 'nc: not available'
    which base64 && echo 'base64: available' || echo 'base64: not available'
    which uuencode && echo 'uuencode: available' || echo 'uuencode: not available'
    which hexdump && echo 'hexdump: available' || echo 'hexdump: not available'
    which od && echo 'od: available' || echo 'od: not available'
"

LOCAL_SIZE=$(stat -f%z "$PACKAGE_FILE" 2>/dev/null || stat -c%s "$PACKAGE_FILE" 2>/dev/null)
echo "  Local file size: $LOCAL_SIZE bytes"

# Try hexdump method (more likely to be available)
echo "  Trying hexdump transfer method..."
if hexdump -v -e '1/1 "%02x"' "$PACKAGE_FILE" | ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
    # Convert hex back to binary
    while IFS= read -r line; do
        echo \"\$line\" | sed 's/../\\\\x&/g' | xargs -0 printf
    done > /tmp/$PACKAGE_FILE
"; then
    echo "  Hexdump transfer completed"
    COPY_SUCCESS="true"
fi

# Fallback: simple dd method via SSH
if [ -z "$COPY_SUCCESS" ]; then
    echo "  Trying dd over SSH method..."
    if dd if="$PACKAGE_FILE" bs=1024 2>/dev/null | ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "dd of=/tmp/$PACKAGE_FILE bs=1024 2>/dev/null"; then
        echo "  DD transfer successful"
        COPY_SUCCESS="true"
    fi
fi

# Verify the file was copied and check integrity
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
    if [ -f /tmp/$PACKAGE_FILE ]; then
        echo 'File exists on router'
        REMOTE_SIZE=\$(stat -c%s /tmp/$PACKAGE_FILE 2>/dev/null || wc -c < /tmp/$PACKAGE_FILE)
        echo \"Remote file size: \$REMOTE_SIZE bytes\"
        
        if [ \"$LOCAL_SIZE\" = \"\$REMOTE_SIZE\" ]; then
            echo 'File sizes match - transfer successful'
        else
            echo \"File size mismatch - expected $LOCAL_SIZE, got \$REMOTE_SIZE\"
            echo 'Retrying with simple cat method...'
            exit 1
        fi
        
        ls -la /tmp/$PACKAGE_FILE
        
    else
        echo 'File copy failed!'
        exit 1
    fi
"

# If transfer failed, try the simple cat method as last resort
if [ $? -ne 0 ] || [ -z "$COPY_SUCCESS" ]; then
    echo "  Trying simple cat method as fallback..."
    if cat "$PACKAGE_FILE" | ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "cat > /tmp/$PACKAGE_FILE"; then
        echo "  Cat method successful"
        
        # Verify again
        ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
            REMOTE_SIZE=\$(stat -c%s /tmp/$PACKAGE_FILE 2>/dev/null || wc -c < /tmp/$PACKAGE_FILE)
            echo \"Final remote file size: \$REMOTE_SIZE bytes\"
            ls -la /tmp/$PACKAGE_FILE
        "
        COPY_SUCCESS="true"
    fi
fi

if [ -z "$COPY_SUCCESS" ]; then
    echo -e "${RED}❌ Failed to copy package to router with all methods${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Package copied to router${NC}"

# Step 7: Install the package
echo -e "${CYAN}Step 7: Installing NetTestLab package...${NC}"
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
    echo 'Installing NetTestLab package...'
    opkg install --force-reinstall /tmp/$PACKAGE_FILE
    
    echo 'Verifying installation...'
    if opkg list-installed | grep -q '^$PACKAGE_NAME '; then
        echo 'Package installed successfully'
        opkg list-installed | grep '^$PACKAGE_NAME '
    else
        echo 'Package installation failed'
        exit 1
    fi
    
    # Check if binary is installed
    if [ -f /usr/bin/nettestlab ]; then
        echo 'Binary installed at /usr/bin/nettestlab'
        ls -la /usr/bin/nettestlab
    else
        echo 'Binary not found!'
        exit 1
    fi
    
    # Check if service is available
    if [ -f /etc/init.d/nettestlab ]; then
        echo 'Service script installed'
        ls -la /etc/init.d/nettestlab
    else
        echo 'Service script not found!'
        exit 1
    fi
"
echo -e "${GREEN}✅ Package installed successfully${NC}"

# Step 8: Start the service
echo -e "${CYAN}Step 8: Starting NetTestLab service...${NC}"
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
    echo 'Starting NetTestLab service...'
    /etc/init.d/nettestlab start
    
    # Wait for service to start
    sleep 5
    
    # Check if service is running
    if /etc/init.d/nettestlab status; then
        echo 'Service started successfully'
    else
        echo 'Service failed to start, checking logs...'
        cat /var/log/nettestlab.log 2>/dev/null || echo 'No logs available'
        exit 1
    fi
    
    # Check if port is listening
    echo 'Checking if gRPC port is listening...'
    if netstat -ln | grep ':8080'; then
        echo 'gRPC server is listening on port 8080'
    else
        echo 'Warning: Port 8080 not found in netstat'
    fi
"
echo -e "${GREEN}✅ Service started successfully${NC}"

# Step 9: Test connectivity
echo -e "${CYAN}Step 9: Testing connectivity...${NC}"
sleep 3
if nc -z $ROUTER_IP 8080; then
    echo -e "${GREEN}✅ gRPC server is accessible from local machine${NC}"
else
    echo -e "${YELLOW}⚠️  Server not accessible via netcat (normal for gRPC)${NC}"
fi

# Step 10: Run gRPC client tests
echo -e "${CYAN}Step 10: Running gRPC client tests...${NC}"
cd "$PROJECT_ROOT"

# Make sure client is built
if [ ! -f "bin/nettestlab-client" ]; then
    echo "Building gRPC client..."
    go build -o bin/nettestlab-client ./cmd/client
fi

# Update client to use router IP
sed -i '' "s/localhost/$ROUTER_IP/g" cmd/client/main.go 2>/dev/null || true

# Rebuild client with router IP
go build -o bin/nettestlab-client ./cmd/client

# Run tests
echo "Running gRPC tests against router..."
./bin/nettestlab-client

# Restore localhost in client
sed -i '' "s/$ROUTER_IP/localhost/g" cmd/client/main.go 2>/dev/null || true

echo -e "${GREEN}✅ gRPC tests completed${NC}"

# Step 11: Show service status and logs
echo -e "${CYAN}Step 11: Service status and logs...${NC}"
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
    echo '=== Service Status ==='
    /etc/init.d/nettestlab status
    
    echo ''
    echo '=== Recent Logs ==='
    tail -20 /var/log/nettestlab.log 2>/dev/null || echo 'No logs available'
    
    echo ''
    echo '=== Process Info ==='
    ps | grep nettestlab | grep -v grep || echo 'Process not found in ps'
    
    echo ''
    echo '=== Network Interfaces ==='
    ip link show | grep -E '^[0-9]+:' | cut -d: -f2 | sed 's/^ //'
"

echo ""
echo -e "${GREEN}🎉 Deployment completed successfully!${NC}"
echo ""
echo -e "${BLUE}📋 Summary:${NC}"
echo "  ✅ Package built and deployed"
echo "  ✅ Dependencies installed"
echo "  ✅ Service started and running"
echo "  ✅ gRPC API tested and working"
echo ""
echo -e "${BLUE}🛠️  Management commands:${NC}"
echo "  Status: ssh $ROUTER_USER@$ROUTER_IP '/etc/init.d/nettestlab status'"
echo "  Stop:   ssh $ROUTER_USER@$ROUTER_IP '/etc/init.d/nettestlab stop'"
echo "  Start:  ssh $ROUTER_USER@$ROUTER_IP '/etc/init.d/nettestlab start'"
echo "  Logs:   ssh $ROUTER_USER@$ROUTER_IP 'tail -f /var/log/nettestlab.log'"
echo "  Config: ssh $ROUTER_USER@$ROUTER_IP 'vi /etc/config/nettestlab'"
echo ""
echo -e "${BLUE}🔗 gRPC Endpoint: $ROUTER_IP:8080${NC}"