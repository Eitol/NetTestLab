#!/bin/bash

# NetTestLab Router Deployment Script (OpenWRT Compatible)
# Simple deployment for limited OpenWRT environment

set -e

ROUTER_IP="192.168.1.4"
ROUTER_USER="root"
ROUTER_PATH="/tmp/nettestlab"
LOCAL_BINARY="bin/nettestlab-server"

echo "🚀 NetTestLab Router Deployment (OpenWRT Compatible)..."

# Check if router is reachable
echo "🔍 Checking router connectivity..."
if ! ping -c 1 -W 3 $ROUTER_IP > /dev/null 2>&1; then
    echo "❌ Router $ROUTER_IP is not reachable"
    exit 1
fi
echo "✅ Router is reachable"

# Build binary for OpenWRT (Linux ARM)
echo "📦 Building binary for OpenWRT..."
GOOS=linux GOARCH=arm go build -o $LOCAL_BINARY ./cmd/server

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
    echo 'Binary size:' \$(du -h $ROUTER_PATH | cut -f1)
"

# Simple server start (no nohup available)
echo "🎯 Starting server on router..."
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
    # Kill any existing instance
    ps | grep nettestlab | grep -v grep | awk '{print \$1}' | xargs kill 2>/dev/null || true
    
    # Start server in background with simple redirection
    $ROUTER_PATH --port 8080 > /tmp/nettestlab.log 2>&1 &
    SERVER_PID=\$!
    echo \"Server started with PID: \$SERVER_PID\"
    
    # Save PID for later reference
    echo \$SERVER_PID > /tmp/nettestlab.pid
    
    # Give it time to start
    sleep 2
    
    # Check if process is still running
    if kill -0 \$SERVER_PID 2>/dev/null; then
        echo 'Server process is running'
        
        # Check logs
        if [ -f /tmp/nettestlab.log ]; then
            echo 'Server logs:'
            cat /tmp/nettestlab.log
        fi
        
        # Check if port is listening (using netstat or ss)
        if netstat -ln 2>/dev/null | grep :8080 || ss -ln 2>/dev/null | grep :8080; then
            echo 'Server is listening on port 8080'
        else
            echo 'Port 8080 not detected in listening state'
        fi
    else
        echo 'Server process died, checking logs:'
        cat /tmp/nettestlab.log 2>/dev/null || echo 'No logs available'
        exit 1
    fi
"

# Test connectivity from local machine
echo "🔍 Testing connectivity from local machine..."
sleep 3

if nc -z $ROUTER_IP 8080 2>/dev/null; then
    echo "✅ Server is accessible on port 8080"
elif telnet $ROUTER_IP 8080 </dev/null 2>/dev/null; then
    echo "✅ Server is accessible on port 8080 (via telnet)"
else
    echo "⚠️  Server not accessible from local machine"
    echo "   Checking server status..."
    ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "
        echo 'Process status:'
        ps | grep nettestlab | grep -v grep || echo 'No nettestlab process found'
        echo 'Recent logs:'
        tail -20 /tmp/nettestlab.log 2>/dev/null || echo 'No logs available'
        echo 'Network connections:'
        netstat -ln 2>/dev/null | grep -E '(:8080|tcp)' || ss -ln 2>/dev/null | grep 8080 || echo 'No port 8080 found'
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
echo "  - PID File: /tmp/nettestlab.pid"
echo ""
echo "📡 Management Commands:"
echo "  - Check status: ssh $ROUTER_USER@$ROUTER_IP 'ps | grep nettestlab'"
echo "  - View logs: ssh $ROUTER_USER@$ROUTER_IP 'tail -f /tmp/nettestlab.log'"
echo "  - Stop server: ssh $ROUTER_USER@$ROUTER_IP 'kill \$(cat /tmp/nettestlab.pid)'"
echo "  - Restart: ./scripts/deploy-router-simple.sh"