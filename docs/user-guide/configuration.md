# Configuration

NetTestLab is designed to work out-of-the-box with minimal configuration. However, you can customize various aspects of its behavior to fit your specific needs.

## Configuration Files

### Main Configuration

NetTestLab uses UCI (Unified Configuration Interface) on OpenWrt for system configuration:

```bash
# Main configuration file
/etc/config/nettestlab
```

### Default Configuration

```bash
# View current configuration
uci show nettestlab

# Output:
nettestlab.server=server
nettestlab.server.enabled='1'
nettestlab.server.port='8080'
nettestlab.server.host='0.0.0.0'
nettestlab.server.log_level='info'
nettestlab.server.profiles_dir='/etc/nettestlab/profiles'
```

## Server Configuration

### Basic Settings

```bash
# Change server port
uci set nettestlab.server.port='9090'

# Bind to specific interface only
uci set nettestlab.server.host='192.168.1.1'

# Enable debug logging
uci set nettestlab.server.log_level='debug'

# Commit changes
uci commit nettestlab

# Restart service
/etc/init.d/nettestlab restart
```

### Available Log Levels

| Level | Description | Use Case |
|-------|-------------|----------|
| `error` | Errors only | Production |
| `warn` | Warnings and errors | Production |
| `info` | General information | Default |
| `debug` | Detailed debugging | Development |
| `trace` | Very verbose | Troubleshooting |

## Profile Configuration

### Profile Directory

Profiles are stored as JSON files in the profiles directory:

```bash
# Default location
/etc/nettestlab/profiles/

# Custom location
uci set nettestlab.server.profiles_dir='/custom/path'
uci commit nettestlab
```

### Creating Custom Profiles

Create a new profile file:

```bash
cat > /etc/nettestlab/profiles/office_wifi.json << 'EOF'
{
  "name": "office_wifi",
  "description": "Typical office WiFi with some congestion",
  "conditions": {
    "latency": {
      "delay_ms": 15,
      "enabled": true
    },
    "bandwidth": {
      "download_bps": 75000000,
      "upload_bps": 25000000,
      "enabled": true
    },
    "packet_loss": {
      "percentage": 0.2,
      "enabled": true
    }
  },
  "built_in": false
}
EOF

# Restart service to load new profile
/etc/init.d/nettestlab restart
```

### Profile Validation

Profiles are validated on startup. Check logs for validation errors:

```bash
logread | grep nettestlab | grep -i error
```

## Network Interface Configuration

### Supported Interfaces

NetTestLab can control traffic on any network interface:

```bash
# List available interfaces
ip link show

# Common interfaces on OpenWrt:
# - eth0: Ethernet
# - wifi: WiFi
# - br-lan: Bridge (LAN)
# - wwan: Mobile/cellular
```

### Interface-Specific Settings

You can configure per-interface defaults:

```bash
# Set default interface for profiles
uci set nettestlab.server.default_interface='wifi'

# Set maximum bandwidth limits per interface
uci set nettestlab.interface_limits=interface_limits
uci set nettestlab.interface_limits.wifi_max_download='200000000'
uci set nettestlab.interface_limits.wifi_max_upload='100000000'

uci commit nettestlab
```

## Security Configuration

### Network Access Control

Restrict access to NetTestLab server:

```bash
# Firewall rule to allow only specific networks
iptables -A INPUT -p tcp --dport 8080 -s 192.168.1.0/24 -j ACCEPT
iptables -A INPUT -p tcp --dport 8080 -j DROP

# Make permanent
echo "iptables -A INPUT -p tcp --dport 8080 -s 192.168.1.0/24 -j ACCEPT" >> /etc/firewall.user
echo "iptables -A INPUT -p tcp --dport 8080 -j DROP" >> /etc/firewall.user
```

### Disable Built-in Profiles

Prevent modification of built-in profiles:

```bash
# Read-only built-in profiles
uci set nettestlab.server.readonly_builtin='1'
uci commit nettestlab
```

## Performance Configuration

### System Resource Limits

Configure resource usage limits:

```bash
# Maximum concurrent connections
uci set nettestlab.server.max_connections='100'

# Request timeout (seconds)
uci set nettestlab.server.request_timeout='30'

# Profile cache size
uci set nettestlab.server.profile_cache_size='50'

uci commit nettestlab
```

### Traffic Control Optimization

Fine-tune traffic control behavior:

```bash
# Quantum size for bandwidth limiting (bytes)
uci set nettestlab.tc.quantum='1500'

# Buffer size for traffic shaping
uci set nettestlab.tc.buffer_size='32768'

# Enable hardware offloading if supported
uci set nettestlab.tc.hw_offload='1'

uci commit nettestlab
```

## Monitoring Configuration

### Metrics Collection

Configure system monitoring:

```bash
# Enable metrics collection
uci set nettestlab.monitoring.enabled='1'

# Metrics collection interval (seconds)
uci set nettestlab.monitoring.interval='5'

# Metrics retention period (hours)
uci set nettestlab.monitoring.retention='24'

uci commit nettestlab
```

### Log Rotation

Configure log rotation to manage disk space:

```bash
# Maximum log file size (MB)
uci set nettestlab.logging.max_size='10'

# Number of log files to keep
uci set nettestlab.logging.max_files='5'

# Enable compression
uci set nettestlab.logging.compress='1'

uci commit nettestlab
```

## Environment Variables

Some settings can be overridden with environment variables:

```bash
# Set in init script or systemd service
export NETTESTLAB_PORT=8080
export NETTESTLAB_LOG_LEVEL=debug
export NETTESTLAB_PROFILES_DIR=/custom/profiles
```

## Configuration Management

### Backup Configuration

```bash
# Backup UCI configuration
uci export nettestlab > /tmp/nettestlab-config.backup

# Backup profiles
tar -czf /tmp/nettestlab-profiles.tar.gz /etc/nettestlab/profiles/
```

### Restore Configuration

```bash
# Restore UCI configuration
uci import nettestlab < /tmp/nettestlab-config.backup
uci commit nettestlab

# Restore profiles
tar -xzf /tmp/nettestlab-profiles.tar.gz -C /
```

### Factory Reset

```bash
# Reset to default configuration
uci delete nettestlab
uci commit nettestlab

# Restart with defaults
/etc/init.d/nettestlab restart
```

## Validation and Testing

### Configuration Validation

Test configuration before applying:

```bash
# Validate UCI configuration
nettestlab --config-check

# Test profile loading
nettestlab --validate-profiles
```

### Service Health Check

```bash
# Check service status
/etc/init.d/nettestlab status

# Test gRPC endpoint
grpcurl -plaintext localhost:8080 list

# Check system logs
logread | grep nettestlab | tail -20
```

## Troubleshooting

### Common Configuration Issues

**Service won't start:**
- Check UCI configuration syntax
- Verify file permissions
- Check available disk space

**Profiles not loading:**
- Validate JSON syntax
- Check file permissions
- Verify profiles directory path

**Can't connect to service:**
- Check firewall rules
- Verify host/port configuration
- Check interface binding

### Debug Configuration

Enable maximum debugging:

```bash
uci set nettestlab.server.log_level='trace'
uci set nettestlab.server.debug_traffic='1'
uci set nettestlab.server.debug_profiles='1'
uci commit nettestlab
/etc/init.d/nettestlab restart

# Watch logs in real-time
logread -f | grep nettestlab
```

---

**Next:** [Troubleshooting →](../openwrt/troubleshooting.md) | [API Reference →](../api/overview.md)