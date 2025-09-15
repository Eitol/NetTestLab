# Integration Tests Makefile for NetTestLab

.PHONY: help test-integration test-build test-deploy test-quick clean-test setup-sdk

# Default target
help:
	@echo "NetTestLab Integration Tests"
	@echo ""
	@echo "Available targets:"
	@echo "  help              Show this help message"
	@echo "  test-integration  Run full integration test suite"
	@echo "  test-build        Test package building only"
	@echo "  test-deploy       Test deployment only (requires built package)"
	@echo "  test-quick        Run tests without building (requires deployed package)"
	@echo "  setup-sdk         Download and setup OpenWrt SDK"
	@echo "  clean-test        Clean test artifacts"
	@echo ""
	@echo "Required environment variables:"
	@echo "  NETTESTLAB_ROUTER_IP        Router IP address"
	@echo "  OPENWRT_SDK_PATH            Path to OpenWrt SDK"
	@echo ""
	@echo "Optional environment variables:"
	@echo "  NETTESTLAB_ROUTER_USER      SSH username (default: root)"
	@echo "  NETTESTLAB_ROUTER_PASSWORD  SSH password"
	@echo "  SSH_KEY_PATH                SSH private key path"
	@echo ""
	@echo "Examples:"
	@echo "  make test-integration NETTESTLAB_ROUTER_IP=192.168.1.10 OPENWRT_SDK_PATH=~/openwrt-sdk"
	@echo "  make test-quick NETTESTLAB_ROUTER_IP=192.168.1.10"

# Check required environment variables
check-env:
	@if [ -z "$(NETTESTLAB_ROUTER_IP)" ]; then \
		echo "Error: NETTESTLAB_ROUTER_IP is required"; \
		exit 1; \
	fi

# Check SDK path for build tests
check-sdk:
	@if [ -z "$(OPENWRT_SDK_PATH)" ]; then \
		echo "Error: OPENWRT_SDK_PATH is required for build tests"; \
		exit 1; \
	fi
	@if [ ! -d "$(OPENWRT_SDK_PATH)" ]; then \
		echo "Error: OpenWrt SDK path does not exist: $(OPENWRT_SDK_PATH)"; \
		exit 1; \
	fi

# Run full integration test suite
test-integration: check-env check-sdk
	@echo "Running full integration test suite..."
	./scripts/run-integration-tests.sh \
		--router-ip $(NETTESTLAB_ROUTER_IP) \
		--sdk-path $(OPENWRT_SDK_PATH) \
		$(if $(NETTESTLAB_ROUTER_USER),--user $(NETTESTLAB_ROUTER_USER)) \
		$(if $(NETTESTLAB_ROUTER_PASSWORD),--password $(NETTESTLAB_ROUTER_PASSWORD)) \
		$(if $(SSH_KEY_PATH),--key-path $(SSH_KEY_PATH)) \
		--verbose

# Test package building only
test-build: check-sdk
	@echo "Testing package build..."
	go test -v ./tests/integration/ -run TestOpenWrtIntegration/BuildPackage

# Test deployment only
test-deploy: check-env
	@echo "Testing package deployment..."
	./scripts/run-integration-tests.sh \
		--router-ip $(NETTESTLAB_ROUTER_IP) \
		--sdk-path $(OPENWRT_SDK_PATH) \
		$(if $(NETTESTLAB_ROUTER_USER),--user $(NETTESTLAB_ROUTER_USER)) \
		$(if $(NETTESTLAB_ROUTER_PASSWORD),--password $(NETTESTLAB_ROUTER_PASSWORD)) \
		$(if $(SSH_KEY_PATH),--key-path $(SSH_KEY_PATH)) \
		--skip-build \
		--verbose

# Run tests without building or deploying (assumes package is already deployed)
test-quick: check-env
	@echo "Running quick tests (skipping build and deploy)..."
	./scripts/run-integration-tests.sh \
		--router-ip $(NETTESTLAB_ROUTER_IP) \
		$(if $(NETTESTLAB_ROUTER_USER),--user $(NETTESTLAB_ROUTER_USER)) \
		$(if $(NETTESTLAB_ROUTER_PASSWORD),--password $(NETTESTLAB_ROUTER_PASSWORD)) \
		$(if $(SSH_KEY_PATH),--key-path $(SSH_KEY_PATH)) \
		--skip-build \
		--skip-deploy \
		--verbose

