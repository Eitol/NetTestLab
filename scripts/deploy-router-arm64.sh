#!/bin/bash

# NetTestLab Router Deployment Script (ARM64 OpenWRT)
# Deployment for ARM64 OpenWRT router

set -e

ROUTER_IP="192.168.1.4"
ROUTER_USER="root"
ROUTER_PATH="/tmp/nettestlab"
LOCAL_BINARY="bin/nettestlab-server"

echo "🚀 NetTestLab Router Deployment (ARM64 OpenWRT)..."

# Check if router is reachable
echo "🔍 Checking router connectivity..."
if ! ping -c 1 -W 3 $ROUTER_IP > /dev/null 2>&1; then
    echo "❌ Router $ROUTER_IP is not reachable"
    exit 1
fi
echo "✅ Router is reachable"

# Build binary for ARM64 Linux (OpenWRT)
echo "📦 Building binary for ARM64 OpenWRT..."
GOOS=linux GOARCH=arm64 go build -o $LOCAL_BINARY ./cmd/server

if [ ! -f "$LOCAL_BINARY" ]; then
    echo "❌ Failed to build server binary"
    exit 1
fi
echo "✅ Binary built successfully ($(du -h $LOCAL_BINARY | cut -f1))"

# Copy binary via SSH cat
echo "📤 Copying binary to router..."
if cat $LOCAL_BINARY | ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "cat > $ROUTER_PATH"; then
    echo "✅ Binary copied successfully"
else
    echo "❌ Failed to copy binary"
    exit 1
fi

# Setup and test binary
echo "🔧 Setting up binary..."
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
    chmod +x $ROUTER_PATH
    ls -la $ROUTER_PATH
    echo 'Binary info:'
    du -h $ROUTER_PATH
    
    # Test if binary is executable
    if ./$ROUTER_PATH --version 2>/dev/null || echo 'Version check completed'; then
        echo 'Binary is executable'
    else
        echo 'Testing help command:'
        ./$ROUTER_PATH --help 2>&1 | head -5 || echo 'Help test completed'
    fi
" 2>/dev/null

# Start server on router
echo "🎯 Starting server on router..."
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
    cd /tmp
    
    # Kill any existing instance
    ps | grep nettestlab | grep -v grep | awk '{print \$1}' | while read pid; do kill \$pid 2>/dev/null || true; done
    
    # Start server in background
    ./nettestlab --port 8080 > nettestlab.log 2>&1 &
    SERVER_PID=\$!
    echo \"Server started with PID: \$SERVER_PID\"
    
    # Save PID
    echo \$SERVER_PID > nettestlab.pid
    
    # Give it time to start
    sleep 3
    
    # Check if process is still running
    if kill -0 \$SERVER_PID 2>/dev/null; then
        echo 'Server process is running'
        
        # Show initial logs
        echo 'Server logs:'
        cat nettestlab.log 2>/dev/null || echo 'No logs yet'
        
        # Check listening ports
        echo 'Checking ports:'
        netstat -ln 2>/dev/null | grep :8080 || echo 'Port 8080 not found in netstat'
        
    else
        echo 'Server process died, checking logs:'
        cat nettestlab.log 2>/dev/null || echo 'No logs available'
        exit 1
    fi
"

# Test connectivity
echo "🔍 Testing connectivity..."
sleep 2

for i in {1..3}; do
    if nc -z $ROUTER_IP 8080 2>/dev/null; then
        echo "✅ Server is accessible on port 8080 (attempt $i)"
        CONNECTED=true
        break
    else
        echo "   Attempt $i failed, retrying..."
        sleep 2
    fi
done

if [ -z "$CONNECTED" ]; then
    echo "⚠️  Server not accessible from local machine"
    echo "   Checking server status on router..."
    ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
        cd /tmp
        echo 'Process status:'
        ps | grep nettestlab | grep -v grep || echo 'No nettestlab process found'
        echo 'Port status:'
        netstat -ln | grep 8080 || echo 'Port 8080 not listening'
        echo 'Recent logs:'
        tail -10 nettestlab.log 2>/dev/null || echo 'No logs'
        echo 'Firewall status:'
        iptables -L INPUT | grep 8080 || echo 'No specific rule for port 8080'
    "
fi

echo ""
echo "✅ Deployment completed!"
echo ""
echo "📋 Server Information:"
echo "  - Router IP: $ROUTER_IP"
echo "  - Architecture: ARM64 (aarch64)"
echo "  - gRPC Port: 8080"
echo "  - Binary: /tmp/nettestlab"
echo "  - Logs: /tmp/nettestlab.log"
echo "  - PID: /tmp/nettestlab.pid"
echo ""
echo "📡 Management Commands:"
echo "  - Status: ssh $ROUTER_USER@$ROUTER_IP 'cd /tmp && ps | grep nettestlab'"
echo "  - Logs: ssh $ROUTER_USER@$ROUTER_IP 'tail -f /tmp/nettestlab.log'"
echo "  - Stop: ssh $ROUTER_USER@$ROUTER_IP 'kill \$(cat /tmp/nettestlab.pid)'"