# Integration Tests

This directory contains integration tests for NetTestLab that test the complete workflow from building the OpenWrt package to deploying and testing on a real router.

## Prerequisites

### 1. OpenWrt SDK

Download and set up the OpenWrt SDK for your target architecture:

```bash
# Example for x86_64
wget https://downloads.openwrt.org/releases/23.05.0/targets/x86/64/openwrt-sdk-23.05.0-x86-64_gcc-12.3.0_musl.Linux-x86_64.tar.xz
tar xf openwrt-sdk-*.tar.xz
cd openwrt-sdk-*

# Install feeds
./scripts/feeds update -a
./scripts/feeds install -a
```

### 2. Test Router

You need a physical OpenWrt router or VM for testing:
- OpenWrt 23.05 or later
- SSH access
- Network connectivity
- At least 50MB free space

### 3. Dependencies

Install required tools:

```bash
# On Ubuntu/Debian
sudo apt-get install sshpass

# On macOS
brew install sshpass
```

## Configuration

Set environment variables before running tests:

```bash
# Required
export OPENWRT_SDK_PATH="/path/to/openwrt-sdk"
export NETTESTLAB_ROUTER_IP="192.168.1.1"

# Optional
export NETTESTLAB_ROUTER_USER="root"               # default: root
export NETTESTLAB_ROUTER_PASSWORD="your-password"  # if using password auth
export SSH_KEY_PATH="/path/to/ssh/key"             # if using key auth
```

### Example Configuration Script

Create a `test-config.sh` file:

```bash
#!/bin/bash
# OpenWrt SDK path
export OPENWRT_SDK_PATH="$HOME/openwrt-sdk-23.05.0-x86-64_gcc-12.3.0_musl.Linux-x86_64"

# Router configuration
export NETTESTLAB_ROUTER_IP="192.168.1.10"
export NETTESTLAB_ROUTER_USER="root"
export NETTESTLAB_ROUTER_PASSWORD="admin123"

# Alternative: SSH key authentication
# export SSH_KEY_PATH="$HOME/.ssh/id_rsa"

echo "Integration test environment configured"
echo "Router: $NETTESTLAB_ROUTER_IP"
echo "SDK: $OPENWRT_SDK_PATH"
```

## Running Tests

### Full Integration Test

```bash
# Source configuration
source test-config.sh

# Run all integration tests
cd /path/to/NetTestLab
go test -v ./tests/integration/
```

### Individual Test Steps

```bash
# Test individual components
go test -v ./tests/integration/ -run TestOpenWrtIntegration/BuildPackage
go test -v ./tests/integration/ -run TestOpenWrtIntegration/DeployPackage
go test -v ./tests/integration/ -run TestOpenWrtIntegration/GRPCConnectivity
```

## Test Scenarios

### 1. Package Build Test
- Copies NetTestLab package to OpenWrt SDK
- Compiles package using OpenWrt build system
- Verifies package is created successfully

### 2. Package Deployment Test
- Copies package to router via SCP
- Installs package using opkg
- Verifies installation success

### 3. Service Status Test
- Checks if NetTestLab service is running
- Starts service if not running
- Verifies service health

### 4. gRPC Connectivity Test
- Connects to NetTestLab gRPC API
- Tests basic API responsiveness
- Verifies system health status

### 5. Profile Management Test
- Lists available profiles
- Verifies built-in profiles exist
- Creates, retrieves, and deletes custom profiles
- Tests profile validation

### 6. Network Control Test
- Applies custom network conditions
- Verifies conditions are active
- Tests profile application
- Resets network conditions

### 7. System Monitoring Test
- Retrieves system metrics
- Gets interface information
- Tests WiFi auto-discovery

## Test Output

Successful test run example:

