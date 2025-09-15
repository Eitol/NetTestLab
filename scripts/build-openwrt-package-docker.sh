#!/bin/bash

# NetTestLab OpenWRT Package Builder (Docker version)
# Builds IPK package inside Docker container with proper opkg-build

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Package metadata (from environment variables)
PACKAGE_NAME="${PACKAGE_NAME:-nettestlab}"
VERSION="${VERSION:-1.0.0}"
RELEASE="${RELEASE:-1}"
ARCHITECTURE="${ARCHITECTURE:-aarch64}"
PACKAGE_FILE="${PACKAGE_NAME}_${VERSION}-${RELEASE}_${ARCHITECTURE}.ipk"

echo -e "${BLUE}🔨 NetTestLab OpenWRT Package Builder (Docker)${NC}"
echo "=================================================="
echo "Architecture: $ARCHITECTURE"
echo "Package: $PACKAGE_FILE"

# Verify binary exists
if [ ! -f "/workspace/nettestlab" ]; then
    echo -e "${RED}❌ Binary not found: /workspace/nettestlab${NC}"
    exit 1
fi

# Create package directory structure
PACKAGE_DIR="/tmp/package"
rm -rf "$PACKAGE_DIR"
mkdir -p "$PACKAGE_DIR"

# Create data directory structure
mkdir -p "$PACKAGE_DIR/usr/bin"
mkdir -p "$PACKAGE_DIR/etc/init.d"
mkdir -p "$PACKAGE_DIR/etc/config"
mkdir -p "$PACKAGE_DIR/etc/uci-defaults"

# Copy binary
echo -e "${YELLOW}📦 Copying binary...${NC}"
cp "/workspace/nettestlab" "$PACKAGE_DIR/usr/bin/"
chmod 755 "$PACKAGE_DIR/usr/bin/nettestlab"

# Copy configuration files
echo -e "${YELLOW}📁 Copying configuration files...${NC}"
if [ -f "/workspace/files/etc/init.d/nettestlab" ]; then
    cp "/workspace/files/etc/init.d/nettestlab" "$PACKAGE_DIR/etc/init.d/"
    chmod 755 "$PACKAGE_DIR/etc/init.d/nettestlab"
fi

if [ -f "/workspace/files/etc/config/nettestlab" ]; then
    cp "/workspace/files/etc/config/nettestlab" "$PACKAGE_DIR/etc/config/"
    chmod 644 "$PACKAGE_DIR/etc/config/nettestlab"
fi

if [ -f "/workspace/files/etc/uci-defaults/nettestlab" ]; then
    cp "/workspace/files/etc/uci-defaults/nettestlab" "$PACKAGE_DIR/etc/uci-defaults/"
    chmod 755 "$PACKAGE_DIR/etc/uci-defaults/nettestlab"
fi

# Create DEBIAN directory for control files
mkdir -p "$PACKAGE_DIR/DEBIAN"

# Create control file
echo -e "${YELLOW}📝 Creating control file...${NC}"
INSTALLED_SIZE=$(du -k "$PACKAGE_DIR" | tail -1 | cut -f1)
cat > "$PACKAGE_DIR/DEBIAN/control" << EOF
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
 .
 Features:
 - gRPC API for network control
 - Built-in network profiles (2G, 3G, 4G, 5G, WiFi, Satellite)
 - Real-time network monitoring
 - Traffic shaping using Linux tc (netem)
 - Cross-platform client libraries support
EOF

# Create conffiles
if [ -f "$PACKAGE_DIR/etc/config/nettestlab" ]; then
    echo "/etc/config/nettestlab" > "$PACKAGE_DIR/DEBIAN/conffiles"
fi

# Create postinst script
cat > "$PACKAGE_DIR/DEBIAN/postinst" << 'EOF'
#!/bin/sh
if [ -z "${IPKG_INSTROOT}" ]; then
	echo "Enabling NetTestLab service..."
	/etc/init.d/nettestlab enable
	
	# Handle configuration file conflicts
	if [ -f /etc/config/nettestlab-opkg ]; then
		echo "Note: New configuration saved as /etc/config/nettestlab-opkg"
		echo "Compare with existing config: diff /etc/config/nettestlab /etc/config/nettestlab-opkg"
	fi
	
	# Run uci-defaults with error handling
	if [ -f /etc/uci-defaults/nettestlab ]; then
		echo "Running post-installation configuration..."
		if /bin/sh /etc/uci-defaults/nettestlab; then
			rm -f /etc/uci-defaults/nettestlab
			echo "Configuration completed successfully"
		else
			echo "Warning: Some configuration steps failed, but service should still work"
			rm -f /etc/uci-defaults/nettestlab
		fi
	fi
	
	echo "NetTestLab installed successfully!"
	echo "Configure: /etc/config/nettestlab"
	echo "Start: /etc/init.d/nettestlab start"
fi
exit 0
EOF

# Create prerm script
cat > "$PACKAGE_DIR/DEBIAN/prerm" << 'EOF'
#!/bin/sh
if [ -z "${IPKG_INSTROOT}" ]; then
	echo "Stopping NetTestLab service..."
	/etc/init.d/nettestlab stop >/dev/null 2>&1 || true
	/etc/init.d/nettestlab disable >/dev/null 2>&1 || true
fi
exit 0
EOF

# Create postrm script
cat > "$PACKAGE_DIR/DEBIAN/postrm" << 'EOF'
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
chmod 755 "$PACKAGE_DIR/DEBIAN/postinst"
chmod 755 "$PACKAGE_DIR/DEBIAN/prerm" 
chmod 755 "$PACKAGE_DIR/DEBIAN/postrm"

# Build the package using opkg-build
echo -e "${YELLOW}📦 Building IPK package with opkg-build...${NC}"
cd /tmp

# Run opkg-build
if opkg-build "$PACKAGE_DIR" "/tmp" ; then
    if [ -f "/tmp/$PACKAGE_FILE" ]; then
        echo -e "${GREEN}✅ Package created successfully: $PACKAGE_FILE${NC}"
        echo -e "${GREEN}📏 Package size: $(du -h "/tmp/$PACKAGE_FILE" | cut -f1)${NC}"
        
        # Copy to output directory
        cp "/tmp/$PACKAGE_FILE" "/workspace/"
        echo -e "${GREEN}📦 Package available at: /workspace/$PACKAGE_FILE${NC}"
    else
        echo -e "${RED}❌ Package file not found after build${NC}"
        echo "Files in /tmp:"
        ls -la /tmp/*.ipk 2>/dev/null || echo "No IPK files found"
        exit 1
    fi
else
    echo -e "${RED}❌ opkg-build failed${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}📋 Package Information:${NC}"
echo "  Name: $PACKAGE_NAME"
echo "  Version: $VERSION-$RELEASE" 
echo "  Architecture: $ARCHITECTURE"
echo "  File: $PACKAGE_FILE"
echo ""
echo -e "${GREEN}🎉 Package built successfully!${NC}"