#!/bin/bash

# NetTestLab Simple Package Builder
# Creates IPK package using traditional OpenWRT methods

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PROJECT_ROOT="/Users/hector/NetTestLab"
BUILD_DIR="/tmp/nettestlab-package-build"
PACKAGE_NAME="nettestlab"
VERSION="1.0.0"
RELEASE="1"
ARCHITECTURE="aarch64"
PACKAGE_FILE="${PACKAGE_NAME}_${VERSION}-${RELEASE}_${ARCHITECTURE}.ipk"

echo -e "${BLUE}🔨 NetTestLab Simple Package Builder${NC}"
echo "========================================"

# Clean and create build directory
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

# Build the Go binary for ARM64
echo -e "${YELLOW}📦 Building Go binary for ARM64...${NC}"
cd "$PROJECT_ROOT"
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o "$BUILD_DIR/nettestlab" ./cmd/server

if [ ! -f "$BUILD_DIR/nettestlab" ]; then
    echo -e "${RED}❌ Failed to build binary${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Binary built successfully${NC}"

# Create package directory structure
echo -e "${YELLOW}📁 Creating package structure...${NC}"
cd "$BUILD_DIR"
mkdir -p ipkg/usr/bin
mkdir -p ipkg/etc/init.d
mkdir -p ipkg/etc/config
mkdir -p ipkg/etc/uci-defaults
mkdir -p ipkg/CONTROL

# Copy binary and files
cp nettestlab ipkg/usr/bin/
cp "$PROJECT_ROOT/openwrt/files/etc/init.d/nettestlab" ipkg/etc/init.d/
cp "$PROJECT_ROOT/openwrt/files/etc/config/nettestlab" ipkg/etc/config/
cp "$PROJECT_ROOT/openwrt/files/etc/uci-defaults/nettestlab" ipkg/etc/uci-defaults/

# Set permissions
chmod 755 ipkg/usr/bin/nettestlab
chmod 755 ipkg/etc/init.d/nettestlab
chmod 644 ipkg/etc/config/nettestlab
chmod 755 ipkg/etc/uci-defaults/nettestlab

# Create control file
echo -e "${YELLOW}📝 Creating control file...${NC}"
INSTALLED_SIZE=$(du -k ipkg | tail -1 | cut -f1)

cat > ipkg/CONTROL/control << EOF
Package: $PACKAGE_NAME
Version: $VERSION-$RELEASE
Depends: tc-bpf, kmod-sched-core, kmod-ifb
License: MIT
Section: net
Architecture: $ARCHITECTURE
Installed-Size: $INSTALLED_SIZE
Maintainer: NetTestLab Team
Description: Network Testing Laboratory gRPC Server
 NetTestLab is a gRPC server for controlling network conditions on OpenWRT routers.
 It provides APIs to simulate various network conditions like latency, packet loss,
 bandwidth limitations, and jitter for automated mobile application testing.
EOF

# Create conffiles
echo "/etc/config/nettestlab" > ipkg/CONTROL/conffiles

# Create postinst script
cat > ipkg/CONTROL/postinst << 'EOF'
#!/bin/sh
if [ -z "${IPKG_INSTROOT}" ]; then
	echo "Enabling NetTestLab service..."
	/etc/init.d/nettestlab enable
	
	# Run uci-defaults
	if [ -f /etc/uci-defaults/nettestlab ]; then
		/bin/sh /etc/uci-defaults/nettestlab
		rm -f /etc/uci-defaults/nettestlab
	fi
	
	echo "NetTestLab installed successfully!"
	echo "Configure: /etc/config/nettestlab"
	echo "Start: /etc/init.d/nettestlab start"
fi
exit 0
EOF

# Create prerm script
cat > ipkg/CONTROL/prerm << 'EOF'
#!/bin/sh
if [ -z "${IPKG_INSTROOT}" ]; then
	echo "Stopping NetTestLab service..."
	/etc/init.d/nettestlab stop >/dev/null 2>&1 || true
	/etc/init.d/nettestlab disable >/dev/null 2>&1 || true
fi
exit 0
EOF

# Create postrm script
cat > ipkg/CONTROL/postrm << 'EOF'
#!/bin/sh
if [ -z "${IPKG_INSTROOT}" ]; then
	# Remove firewall rule
	uci -q delete firewall.nettestlab
	uci commit firewall >/dev/null 2>&1 || true
	
	# Remove log files
	rm -f /var/log/nettestlab.log*
	rm -f /etc/logrotate.d/nettestlab
	
	echo "NetTestLab removed completely."
fi
exit 0
EOF

# Set permissions for control scripts
chmod 755 ipkg/CONTROL/postinst
chmod 755 ipkg/CONTROL/prerm 
chmod 755 ipkg/CONTROL/postrm

# Create package using traditional tar + gzip method
echo -e "${YELLOW}📦 Creating IPK package...${NC}"

# Create data.tar.gz (all files except CONTROL)
cd ipkg
tar --exclude=CONTROL -czf ../data.tar.gz .

# Create control.tar.gz
cd CONTROL
tar -czf ../../control.tar.gz .
cd ..

# Create the IPK using simple tar
cd "$BUILD_DIR"
echo "2.0" > debian-binary

# Use tar instead of ar for better compatibility
tar -czf "$PACKAGE_FILE" debian-binary control.tar.gz data.tar.gz

if [ -f "$PACKAGE_FILE" ]; then
    echo -e "${GREEN}✅ Package created successfully: $PACKAGE_FILE${NC}"
    echo -e "${GREEN}📏 Package size: $(du -h "$PACKAGE_FILE" | cut -f1)${NC}"
    
    # Move to project root for easy access
    mv "$PACKAGE_FILE" "$PROJECT_ROOT/"
    echo -e "${GREEN}📦 Package available at: $PROJECT_ROOT/$PACKAGE_FILE${NC}"
    
    # Show package info
    echo ""
    echo -e "${BLUE}📋 Package Information:${NC}"
    echo "  Name: $PACKAGE_NAME"
    echo "  Version: $VERSION-$RELEASE" 
    echo "  Architecture: $ARCHITECTURE"
    echo "  File: $PACKAGE_FILE"
    echo "  Format: Traditional tar.gz (OpenWRT compatible)"
    
else
    echo -e "${RED}❌ Failed to create package${NC}"
    exit 1
fi

# Cleanup
cd "$PROJECT_ROOT"
rm -rf "$BUILD_DIR"

echo ""
echo -e "${GREEN}✅ Simple package build completed!${NC}"