```
=== RUN   TestOpenWrtIntegration
=== RUN   TestOpenWrtIntegration/BuildPackage
    Building OpenWrt package...
    Compiling package...
    Package built successfully: /path/to/nettestlab_1.0.0-1_x86_64.ipk
--- PASS: TestOpenWrtIntegration/BuildPackage (45.32s)
=== RUN   TestOpenWrtIntegration/DeployPackage
    Deploying package to router...
    Removing existing package...
    Installing package...
    Package installed successfully
--- PASS: TestOpenWrtIntegration/DeployPackage (8.41s)
=== RUN   TestOpenWrtIntegration/ServiceStatus
    Checking service status...
    Service is running successfully
--- PASS: TestOpenWrtIntegration/ServiceStatus (2.15s)
=== RUN   TestOpenWrtIntegration/GRPCConnectivity
    Testing gRPC connectivity...
    gRPC connectivity test passed
--- PASS: TestOpenWrtIntegration/GRPCConnectivity (1.23s)
=== RUN   TestOpenWrtIntegration/ProfileManagement
    Testing profile management...
    Listing profiles...
    Found 6 profiles
    Creating custom profile...
    Profile management test passed
--- PASS: TestOpenWrtIntegration/ProfileManagement (3.45s)
=== RUN   TestOpenWrtIntegration/NetworkControl
    Testing network control...
    Getting system status...
    Using interface wlan0 for testing
    Applying custom network conditions...
    Network conditions applied successfully
    Applying 3G profile...
    3G profile applied successfully
    Resetting network conditions...
    Network control test passed
--- PASS: TestOpenWrtIntegration/NetworkControl (12.67s)
=== RUN   TestOpenWrtIntegration/SystemMonitoring
    Testing system monitoring...
    Getting system metrics...
    System metrics: CPU 15.2%, Memory 42.1%, Uptime 3600s
    Getting interface information...
    Interface wlan0: up
    System monitoring test passed
--- PASS: TestOpenWrtIntegration/SystemMonitoring (2.34s)
--- PASS: TestOpenWrtIntegration (75.57s)
PASS
ok      github.com/Eitol/NetTestLab/tests/integration   75.574s
```

## Troubleshooting

### Build Issues

**Package compilation fails:**
```bash
# Check SDK setup
ls -la $OPENWRT_SDK_PATH/
./scripts/feeds list -i | grep golang

# Verify dependencies
make package/nettestlab/download
make package/nettestlab/prepare
```

### Connection Issues

**SSH connection fails:**
```bash
# Test SSH connectivity
ssh $NETTESTLAB_ROUTER_USER@$NETTESTLAB_ROUTER_IP "echo 'Connection OK'"

# Check SSH key permissions
chmod 600 $SSH_KEY_PATH
```

**gRPC connection fails:**
```bash
# Check service status on router
ssh root@$NETTESTLAB_ROUTER_IP "/etc/init.d/nettestlab status"

# Check logs
ssh root@$NETTESTLAB_ROUTER_IP "logread | grep nettestlab"

# Check firewall
ssh root@$NETTESTLAB_ROUTER_IP "iptables -L | grep 8080"
```

### Service Issues

**Service won't start:**
```bash
# Check dependencies
ssh root@router "opkg list-installed | grep -E 'tc|kmod-sched'"

# Check configuration
ssh root@router "uci show nettestlab"

# Manual start with debugging
ssh root@router "nettestlab -port 8080 -log-level debug"
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: OpenWrt Integration Tests

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  integration-test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Download OpenWrt SDK
      run: |
        wget https://downloads.openwrt.org/releases/23.05.0/targets/x86/64/openwrt-sdk-23.05.0-x86-64_gcc-12.3.0_musl.Linux-x86_64.tar.xz
        tar xf openwrt-sdk-*.tar.xz
        
    - name: Setup SDK
      run: |
        cd openwrt-sdk-*
        ./scripts/feeds update -a
        ./scripts/feeds install -a
        
    - name: Run Integration Tests
      env:
        OPENWRT_SDK_PATH: ${{ github.workspace }}/openwrt-sdk-23.05.0-x86-64_gcc-12.3.0_musl.Linux-x86_64
        NETTESTLAB_ROUTER_IP: ${{ secrets.TEST_ROUTER_IP }}
        NETTESTLAB_ROUTER_PASSWORD: ${{ secrets.TEST_ROUTER_PASSWORD }}
      run: |
        go test -v ./tests/integration/
```

## Security Notes

- Never commit router passwords or SSH keys
- Use environment variables or secrets management
- Consider using dedicated test networks
- Router firewall rules may need adjustment

## Hardware Recommendations

### Test Router Requirements
- **RAM**: Minimum 128MB (256MB+ recommended)
- **Storage**: 50MB+ free space
- **CPU**: Any OpenWrt-supported architecture
- **Network**: Ethernet + WiFi for complete testing

### Recommended Test Routers
- **TP-Link Archer C7 v2/v5**: Good for testing, affordable
- **Netgear R7800**: High performance, good for stress testing
- **GL.iNet GL-AX1800**: Modern hardware, good OpenWrt support
- **x86_64 VM**: Best for CI/CD, easy to automate

---

**Next Steps**: Run the integration tests to validate your NetTestLab deployment!