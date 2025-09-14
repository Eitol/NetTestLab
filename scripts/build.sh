#!/bin/bash

# NetTestLab Build Script
# Builds the gRPC server and client libraries

set -e

echo "🚀 Building NetTestLab..."

# Generate Protocol Buffer files
echo "📦 Generating protobuf files..."
buf generate

# Build Go server
echo "🔨 Building Go server..."
go build -o bin/nettestlab-server ./cmd/server

# Build client libraries
echo "📚 Building client libraries..."

# JavaScript/TypeScript
if command -v npm &> /dev/null; then
    echo "  - Building JavaScript/TypeScript client..."
    cd clients/javascript
    npm install
    npm run build
    cd ../..
fi

# Python
if command -v python3 &> /dev/null; then
    echo "  - Building Python client..."
    cd clients/python
    python3 setup.py build
    cd ../..
fi

# Java
if command -v mvn &> /dev/null; then
    echo "  - Building Java client..."
    cd clients/java
    mvn clean compile
    cd ../..
fi

# Dart/Flutter
if command -v dart &> /dev/null; then
    echo "  - Building Dart client..."
    cd clients/dart
    dart pub get
    cd ../..
fi

# Go client
echo "  - Building Go client..."
cd clients/go
go mod tidy
go build ./...
cd ../..

echo "✅ Build completed successfully!"
echo "Server binary: bin/nettestlab-server"