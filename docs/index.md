# NetTestLab Documentation

<div align="center">

![NetTestLab Logo](assets/logo.png){ width="200" }

**A comprehensive network traffic control system for mobile application testing**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![OpenWrt](https://img.shields.io/badge/OpenWrt-23.05+-green.svg)](https://openwrt.org)
[![gRPC](https://img.shields.io/badge/gRPC-1.50+-blue.svg)](https://grpc.io)

[Get Started](getting-started/overview.md){ .md-button .md-button--primary }
[View on GitHub](https://github.com/Eitol/NetTestLab){ .md-button }

</div>

## Overview

NetTestLab is a powerful network traffic control system designed specifically for mobile application testing. It allows developers and QA engineers to simulate various network conditions (2G, 3G, 4G, 5G, WiFi, satellite) directly on OpenWrt routers, providing realistic testing environments for mobile applications.

## Key Features

### 🌐 WiFi Auto-Discovery
Use `"wifi"` as interface name to automatically target all WiFi interfaces without manual configuration.

### 📱 Mobile Network Simulation
Pre-configured profiles for 2G, 3G, 4G, 5G, and satellite networks with realistic latency, bandwidth, and packet loss characteristics.

### 🎛️ Advanced Traffic Control
- **Latency simulation** with configurable delays
- **Packet loss** with random and burst patterns
- **Bandwidth limiting** for upload and download
- **Jitter simulation** with uniform distribution
- **Packet corruption** for edge case testing

### 🔧 OpenWrt Integration
Native OpenWrt package with proper init scripts, UCI configuration, and automatic dependency management.

### 🚀 High-Performance gRPC API
Protocol Buffers-based API for efficient communication and multi-language client support.

### 📊 Real-time Monitoring
System health monitoring, interface status tracking, and condition monitoring with metrics streaming.

### 🏗️ Multi-language Support
Client libraries for Go, JavaScript/TypeScript, Python, Java, and Dart/Flutter.

## Architecture

```mermaid
graph TB
    subgraph "Client Applications"
        MA[Mobile Apps]
        CA[Client Apps]
        WI[Web Interface]
    end
    
    subgraph "gRPC Clients"
        GC[Go Client]
        JS[JavaScript Client]
        PY[Python Client]
        JA[Java Client]
        DA[Dart Client]
    end
    
    subgraph "OpenWrt Router"
        subgraph "NetTestLab Server"
            NS[Network Service]
            PS[Profile Service]
            MS[Monitoring Service]
        end
        
        subgraph "Traffic Control Engine"
            TC[tc (Traffic Control)]
            IF[Interface Management]
            QD[Queue Disciplines]
        end
        
        subgraph "Network Interfaces"
            WF[WiFi Interfaces]
            ET[Ethernet Interfaces]
            BR[Bridge Interfaces]
        end
    end
    
    MA --> GC
    CA --> JS
    CA --> PY
    WI --> JS
    
    GC --> NS
    JS --> PS
    PY --> MS
    JA --> NS
    DA --> PS
    
    NS --> TC
    PS --> TC
    MS --> IF
    
    TC --> QD
    IF --> WF
    IF --> ET
    IF --> BR
```

## Quick Start

Get up and running with NetTestLab in minutes:

1. **[Install on OpenWrt Router](getting-started/installation.md)**
2. **[Configure Basic Settings](getting-started/configuration.md)**
3. **[Run Your First Test](getting-started/quickstart.md)**

## Use Cases

### Mobile App Testing
Test how your mobile applications perform under various network conditions:

- **Slow connections**: 2G/3G simulation for emerging markets
- **High latency**: Satellite network simulation
- **Unstable connections**: Packet loss and jitter simulation
- **Bandwidth limits**: Test with limited upload/download speeds

### CI/CD Integration
Integrate network testing into your continuous integration pipeline:

- **Automated testing** with different network profiles
- **Performance regression** detection
- **Multi-environment** testing scenarios

### Development Testing
Local development with realistic network conditions:

- **API timeout testing** with high latency simulation
- **Offline behavior** testing with complete network cuts
- **Progressive loading** testing with bandwidth limits

## Community & Support

- **GitHub Repository**: [Eitol/NetTestLab](https://github.com/Eitol/NetTestLab)
- **Issues & Bug Reports**: [GitHub Issues](https://github.com/Eitol/NetTestLab/issues)
- **Discussions**: [GitHub Discussions](https://github.com/Eitol/NetTestLab/discussions)
- **Documentation**: You're reading it! 📚

## License

NetTestLab is released under the [MIT License](https://opensource.org/licenses/MIT).

---

<div align="center">
Made with ❤️ for mobile app developers and QA engineers who need realistic network testing environments.
</div>