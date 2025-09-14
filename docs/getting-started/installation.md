# Installation

This guide will walk you through installing NetTestLab on your OpenWrt router.

## Prerequisites

Before installing NetTestLab, ensure your router meets the following requirements:

### Hardware Requirements

- **Router**: OpenWrt 23.05+ compatible device
- **Architecture**: ARM64 or x86_64 
- **RAM**: Minimum 64MB available
- **Storage**: 10MB free space

### Software Requirements

Your router should have the following kernel modules (automatically installed):

- `kmod-sched-core` - Core traffic scheduling
- `kmod-ifb` - Intermediate functional block device  
- `kmod-netem` - Network emulation
- `tc-bpf` - Traffic control tools

## Installation Methods

### Method 1: IPK Package (Recommended)

Download and install the pre-built IPK package:

```bash
# Download the latest release
wget https://github.com/Eitol/NetTestLab/releases/latest/download/nettestlab_1.0.0_aarch64.ipk

# Install the package
opkg install nettestlab_1.0.0_aarch64.ipk
```

### Method 2: From Source

If you prefer to build from source:

```bash
# On your development machine
git clone https://github.com/Eitol/NetTestLab.git
cd NetTestLab

# Build for your router architecture
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o nettestlab cmd/server/main.go

# Copy to router
scp nettestlab root@your-router-ip:/usr/bin/

# On the router, create systemd service or init script
# (See configuration section for details)
```

## Post-Installation Setup

### 1. Enable and Start Service

```bash
# Enable service to start on boot
/etc/init.d/nettestlab enable

# Start the service
/etc/init.d/nettestlab start

# Check service status
/etc/init.d/nettestlab status
```

### 2. Verify Installation

Check that NetTestLab is running:

```bash
# Check if the process is running
ps | grep nettestlab

# Check if the port is listening
netstat -tulpn | grep :8080

# View service logs
logread | grep nettestlab
```

### 3. Test Basic Functionality

```bash
# Test with grpcurl (if available)
grpcurl -plaintext your-router-ip:8080 list

# Or test with curl (HTTP/REST gateway if enabled)
curl http://your-router-ip:8080/health
```

## Configuration

### UCI Configuration

NetTestLab uses UCI for configuration. Edit `/etc/config/nettestlab`:

```bash
uci set nettestlab.main.enabled='1'
uci set nettestlab.main.port='8080'
uci set nettestlab.main.host='0.0.0.0'
uci commit nettestlab
```

### Advanced Configuration

For advanced configuration options, see the [Configuration Guide](configuration.md).

## Firewall Configuration

If you have firewall rules, allow access to the NetTestLab port:

```bash
# Allow access from LAN
uci add firewall rule
uci set firewall.@rule[-1].name='Allow NetTestLab'
uci set firewall.@rule[-1].src='lan'
uci set firewall.@rule[-1].dest_port='8080'
uci set firewall.@rule[-1].proto='tcp'
uci set firewall.@rule[-1].target='ACCEPT'
uci commit firewall
/etc/init.d/firewall restart
```

## Troubleshooting

### Common Issues

#### Service Won't Start

```bash
# Check logs for errors
logread | grep nettestlab

# Verify dependencies are installed
opkg list-installed | grep -E "(tc-|kmod-)"

# Check configuration syntax
uci show nettestlab
```

#### Port Already in Use

```bash
# Check what's using port 8080
netstat -tulpn | grep :8080

# Change NetTestLab port if needed
uci set nettestlab.main.port='8081'
uci commit nettestlab
/etc/init.d/nettestlab restart
```

#### Permission Issues

```bash
# Ensure NetTestLab has proper permissions
chmod +x /usr/bin/nettestlab

# Check if tc command is available
which tc
tc qdisc show
```

### Getting Help

If you encounter issues:

1. Check the [Troubleshooting Guide](../openwrt/troubleshooting.md)
2. Search [GitHub Issues](https://github.com/Eitol/NetTestLab/issues)
3. Create a new issue with detailed logs and system information

## Next Steps

After successful installation:

1. **[Configure NetTestLab →](configuration.md)** - Set up basic configuration
2. **[Quick Start Tutorial →](quickstart.md)** - Run your first network simulation
3. **[API Reference →](../api/overview.md)** - Learn about the gRPC API

---

**Note**: Installation requires root access to your OpenWrt router. Make sure you have SSH access and sufficient permissions before proceeding.