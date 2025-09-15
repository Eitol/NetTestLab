#!/bin/bash#!/bin/bash



# NetTestLab Deployment Script# NetTestLab Router Deployment Script

# Deploys NetTestLab to router at 192.168.1.4# Deploys and tests the gRPC server on OpenWRT router



ROUTER_IP="192.168.1.4"set -e

ROUTER_USER="root"

REMOTE_DIR="/opt/nettestlab"ROUTER_IP="192.168.1.4"

LOCAL_DIR="/Users/hector/NetTestLab"ROUTER_USER="root"

ROUTER_PATH="/tmp/nettestlab"

echo "🚀 Starting NetTestLab deployment to router ${ROUTER_IP}"LOCAL_BINARY="bin/nettestlab-server"



# Create remote directoryecho "🚀 NetTestLab Router Deployment Starting..."

echo "📁 Creating remote directory..."

ssh ${ROUTER_USER}@${ROUTER_IP} "mkdir -p ${REMOTE_DIR} && mkdir -p ${REMOTE_DIR}/web"# Check if router is reachable

echo "🔍 Checking router connectivity..."

# Copy binaryif ! ping -c 1 -W 3 $ROUTER_IP > /dev/null 2>&1; then

echo "📦 Copying NetTestLab binary..."    echo "❌ Router $ROUTER_IP is not reachable"

scp ${LOCAL_DIR}/bin/nettestlab-linux ${ROUTER_USER}@${ROUTER_IP}:${REMOTE_DIR}/nettestlab    exit 1

ssh ${ROUTER_USER}@${ROUTER_IP} "chmod +x ${REMOTE_DIR}/nettestlab"fi

echo "✅ Router is reachable"

# Copy web interface

echo "🌐 Copying web interface..."# Build binary for OpenWRT (Linux ARM/MIPS)

scp -r ${LOCAL_DIR}/web/* ${ROUTER_USER}@${ROUTER_IP}:${REMOTE_DIR}/web/echo "📦 Building binary for OpenWRT..."

if ! go build -o $LOCAL_BINARY ./cmd/server; then

# Create systemd service file    echo "❌ Failed to build server binary"

echo "⚙️  Creating systemd service..."    exit 1

cat > /tmp/nettestlab.service << EOFfi

[Unit]echo "✅ Binary built successfully"

Description=NetTestLab Network Testing Service

After=network.target# Copy binary to router

Wants=network.targetecho "📤 Copying binary to router..."

if ! scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -q $LOCAL_BINARY $ROUTER_USER@$ROUTER_IP:$ROUTER_PATH; then

[Service]    echo "❌ Failed to copy binary to router"

Type=simple    exit 1

ExecStart=${REMOTE_DIR}/nettestlab -host 0.0.0.0 -port 8080 -web-dir ${REMOTE_DIR}/webfi

Restart=alwaysecho "✅ Binary copied to router"

RestartSec=5

User=root# Make binary executable

WorkingDirectory=${REMOTE_DIR}echo "🔧 Setting up binary permissions..."

ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "chmod +x $ROUTER_PATH"

[Install]

WantedBy=multi-user.target# Test binary on router

EOFecho "🧪 Testing binary on router..."

ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "

scp /tmp/nettestlab.service ${ROUTER_USER}@${ROUTER_IP}:/etc/systemd/system/    echo 'Testing NetTestLab binary...'

    timeout 10s $ROUTER_PATH --help && echo 'Binary is working!' || echo 'Binary test failed'

# Enable and start service"

echo "🔄 Enabling and starting NetTestLab service..."

ssh ${ROUTER_USER}@${ROUTER_IP} "# Start server in background for testing

    systemctl daemon-reload &&echo "🎯 Starting server on router..."

    systemctl enable nettestlab &&ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $ROUTER_USER@$ROUTER_IP "

    systemctl restart nettestlab &&    # Kill any existing instance

    sleep 2 &&    pkill -f nettestlab || true

    systemctl status nettestlab    

"    # Start server in background

    nohup $ROUTER_PATH --port 8080 > /tmp/nettestlab.log 2>&1 &

echo "✅ Deployment completed!"    

echo "🌐 Web interface should be available at: http://${ROUTER_IP}:8080/web"    # Give it time to start

echo "🔧 gRPC API available at: ${ROUTER_IP}:8080"    sleep 3

    

# Test connectivity    # Check if it's running

echo "🧪 Testing connectivity..."    if pgrep -f nettestlab > /dev/null; then

curl -f http://${ROUTER_IP}:8080/api/health && echo " - Health check: ✅" || echo " - Health check: ❌"        echo 'Server started successfully'

        

rm -f /tmp/nettestlab.service        # Check if port is open
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