# Setup OpenWrt SDK (x86_64 example)
setup-sdk:
	@echo "Setting up OpenWrt SDK..."
	@if [ -z "$(SDK_VERSION)" ]; then \
		SDK_VERSION="23.05.0"; \
	fi; \
	if [ -z "$(SDK_ARCH)" ]; then \
		SDK_ARCH="x86-64"; \
	fi; \
	SDK_NAME="openwrt-sdk-$$SDK_VERSION-$$SDK_ARCH_gcc-12.3.0_musl.Linux-x86_64"; \
	if [ ! -d "$$SDK_NAME" ]; then \
		echo "Downloading OpenWrt SDK..."; \
		wget "https://downloads.openwrt.org/releases/$$SDK_VERSION/targets/x86/64/$$SDK_NAME.tar.xz"; \
		tar xf "$$SDK_NAME.tar.xz"; \
		rm "$$SDK_NAME.tar.xz"; \
	fi; \
	cd "$$SDK_NAME"; \
	echo "Setting up feeds..."; \
	./scripts/feeds update -a; \
	./scripts/feeds install -a; \
	echo "SDK setup complete: $$(pwd)/$$SDK_NAME"; \
	echo "Set OPENWRT_SDK_PATH=$$(pwd)/$$SDK_NAME"

# Clean test artifacts
clean-test:
	@echo "Cleaning test artifacts..."
	rm -rf /tmp/nettestlab-test*
	go clean -testcache

# Test individual components
test-grpc: check-env
	@echo "Testing gRPC connectivity..."
	./scripts/run-integration-tests.sh \
		--router-ip $(NETTESTLAB_ROUTER_IP) \
		$(if $(NETTESTLAB_ROUTER_USER),--user $(NETTESTLAB_ROUTER_USER)) \
		$(if $(NETTESTLAB_ROUTER_PASSWORD),--password $(NETTESTLAB_ROUTER_PASSWORD)) \
		$(if $(SSH_KEY_PATH),--key-path $(SSH_KEY_PATH)) \
		--test GRPCConnectivity \
		--skip-build \
		--skip-deploy

test-profiles: check-env
	@echo "Testing profile management..."
	./scripts/run-integration-tests.sh \
		--router-ip $(NETTESTLAB_ROUTER_IP) \
		$(if $(NETTESTLAB_ROUTER_USER),--user $(NETTESTLAB_ROUTER_USER)) \
		$(if $(NETTESTLAB_ROUTER_PASSWORD),--password $(NETTESTLAB_ROUTER_PASSWORD)) \
		$(if $(SSH_KEY_PATH),--key-path $(SSH_KEY_PATH)) \
		--test ProfileManagement \
		--skip-build \
		--skip-deploy

test-network: check-env
	@echo "Testing network control..."
	./scripts/run-integration-tests.sh \
		--router-ip $(NETTESTLAB_ROUTER_IP) \
		$(if $(NETTESTLAB_ROUTER_USER),--user $(NETTESTLAB_ROUTER_USER)) \
		$(if $(NETTESTLAB_ROUTER_PASSWORD),--password $(NETTESTLAB_ROUTER_PASSWORD)) \
		$(if $(SSH_KEY_PATH),--key-path $(SSH_KEY_PATH)) \
		--test NetworkControl \
		--skip-build \
		--skip-deploy

test-monitoring: check-env
	@echo "Testing system monitoring..."
	./scripts/run-integration-tests.sh \
		--router-ip $(NETTESTLAB_ROUTER_IP) \
		$(if $(NETTESTLAB_ROUTER_USER),--user $(NETTESTLAB_ROUTER_USER)) \
		$(if $(NETTESTLAB_ROUTER_PASSWORD),--password $(NETTESTLAB_ROUTER_PASSWORD)) \
		$(if $(SSH_KEY_PATH),--key-path $(SSH_KEY_PATH)) \
		--test SystemMonitoring \
		--skip-build \
		--skip-deploy