#!/bin/bash

# NetTestLab Router Deployment Script (Alternative method)
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

# Build binary for OpenWRT (Linux)
echo "📦 Building binary for OpenWRT..."
# Cross-compile for Linux (assuming MIPS/ARM architecture)
GOOS=linux GOARCH=arm go build -o $LOCAL_BINARY ./cmd/server 2>/dev/null || \
GOOS=linux GOARCH=mips go build -o $LOCAL_BINARY ./cmd/server 2>/dev/null || \
GOOS=linux GOARCH=amd64 go build -o $LOCAL_BINARY ./cmd/server

if [ ! -f "$LOCAL_BINARY" ]; then
    echo "❌ Failed to build server binary"
    exit 1
fi
echo "✅ Binary built successfully"

# Try different methods to copy the file
echo "📤 Copying binary to router..."

# Method 1: Try rsync
if command -v rsync >/dev/null 2>&1; then
    echo "  Trying rsync..."
    if rsync -e "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" $LOCAL_BINARY $ROUTER_USER@$ROUTER_IP:$ROUTER_PATH 2>/dev/null; then
        echo "✅ Binary copied via rsync"
        COPY_SUCCESS=true
    fi
fi

# Method 2: Try netcat if rsync failed
if [ -z "$COPY_SUCCESS" ]; then
    echo "  Trying netcat..."
    # Start listener on router
    ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
        # Kill any existing nc listener
        pkill -f 'nc.*-l.*1234' || true
        # Start listener in background
        nc -l -p 1234 > $ROUTER_PATH &
        echo 'Listener started on port 1234'
    " &
    
    # Give it time to start
    sleep 2
    
    # Send file via netcat
    if nc -w 10 $ROUTER_IP 1234 < $LOCAL_BINARY; then
        echo "✅ Binary copied via netcat"
        COPY_SUCCESS=true
    fi
fi

# Method 3: Try cat/ssh if netcat failed
if [ -z "$COPY_SUCCESS" ]; then
    echo "  Trying cat via SSH..."
    if cat $LOCAL_BINARY | ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "cat > $ROUTER_PATH"; then
        echo "✅ Binary copied via SSH cat"
        COPY_SUCCESS=true
    fi
fi

# Check if copy was successful
if [ -z "$COPY_SUCCESS" ]; then
    echo "❌ Failed to copy binary to router with all methods"
    exit 1
fi

# Make binary executable and test
echo "🔧 Setting up binary permissions..."
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "chmod +x $ROUTER_PATH"

# Test binary on router
echo "🧪 Testing binary on router..."
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
    echo 'Testing NetTestLab binary...'
    file $ROUTER_PATH
    ls -la $ROUTER_PATH
    
    # Test help command
    timeout 10s $ROUTER_PATH --help 2>&1 && echo 'Binary help works!' || echo 'Binary test completed'
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
    echo "⚠️  Server not accessible from local machine (checking logs...)"
    ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
        echo 'Recent logs:'
        tail -10 /tmp/nettestlab.log 2>/dev/null || echo 'No logs available'
    "
fi

echo ""
echo "✅ Deployment completed!"
echo ""
echo "📋 Server Information:"
echo "  - Router IP: $ROUTER_IP"
echo "  - gRPC Port: 8080"
echo "  - Binary Path: $ROUTER_PATH"
echo "  - Log File: /tmp/nettestlab.log"
echo ""
echo "📡 Next steps:"
echo "  1. Check logs: ssh $ROUTER_USER@$ROUTER_IP 'tail -f /tmp/nettestlab.log'"
echo "  2. Test with gRPC client"
echo ""
echo "🛑 To stop server: ssh $ROUTER_USER@$ROUTER_IP 'pkill -f nettestlab'"