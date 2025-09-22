# Scripts

Utility scripts for NetTestLab deployment and testing.This directory contains utility scripts for building, deploying, and testing NetTestLab.

## Core Scripts

### `deploy-to-device.sh`

Complete automated deployment to OpenWRT router. Builds package, installs on device, starts service, and runs tests.Complete automated deployment script that handles the entire workflow:

- Detects router architecture automatically  

- Stops existing processes on router

- Builds OpenWRT IPK package using Docker for cross-compilation. (Using build-openwrt-package.sh)

- Installs/updates the package on router

- Starts the service

### `run-integration-test.sh`

Runs comprehensive integration tests against deployed NetTestLab instance.

**Usage:**

```bash

## Usage# Deploy to default router (192.168.1.4)

./scripts/deploy-auto.sh

```bash

# Deploy everything to router# Deploy to custom router

./scripts/deploy-to-device.shROUTER_IP=192.168.1.4 ./scripts/deploy-auto.sh

```

### `build-openwrt-package.sh`

Builds OpenWRT IPK package using Docker for cross-compilation.
