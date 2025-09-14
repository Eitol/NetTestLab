# Scripts Directory

This directory contains utility scripts for building and deploying NetTestLab.

## Available Scripts

### `build-openwrt-package.sh`

Builds the OpenWrt IPK package for router deployment. This is the main build script that handles the complete package creation process.

**Usage:**

```bash
# Build package
./scripts/build-openwrt-package.sh

# Clean build directory
./scripts/build-openwrt-package.sh clean
```

**Requirements:**

- Go 1.21+
- buf CLI
- binutils (for `ar` command)

**Features:**
- Cross-compiles Go binary for ARM64
- Generates Protocol Buffer files
- Creates proper OpenWrt package structure
- Includes control files and scripts
- Validates final IPK package

## Future Scripts

The following scripts may be added in future versions:

### `test-wifi-discovery.sh`
Tests WiFi auto-discovery functionality on target router.

### `deploy-to-router.sh`
Automated deployment script for uploading and installing package on router.

### `benchmark.sh`
Performance benchmarking script for different network conditions.

## Development Workflow

1. **Build package:**

   ```bash
   ./scripts/build-openwrt-package.sh
   ```

2. **Deploy to router:**

   ```bash
   scp nettestlab_1.0.0_aarch64.ipk root@192.168.1.4:/tmp/
   ssh root@192.168.1.4 "opkg install /tmp/nettestlab_1.0.0_aarch64.ipk"
   ```

3. **Start service:**

   ```bash
   ssh root@192.168.1.4 "/etc/init.d/nettestlab start"
   ```

4. **Test functionality:**

   ```bash
   go run cmd/wifi-test/main.go -server 192.168.1.4:8080
   ```