#!/bin/bash

# NetTestLab OpenWRT Package Builder
# Builds IPK package for OpenWRT installation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PROJECT_ROOT="/Users/hector/NetTestLab"
OPENWRT_DIR="$PROJECT_ROOT/openwrt"
BUILD_DIR="$OPENWRT_DIR/build"
PACKAGE_DIR="$BUILD_DIR/package"
CONTROL_DIR="$PACKAGE_DIR/CONTROL"
DATA_DIR="$PACKAGE_DIR/data"

# Package metadata
PACKAGE_NAME="nettestlab"
VERSION="1.0.0"
RELEASE="1"
ARCHITECTURE="aarch64"  # Target router architecture
PACKAGE_FILE="${PACKAGE_NAME}_${VERSION}-${RELEASE}_${ARCHITECTURE}.ipk"

echo -e "${BLUE}🔨 NetTestLab OpenWRT Package Builder${NC}"
echo "========================================"

# Clean previous builds
echo -e "${YELLOW}🧹 Cleaning previous builds...${NC}"
rm -rf "$BUILD_DIR"
mkdir -p "$CONTROL_DIR" "$DATA_DIR"

# Build the Go binary for ARM64
echo -e "${YELLOW}📦 Building Go binary for ARM64...${NC}"
cd "$PROJECT_ROOT"
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o "$DATA_DIR/usr/bin/nettestlab" ./cmd/server

if [ ! -f "$DATA_DIR/usr/bin/nettestlab" ]; then
    echo -e "${RED}❌ Failed to build binary${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Binary built successfully${NC}"

# Copy package files
echo -e "${YELLOW}📁 Copying package files...${NC}"

# Create directory structure
mkdir -p "$DATA_DIR/usr/bin"
mkdir -p "$DATA_DIR/etc/init.d"
mkdir -p "$DATA_DIR/etc/config"
mkdir -p "$DATA_DIR/etc/uci-defaults"

# Copy files
cp "$OPENWRT_DIR/files/etc/init.d/nettestlab" "$DATA_DIR/etc/init.d/"
cp "$OPENWRT_DIR/files/etc/config/nettestlab" "$DATA_DIR/etc/config/"
cp "$OPENWRT_DIR/files/etc/uci-defaults/nettestlab" "$DATA_DIR/etc/uci-defaults/"

# Set correct permissions
chmod 755 "$DATA_DIR/usr/bin/nettestlab"
chmod 755 "$DATA_DIR/etc/init.d/nettestlab"
chmod 644 "$DATA_DIR/etc/config/nettestlab"
chmod 755 "$DATA_DIR/etc/uci-defaults/nettestlab"

# Create control file
echo -e "${YELLOW}📝 Creating control file...${NC}"
cat > "$CONTROL_DIR/control" << EOF
Package: $PACKAGE_NAME
Version: $VERSION-$RELEASE
Depends: tc-bpf, kmod-sched-core, kmod-ifb
License: MIT
Section: net
Architecture: $ARCHITECTURE
Installed-Size: $(du -k "$DATA_DIR" | tail -1 | cut -f1)
Maintainer: NetTestLab Team
Description: Network Testing Laboratory gRPC Server
 NetTestLab is a gRPC server for controlling network conditions on OpenWRT routers.
 It provides APIs to simulate various network conditions like latency, packet loss,
 bandwidth limitations, and jitter for automated mobile application testing.
 .
 Features:
 - gRPC API for network control
 - Built-in network profiles (2G, 3G, 4G, 5G, WiFi, Satellite)
 - Real-time network monitoring
 - Traffic shaping using Linux tc (netem)
 - Cross-platform client libraries support
EOF

# Create conffiles
echo "/etc/config/nettestlab" > "$CONTROL_DIR/conffiles"

# Create postinst script
cat > "$CONTROL_DIR/postinst" << 'EOF'
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
cat > "$CONTROL_DIR/prerm" << 'EOF'
#!/bin/sh
if [ -z "${IPKG_INSTROOT}" ]; then
	echo "Stopping NetTestLab service..."
	/etc/init.d/nettestlab stop >/dev/null 2>&1 || true
	/etc/init.d/nettestlab disable >/dev/null 2>&1 || true
