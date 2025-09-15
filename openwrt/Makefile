# OpenWrt package Makefile for NetTestLab
include $(TOPDIR)/rules.mk

PKG_NAME:=nettestlab
PKG_VERSION:=1.0.0
PKG_RELEASE:=1

PKG_SOURCE:=$(PKG_NAME)-$(PKG_VERSION).tar.gz
PKG_SOURCE_URL:=https://github.com/Eitol/NetTestLab/archive/v$(PKG_VERSION)/
PKG_HASH:=skip
PKG_SOURCE_SUBDIR:=$(PKG_NAME)-$(PKG_VERSION)

PKG_LICENSE:=MIT
PKG_LICENSE_FILES:=LICENSE
PKG_MAINTAINER:=Eitol <hector@example.com>

PKG_BUILD_DEPENDS:=golang/host
PKG_BUILD_PARALLEL:=1
PKG_USE_MIPS16:=0

GO_PKG:=github.com/Eitol/NetTestLab
GO_PKG_BUILD_PKG:=github.com/Eitol/NetTestLab/cmd/server
GO_PKG_LDFLAGS_X:=main.version=v$(PKG_VERSION)

include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

define Package/nettestlab
  SECTION:=net
  CATEGORY:=Network
  TITLE:=Network traffic control and simulation tool
  URL:=https://github.com/Eitol/NetTestLab
  DEPENDS:=$(GO_ARCH_DEPENDS) +tc +kmod-sched-core +kmod-ifb +kmod-netem
  PKGARCH:=all
endef

define Package/nettestlab/description
  NetTestLab is a network traffic control and simulation tool for OpenWrt routers.
  It provides a gRPC API to control network conditions like latency, bandwidth,
  and packet loss for testing mobile applications and network scenarios.
  
  Features:
  - Real-time network condition control via gRPC API
  - Built-in profiles for common mobile network scenarios (2G, 3G, 4G, WiFi)
  - Custom profile creation and management
  - System monitoring and metrics collection
  - Cross-platform client libraries (Go, Python, JavaScript, Java, Dart)
endef

define Package/nettestlab/conffiles
/etc/config/nettestlab
endef

define Build/Compile
	$(call GoPackage/Build/Compile)
endef

define Package/nettestlab/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(GO_PKG_BUILD_BIN_DIR)/server $(1)/usr/bin/nettestlab

	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) ./files/etc/init.d/nettestlab $(1)/etc/init.d/nettestlab

	$(INSTALL_DIR) $(1)/etc/config
	$(INSTALL_CONF) ./files/etc/config/nettestlab $(1)/etc/config/nettestlab

	$(INSTALL_DIR) $(1)/etc/uci-defaults
	$(INSTALL_BIN) ./files/etc/uci-defaults/nettestlab $(1)/etc/uci-defaults/nettestlab
endef

define Package/nettestlab/postinst
#!/bin/sh
[ -n "$${IPKG_INSTROOT}" ] || {
	# Enable and start the service
	/etc/init.d/nettestlab enable
	/etc/init.d/nettestlab start
	
	# Create data directory for profiles if it doesn't exist
	[ -d /var/lib/nettestlab ] || mkdir -p /var/lib/nettestlab
	
	echo "NetTestLab installed successfully!"
	echo "Access the gRPC API at: your-router-ip:8080"
	echo "View logs with: logread | grep nettestlab"
	echo "Built-in profiles will be created automatically on first start"
}
exit 0
endef

define Package/nettestlab/prerm
#!/bin/sh
[ -n "$${IPKG_INSTROOT}" ] || {
	/etc/init.d/nettestlab stop
	/etc/init.d/nettestlab disable
}
exit 0
endef

define Package/nettestlab/postrm
#!/bin/sh
[ -n "$${IPKG_INSTROOT}" ] || {
	# Remove data directory if user wants to
	echo "To completely remove all NetTestLab data, run:"
	echo "rm -rf /var/lib/nettestlab"
}
exit 0
endef

$(eval $(call GoBinPackage,nettestlab))
$(eval $(call BuildPackage,nettestlab))