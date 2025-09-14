# NetTestLab

Network simulation platform for mobile application testing.

## Overview

NetTestLab provides precise control over network conditions for testing mobile applications on real device farms. Through a unified gRPC API, you can simulate various network conditions including latency, packet loss, bandwidth limitations, and jitter.

## Key Features

- **Real-time Network Control**: Apply network conditions instantly
- **Predefined Profiles**: Ready-to-use configurations for 2G, 3G, 4G, 5G networks  
- **Multi-language Support**: Client libraries for JavaScript, Python, Java, Dart, and Go
- **OpenWRT Integration**: Easy installation on router hardware
- **Traffic Shaping**: Advanced control using Linux tc and netem

## Architecture

The system consists of:

- **gRPC Server**: Core network control service running on OpenWRT
- **Client Libraries**: Generated from Protocol Buffers for multiple languages
- **Network Controller**: Interface to Linux traffic control subsystem
- **Profile Manager**: Predefined and custom network condition presets

## Getting Started

1. [Install the server](installation/index.md) on your OpenWRT router
2. [Choose a client library](clients/javascript.md) for your preferred language
3. [Follow the examples](examples/basic.md) to start controlling network conditions

## Support

- GitHub Issues: [Report bugs and request features](https://github.com/Eitol/NetTestLab/issues)
- Documentation: This site contains comprehensive guides and API reference
- Examples: Working code samples for common use cases