fi
exit 0
EOF

# Create postrm script
cat > "$CONTROL_DIR/postrm" << 'EOF'
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
chmod 755 "$CONTROL_DIR/postinst"
chmod 755 "$CONTROL_DIR/prerm" 
chmod 755 "$CONTROL_DIR/postrm"

# Create data.tar.gz
echo -e "${YELLOW}📦 Creating package structure...${NC}"

# Instead of creating tar files manually, prepare the package directory structure
# that opkg-build expects
PACKAGE_ROOT="$BUILD_DIR/package_root"
rm -rf "$PACKAGE_ROOT"
mkdir -p "$PACKAGE_ROOT"

# Copy data files to package root
cp -r "$DATA_DIR"/* "$PACKAGE_ROOT/"

# Copy control files to DEBIAN directory (opkg-build expects this structure)
mkdir -p "$PACKAGE_ROOT/DEBIAN"
cp -r "$CONTROL_DIR"/* "$PACKAGE_ROOT/DEBIAN/"

# Create the IPK package using opkg-build
echo -e "${YELLOW}📦 Creating IPK package with opkg-build...${NC}"
cd "$BUILD_DIR"

# Check if opkg-build is available
if command -v opkg-build >/dev/null 2>&1; then
    # Use opkg-build (preferred method)
    opkg-build "$PACKAGE_ROOT" .
    
    # Find the generated package
    GENERATED_PACKAGE=$(find . -name "*.ipk" -type f | head -1)
    if [ -n "$GENERATED_PACKAGE" ]; then
        mv "$GENERATED_PACKAGE" "$PACKAGE_FILE"
    fi
else
    echo -e "${YELLOW}⚠️  opkg-build not found, attempting manual creation...${NC}"
    
    # Create tarballs manually as fallback
    cd "$PACKAGE_ROOT"
    tar -czf "$BUILD_DIR/data.tar.gz" --exclude=DEBIAN .
    
    cd "$PACKAGE_ROOT/DEBIAN"
    tar -czf "$BUILD_DIR/control.tar.gz" .
    
    cd "$BUILD_DIR"
    echo "2.0" > debian-binary
    
    # Check if archives exist
    if [ ! -f "control.tar.gz" ] || [ ! -f "data.tar.gz" ]; then
        echo -e "${RED}❌ Missing archive files${NC}"
        ls -la
        exit 1
    fi
    
    # Use ar with specific options for OpenWrt compatibility
    if command -v ar >/dev/null 2>&1; then
        # Create IPK using ar (preferred method)
        ar rcs "$PACKAGE_FILE" debian-binary control.tar.gz data.tar.gz
    else
        # Fallback: create using tar (less ideal but works)
        echo -e "${YELLOW}⚠️  ar not found, using tar fallback${NC}"
        tar -cf "$PACKAGE_FILE" debian-binary control.tar.gz data.tar.gz
    fi
fi

if [ -f "$PACKAGE_FILE" ]; then
    echo -e "${GREEN}✅ Package created successfully: $PACKAGE_FILE${NC}"
    echo -e "${GREEN}📏 Package size: $(du -h "$PACKAGE_FILE" | cut -f1)${NC}"
    
    # Move to project root for easy access
    mv "$PACKAGE_FILE" "$PROJECT_ROOT/"
    echo -e "${GREEN}📦 Package available at: $PROJECT_ROOT/$PACKAGE_FILE${NC}"
else
    echo -e "${RED}❌ Failed to create package${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}📋 Package Information:${NC}"
echo "  Name: $PACKAGE_NAME"
echo "  Version: $VERSION-$RELEASE" 
echo "  Architecture: $ARCHITECTURE"
echo "  File: $PACKAGE_FILE"
echo ""
echo -e "${BLUE}🚀 Next steps:${NC}"
echo "  1. Copy package to router: scp $PACKAGE_FILE root@192.168.1.4:/tmp/"
echo "  2. Install on router: ssh root@192.168.1.4 'opkg install /tmp/$PACKAGE_FILE'"
echo "  3. Start service: ssh root@192.168.1.4 '/etc/init.d/nettestlab start'"