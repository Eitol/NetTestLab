# NetTestLab OpenWrt Package

This directory contains the OpenWrt package definition for NetTestLab.

## Package Structure

```
openwrt/
├── Makefile              # OpenWrt package Makefile
├── README.md            # This file
└── files/              # Package files
    └── etc/
        ├── config/
        │   └── nettestlab     # UCI configuration
        ├── init.d/
        │   └── nettestlab     # Procd init script
        └── uci-defaults/
            └── nettestlab     # Default configuration
```

## Building the Package

### Prerequisites

1. OpenWrt build environment
2. Go feed installed: `feeds/packages/lang/golang`

### Build Commands

```bash
# In OpenWrt build root
make package/nettestlab/compile V=s
```

### Installation

```bash
# Copy to router
scp bin/packages/*/packages/nettestlab_*.ipk root@router:/tmp/

# Install on router
opkg install /tmp/nettestlab_*.ipk
```

## Configuration

### UCI Configuration (`/etc/config/nettestlab`)

```bash
config nettestlab 'main'
    option enabled '1'
    option port '8080'
    option host '0.0.0.0'
    option log_level 'info'
    option profiles_dir '/var/lib/nettestlab/profiles'
    option data_dir '/var/lib/nettestlab'
```

### Service Management

```bash
# Start service
/etc/init.d/nettestlab start

# Stop service
/etc/init.d/nettestlab stop

# Enable auto-start
/etc/init.d/nettestlab enable

# Check status
/etc/init.d/nettestlab status
```

## Dependencies

### Runtime Dependencies

- `tc` - Traffic control utilities
- `kmod-sched-core` - Core traffic control kernel modules
- `kmod-ifb` - Intermediate Functional Block device
- `kmod-netem` - Network emulation kernel module

### Build Dependencies

- `golang/host` - Go compiler for host system

## Package Details

- **Category**: Network
- **Section**: net
- **License**: MIT
- **Size**: ~8MB (static Go binary)
- **Memory**: ~10-20MB runtime

## Files Installed

```
/usr/bin/nettestlab              # Main binary
/etc/init.d/nettestlab           # Init script
/etc/config/nettestlab           # UCI configuration
/etc/uci-defaults/nettestlab     # Default setup
```

## Automatic Setup

The package automatically:

1. Creates data directories (`/var/lib/nettestlab/`)
2. Sets up firewall rules (allows port 8080)
3. Enables and starts the service
4. Creates built-in profiles on first start

## Verification

After installation, verify the service:

```bash
# Check service status
/etc/init.d/nettestlab status

# Check logs
logread | grep nettestlab

# Test gRPC API
grpcurl -plaintext localhost:8080 list
```

Expected output:
```
nettestlab.v1.MonitoringService
nettestlab.v1.NetworkControlService
nettestlab.v1.ProfileService
```

## Troubleshooting

### Service Won't Start

1. Check configuration: `uci show nettestlab`
2. Check dependencies: `opkg list-installed | grep -E "(tc|kmod-sched|kmod-ifb|kmod-netem)"`
3. Check logs: `logread | grep nettestlab`

### Port Already in Use

Change port in configuration:
```bash
uci set nettestlab.main.port='9090'
uci commit nettestlab
/etc/init.d/nettestlab restart
```

### Profiles Not Created

Profiles are created automatically on first start. Check:
```bash
ls -la /var/lib/nettestlab/profiles/
```

## Package Submission

This package is designed for submission to the official OpenWrt packages feed.

See [OPENWRT_SUBMISSION.md](../OPENWRT_SUBMISSION.md) for detailed submission process.

## Support

- **GitHub**: https://github.com/Eitol/NetTestLab
- **Issues**: https://github.com/Eitol/NetTestLab/issues
- **Documentation**: https://eitol.github.io/NetTestLab/