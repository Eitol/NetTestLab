# gRPC API Definition Summary

## Overview

The NetTestLab gRPC API has been successfully defined using Protocol Buffers with comprehensive service interfaces for network simulation and control.

## Services Defined

### 1. NetworkControlService (`network.proto`)
**Core network condition management**

- `ApplyNetworkConditions` - Apply specific network conditions to interfaces
- `ResetNetworkConditions` - Remove all conditions from interfaces  
- `GetNetworkConditions` - Query current conditions
- `GetSystemStatus` - Get system and interface status

**Key Features:**
- Latency control (0-5000ms)
- Packet loss simulation (0-100% with patterns)
- Bandwidth limiting (upload/download)
- Jitter simulation with distribution types
- Packet corruption simulation
- Traffic direction control (ingress/egress/both)

### 2. ProfileService (`profiles.proto`)
**Network profile management**

- `ListProfiles` - Get all available profiles
- `GetProfile` - Retrieve specific profile details
- `CreateProfile` / `UpdateProfile` / `DeleteProfile` - Profile CRUD operations
- `ApplyProfile` - Apply predefined profiles to interfaces

**Profile Types:**
- Mobile (2G, 3G, 4G, 5G)
- WiFi networks
- Satellite connections
- Custom user-defined
- Testing-specific profiles

### 3. MonitoringService (`monitoring.proto`)
**System monitoring and observability**

- `GetHealth` - Health check endpoint
- `GetMetrics` - System and network metrics
- `GetInterfaceStats` - Interface-specific statistics
- `StreamMetrics` - Real-time metrics streaming
- `GetHistoricalMetrics` - Time-series historical data

**Monitoring Features:**
- System metrics (CPU, memory, disk, load)
- Network interface statistics (traffic, errors, drops)
- NetTestLab-specific metrics (active conditions, API performance)
- Historical data with configurable resolution

## Data Models

### Network Conditions
```protobuf
message NetworkConditions {
  LatencyConfig latency = 1;
  PacketLossConfig packet_loss = 2;
  BandwidthConfig bandwidth = 3;
  JitterConfig jitter = 4;
  CorruptionConfig corruption = 5;
}
```

### Network Profiles
```protobuf
message NetworkProfile {
  string name = 1;
  string display_name = 2;
  string description = 3;
  NetworkConditions conditions = 4;
  ProfileType type = 5;
  bool built_in = 6;
  // ... metadata fields
}
```

### System Metrics
```protobuf
message SystemMetrics {
  float cpu_usage = 1;
  float memory_usage = 2;
  uint64 total_memory = 3;
  // ... additional metrics
}
```

## Code Generation Results

### ✅ Generated Components

**Go API (api/nettestlab/v1/)**
- `network.pb.go` / `network_grpc.pb.go` - Network control service
- `profiles.pb.go` / `profiles_grpc.pb.go` - Profile management service  
- `monitoring.pb.go` / `monitoring_grpc.pb.go` - Monitoring service

**Python Client (clients/python/src/)**
- `*_pb2.py` - Message definitions
- `*_pb2_grpc.py` - Service stubs

**Java Client (clients/java/src/main/java/)**
- `*.java` - Complete Java classes for messages and services

**JavaScript Client (clients/javascript/src/)**
- Protocol buffer and gRPC-Web generated files

## Configuration Files

### buf.yaml (Protocol Buffer management)
```yaml
version: v1
breaking:
  use: [FILE]
lint:
  use: [STANDARD, COMMENTS, FILE_LOWER_SNAKE_CASE]
deps:
  - buf.build/googleapis/googleapis
  - buf.build/bufbuild/protovalidate
```

### buf.gen.yaml (Code generation)
- Multi-language code generation for Go, Python, Java, JavaScript
- Proper output directories for each client library
- Language-specific options and configurations

### buf.work.yaml (Workspace configuration)
```yaml
version: v1
directories:
  - proto
```

## API Design Principles

### 1. **Consistency**
- Consistent naming patterns across all services
- Standard request/response message pairs
- Uniform error handling patterns

### 2. **Extensibility**
- Enum zero values with `_UNSPECIFIED` suffix
- Optional fields for backward compatibility
- Versioned package structure (`nettestlab.v1`)

### 3. **Type Safety**
- Strong typing for all parameters
- Enums for categorical values
- Proper field validation

### 4. **Real-world Usage**
- Practical network condition ranges
- Common network profile types
- Comprehensive monitoring capabilities

## Next Steps

The API definitions are complete and ready for:

1. **Go Backend Implementation** - Implement the gRPC service handlers
2. **Client Library Development** - Create language-specific wrapper libraries
3. **Testing & Validation** - Comprehensive API testing
4. **Documentation Generation** - Auto-generate API documentation

## Validation Status

- ✅ Protocol Buffer syntax validation (buf lint)
- ✅ Multi-language code generation (buf generate)  
- ✅ Import resolution and dependencies
- ✅ gRPC service definitions
- ✅ Message structure validation