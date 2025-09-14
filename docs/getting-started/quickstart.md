# Quick Start

Get up and running with NetTestLab in just 5 minutes! This tutorial will walk you through your first network simulation.

## Before You Begin

Make sure you have:

- ✅ NetTestLab [installed](installation.md) on your OpenWrt router
- ✅ Network access to your router
- ✅ A client device connected to the router's network

## Step 1: Verify Installation

First, let's make sure NetTestLab is running:

```bash
# SSH into your router
ssh root@your-router-ip

# Check service status
/etc/init.d/nettestlab status

# Should show: nettestlab is running
```

If the service isn't running, start it:

```bash
/etc/init.d/nettestlab start
```

## Step 2: Test Basic Connectivity

Test that the gRPC server is accessible:

=== "Using grpcurl"

    ```bash
    # List available services
    grpcurl -plaintext your-router-ip:8080 list
    
    # Should show:
    # nettestlab.v1.MonitoringService
    # nettestlab.v1.NetworkControlService
    # nettestlab.v1.ProfileService
    ```

=== "Using curl (if HTTP gateway enabled)"

    ```bash
    # Check health endpoint
    curl http://your-router-ip:8080/health
    
    # Should return status information
    ```

## Step 3: Your First Network Simulation

Let's simulate a 3G mobile network connection:

### Apply 3G Conditions

```bash
grpcurl -plaintext -d '{
  "interface": "wifi",
  "conditions": {
    "latency": {"delay_ms": 150, "enabled": true},
    "packet_loss": {"percentage": 0.5, "enabled": true},
    "bandwidth": {"download_bps": 1600000, "upload_bps": 384000, "enabled": true}
  }
}' your-router-ip:8080 nettestlab.v1.NetworkControlService/ApplyNetworkConditions
```

### Test the Conditions

From a device connected to your router's WiFi:

```bash
# Test latency (should show ~150ms additional delay)
ping google.com

# Test bandwidth (should be limited to ~1.6Mbps down, 384Kbps up)
speedtest-cli
```

### Reset Conditions

When you're done testing, reset the network to normal:

```bash
grpcurl -plaintext -d '{
  "interface": "wifi"
}' your-router-ip:8080 nettestlab.v1.NetworkControlService/ResetNetworkConditions
```

## Step 4: Using Built-in Profiles

NetTestLab comes with pre-configured profiles for common scenarios:

### List Available Profiles

```bash
grpcurl -plaintext your-router-ip:8080 nettestlab.v1.ProfileService/ListProfiles
```

### Apply a Profile

Apply the 2G profile (very slow connection):

```bash
grpcurl -plaintext -d '{
  "interface": "wifi",
  "profile_name": "2g"
}' your-router-ip:8080 nettestlab.v1.ProfileService/ApplyProfile
```

Test the slow connection and then reset:

```bash
# Reset when done
grpcurl -plaintext -d '{
  "interface": "wifi"
}' your-router-ip:8080 nettestlab.v1.NetworkControlService/ResetNetworkConditions
```

## Step 5: Monitor System Status

Check system health and current conditions:

```bash
# Get system status
grpcurl -plaintext your-router-ip:8080 nettestlab.v1.NetworkControlService/GetSystemStatus

# Get current conditions for an interface
grpcurl -plaintext -d '{
  "interface": "wifi"
}' your-router-ip:8080 nettestlab.v1.NetworkControlService/GetNetworkConditions
```

## What's Next?

Congratulations! You've successfully:

- ✅ Verified NetTestLab installation
- ✅ Applied your first network conditions
- ✅ Used a built-in profile
- ✅ Reset network conditions
- ✅ Monitored system status

### Explore More

- **[Client Libraries →](../clients/overview.md)** - Use NetTestLab from your applications
- **[Custom Profiles →](../examples/custom-profiles.md)** - Create your own network profiles
- **[API Reference →](../api/overview.md)** - Learn about all available operations
- **[Mobile Testing →](../examples/mobile-testing.md)** - Test mobile applications

### Common Commands Reference

| Action | Command |
|--------|---------|
| Apply 3G Profile | `grpcurl -plaintext -d '{"interface": "wifi", "profile_name": "3g"}' router:8080 nettestlab.v1.ProfileService/ApplyProfile` |
| Apply 4G Profile | `grpcurl -plaintext -d '{"interface": "wifi", "profile_name": "4g"}' router:8080 nettestlab.v1.ProfileService/ApplyProfile` |
| Reset Conditions | `grpcurl -plaintext -d '{"interface": "wifi"}' router:8080 nettestlab.v1.NetworkControlService/ResetNetworkConditions` |
| List Profiles | `grpcurl -plaintext router:8080 nettestlab.v1.ProfileService/ListProfiles` |
| System Status | `grpcurl -plaintext router:8080 nettestlab.v1.NetworkControlService/GetSystemStatus` |

### Troubleshooting

#### Can't Connect to gRPC Server

1. Check if service is running: `/etc/init.d/nettestlab status`
2. Check firewall rules allow port 8080
3. Verify correct IP address and port

#### Network Conditions Not Applied

1. Check interface name exists: `ip link show`
2. Verify traffic control is working: `tc qdisc show`
3. Check logs: `logread | grep nettestlab`

#### Commands Not Working

1. Install grpcurl: `go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest`
2. Or use a client library instead of command line
3. Check service logs for errors

---

Need help? Check the [troubleshooting guide](../openwrt/troubleshooting.md) or [open an issue](https://github.com/Eitol/NetTestLab/issues).