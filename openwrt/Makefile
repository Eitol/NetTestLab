#
# Copyright (C) 2025 NetTestLab
#
# This is free software, licensed under the MIT License.
#

include $(TOPDIR)/rules.mk

PKG_NAME:=nettestlab
PKG_VERSION:=1.0.0
PKG_RELEASE:=1

PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)-$(PKG_VERSION)
PKG_SOURCE_SUBDIR:=$(PKG_NAME)-$(PKG_VERSION)

include $(INCLUDE_DIR)/package.mk

define Package/nettestlab
	SECTION:=net
	CATEGORY:=Network
	TITLE:=Network Testing Laboratory gRPC Server
	DEPENDS:=+tc-bpf +kmod-sched-core +kmod-ifb
	MAINTAINER:=NetTestLab Team
	URL:=https://github.com/Eitol/NetTestLab
endef

define Package/nettestlab/description
	NetTestLab is a gRPC server for controlling network conditions on OpenWRT routers.
	It provides APIs to simulate various network conditions like latency, packet loss,
	bandwidth limitations, and jitter for automated mobile application testing.
	
	Features:
	- gRPC API for network control
	- Built-in network profiles (2G, 3G, 4G, 5G, WiFi, Satellite)
	- Real-time network monitoring
	- Traffic shaping using Linux tc (netem)
	- Cross-platform client libraries support
endef

define Package/nettestlab/conffiles
/etc/config/nettestlab
endef

define Build/Configure
endef

define Build/Compile
	# Cross-compile for the target architecture
	cd $(PKG_BUILD_DIR) && \
	GOOS=linux GOARCH=$(if $(CONFIG_aarch64),arm64,$(if $(CONFIG_arm),arm,$(if $(CONFIG_mips),mips,$(if $(CONFIG_mipsel),mipsle,$(if $(CONFIG_x86_64),amd64,$(CONFIG_ARCH)))))) \
	CGO_ENABLED=0 \
	go build -ldflags="-s -w" -o nettestlab ./cmd/server
endef

define Package/nettestlab/install
	# Install binary
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/nettestlab $(1)/usr/bin/

	# Install init script
	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) ./files/etc/init.d/nettestlab $(1)/etc/init.d/

	# Install configuration
	$(INSTALL_DIR) $(1)/etc/config
	$(INSTALL_CONF) ./files/etc/config/nettestlab $(1)/etc/config/

	# Install uci defaults
	$(INSTALL_DIR) $(1)/etc/uci-defaults
	$(INSTALL_BIN) ./files/etc/uci-defaults/nettestlab $(1)/etc/uci-defaults/
endef

define Package/nettestlab/postinst
#!/bin/sh
if [ -z "$${IPKG_INSTROOT}" ]; then
	echo "Enabling NetTestLab service..."
	/etc/init.d/nettestlab enable
	echo "NetTestLab installed successfully!"
	echo "Configure: /etc/config/nettestlab"
	echo "Start: /etc/init.d/nettestlab start"
fi
endef

define Package/nettestlab/prerm
#!/bin/sh
if [ -z "$${IPKG_INSTROOT}" ]; then
	echo "Stopping NetTestLab service..."
	/etc/init.d/nettestlab stop >/dev/null 2>&1 || true
	/etc/init.d/nettestlab disable >/dev/null 2>&1 || true
fi
endef

# Prepare source code
define Build/Prepare
	$(INSTALL_DIR) $(PKG_BUILD_DIR)
	$(CP) $(CURDIR)/../* $(PKG_BUILD_DIR)/
endef

$(eval $(call BuildPackage,nettestlab))