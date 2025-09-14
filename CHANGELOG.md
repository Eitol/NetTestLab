# Changelog

All notable changes to NetTestLab will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- WiFi auto-discovery feature - use "wifi" as interface name
- Comprehensive gRPC API with Protocol Buffers
- OpenWrt package with proper init scripts
- Multiple network simulation profiles (2G, 3G, 4G, 5G, WiFi, Satellite)
- Traffic control capabilities (latency, packet loss, bandwidth, jitter)
- Real-time monitoring and system health endpoints
- Multi-language client library support

### Changed
- Improved error handling and logging
- Enhanced configuration management with UCI integration
- Better service management with procd

### Fixed
- Service startup issues with incorrect command flags
- Interface resolution for wireless devices
- Memory and resource management improvements

## [1.0.0] - 2024-12-XX

### Added
- Initial release of NetTestLab
- Basic network traffic control functionality
- gRPC server implementation
- OpenWrt package support
- Command-line client tools
- Documentation and setup guides

### Features
- **Network Simulation**: Control latency, packet loss, bandwidth, and jitter
- **OpenWrt Integration**: Native package installation and service management
- **gRPC API**: High-performance cross-platform interface
- **WiFi Focus**: Specialized support for wireless interface testing
- **Mobile Testing**: Realistic network conditions for mobile app development

### Technical Details
- Go 1.21+ server implementation
- Protocol Buffers for API definitions
- Traffic control (tc) integration with Linux kernel
- UCI configuration system integration
- Procd service management

### Supported Platforms
- OpenWrt 23.05+ on ARM64 architecture
- Compatible with major OpenWrt router models
- Tested on real hardware deployments

---

For more information about each release, see the [GitHub Releases](https://github.com/yourusername/nettestlab/releases) page.