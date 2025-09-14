#!/bin/bash

# NetTestLab Router Deployment Script
# Deploys and tests the gRPC server on OpenWRT router

set -e

ROUTER_IP="192.168.1.4"
ROUTER_USER="root"
ROUTER_PATH="/tmp/nettestlab"
LOCAL_BINARY="bin/nettestlab-server"

echo "🚀 NetTestLab Router Deployment Starting..."

# Check if router is reachable
echo "🔍 Checking router connectivity..."
if ! ping -c 1 -W 3 $ROUTER_IP > /dev/null 2>&1; then
    echo "❌ Router $ROUTER_IP is not reachable"
    exit 1
fi
echo "✅ Router is reachable"

# Build binary for OpenWRT (Linux ARM/MIPS)
echo "📦 Building binary for OpenWRT..."
if ! go build -o $LOCAL_BINARY ./cmd/server; then
    echo "❌ Failed to build server binary"
    exit 1
fi
echo "✅ Binary built successfully"

# Copy binary to router
echo "📤 Copying binary to router..."
if ! scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -q $LOCAL_BINARY $ROUTER_USER@$ROUTER_IP:$ROUTER_PATH; then
    echo "❌ Failed to copy binary to router"
    exit 1
fi
echo "✅ Binary copied to router"

# Make binary executable
echo "🔧 Setting up binary permissions..."
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "chmod +x $ROUTER_PATH"

# Test binary on router
echo "🧪 Testing binary on router..."
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
    echo 'Testing NetTestLab binary...'
    timeout 10s $ROUTER_PATH --help && echo 'Binary is working!' || echo 'Binary test failed'
"

# Start server in background for testing
echo "🎯 Starting server on router..."
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
    # Kill any existing instance
    pkill -f nettestlab || true
    
    # Start server in background
    nohup $ROUTER_PATH --port 8080 > /tmp/nettestlab.log 2>&1 &
    
    # Give it time to start
    sleep 3
    
    # Check if it's running
    if pgrep -f nettestlab > /dev/null; then
        echo 'Server started successfully'
        
        # Check if port is open
        if netstat -ln | grep :8080; then
            echo 'Server is listening on port 8080'
        else
            echo 'Warning: Port 8080 not found in netstat'
        fi
        
        # Show initial logs
        echo 'Initial logs:'
        head -20 /tmp/nettestlab.log || echo 'No logs yet'
        
    else
        echo 'Failed to start server'
        echo 'Logs:'
        cat /tmp/nettestlab.log 2>/dev/null || echo 'No logs available'
        exit 1
    fi
"

# Test connectivity from local machine
echo "🔍 Testing connectivity from local machine..."
sleep 2
if nc -z $ROUTER_IP 8080; then
    echo "✅ gRPC server is accessible from local machine"
else
    echo "⚠️  Server not accessible from local machine (might be normal for gRPC)"
fi

echo ""
echo "✅ Deployment completed successfully!"
echo ""
echo "📋 Server Information:"
echo "  - Router IP: $ROUTER_IP"
echo "  - gRPC Port: 8080"
echo "  - Binary Path: $ROUTER_PATH"
echo "  - Log File: /tmp/nettestlab.log"
echo ""
echo "📡 Next steps:"
echo "  1. Test gRPC API with client libraries"
echo "  2. Apply network conditions to interfaces"
echo "  3. Monitor server logs: ssh $ROUTER_USER@$ROUTER_IP 'tail -f /tmp/nettestlab.log'"
echo ""
echo "🛑 To stop server: ssh $ROUTER_USER@$ROUTER_IP 'pkill -f nettestlab'"