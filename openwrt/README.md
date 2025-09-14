# NetTestLab OpenWRT Package

This directory contains the OpenWRT package configuration for NetTestLab, a network simulation server for automated mobile testing.

## Package Structure

```
openwrt/
├── Makefile              # OpenWRT package build configuration
├── files/
│   ├── etc/
│   │   ├── config/nettestlab    # Configuration file
│   │   └── init.d/nettestlab    # Init script for service
│   └── usr/
│       └── bin/nettestlab       # Binary location (installed by package)
├── patches/              # Any patches needed
└── src/                  # Source code (copied during build)
```

## Dependencies

The package automatically installs required dependencies:
- `tc` (traffic control) - Part of iproute2 package
- `kmod-sched-core` - Kernel modules for traffic scheduling  
- `kmod-ifb` - Intermediate Functional Block device
- `kmod-act-police` - Traffic policing action

## Installation

1. Build the package:
```bash
make package/nettestlab/compile
```

2. Install on router:
```bash
opkg install nettestlab_*.ipk
```

3. Start the service:
```bash
/etc/init.d/nettestlab start
```

## Configuration

Edit `/etc/config/nettestlab` to configure the server:
- Port (default: 8080)
- Log level
- Interfaces to manage