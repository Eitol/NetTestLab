#!/bin/bash

# NetTestLab Test Script
# Tests the gRPC server locally before deployment

set -e

echo "🚀 Starting NetTestLab Test Suite..."

# Build the server
echo "📦 Building server..."
go build -o bin/nettestlab-server ./cmd/server

# Start server in background
echo "🎯 Starting server..."
./bin/nettestlab-server --port 8080 &
SERVER_PID=$!

# Wait for server to start
echo "⏳ Waiting for server to start..."
sleep 3

# Test if server is responding
echo "🔍 Testing server health..."
if command -v grpcurl &> /dev/null; then
    echo "Testing with grpcurl..."
    grpcurl -plaintext localhost:8080 list
    echo "✅ Server is responding to gRPC requests"
else
    echo "⚠️  grpcurl not found, testing with nc..."
    if nc -z localhost 8080; then
        echo "✅ Server is listening on port 8080"
    else
        echo "❌ Server is not responding"
        kill $SERVER_PID
        exit 1
    fi
fi

# Clean shutdown
echo "🛑 Stopping server..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null || true

echo "✅ Local test completed successfully!"
echo ""
echo "📋 Next steps:"
echo "  1. Server builds and starts correctly"
echo "  2. gRPC port 8080 is accessible"
echo "  3. Ready for router deployment"