# API Overview

NetTestLab exposes a gRPC API with three main services for network traffic control and monitoring. All services are defined using Protocol Buffers for cross-language compatibility.

## Service Architecture

```mermaid
graph TB
    Client[Client Applications] --> Gateway[gRPC Gateway]
    Gateway --> NS[NetworkControlService]
    Gateway --> PS[ProfileService] 
    Gateway --> MS[MonitoringService]
    
    NS --> TC[Traffic Control]
    PS --> PM[Profile Manager]
    MS --> SysInfo[System Info]
    
    TC --> Kernel[Linux Kernel tc/netem]
    PM --> FS[File System]
    SysInfo --> Proc[/proc filesystem]
```

## Core Services

### NetworkControlService

Controls network traffic conditions in real-time.

**Key Operations:**
- `ApplyNetworkConditions` - Apply latency, bandwidth, and packet loss
- `ResetNetworkConditions` - Remove all applied conditions
- `GetNetworkConditions` - Query current conditions for an interface
- `GetSystemStatus` - Get overall system health and capabilities

### ProfileService

Manages reusable network condition profiles.

**Key Operations:**
- `CreateProfile` - Create a new profile
- `ListProfiles` - Get all available profiles
- `GetProfile` - Retrieve a specific profile
- `UpdateProfile` - Modify an existing profile
- `DeleteProfile` - Remove a profile
- `ApplyProfile` - Apply a profile to an interface

### MonitoringService

Provides system monitoring and metrics collection.

**Key Operations:**
- `GetSystemMetrics` - CPU, memory, network interface stats
- `StreamMetrics` - Real-time metrics streaming
- `GetInterfaceInfo` - Detailed interface information

## Data Models

### NetworkConditions

```protobuf
message NetworkConditions {
  LatencyCondition latency = 1;
  BandwidthCondition bandwidth = 2;
  PacketLossCondition packet_loss = 3;
}
```

### Profile

```protobuf
message Profile {
  string name = 1;
  string description = 2;
  NetworkConditions conditions = 3;
  bool built_in = 4;
  google.protobuf.Timestamp created_at = 5;
  google.protobuf.Timestamp updated_at = 6;
}
```

## Built-in Profiles

NetTestLab includes several pre-configured profiles:

| Profile | Download | Upload | Latency | Packet Loss | Description |
|---------|----------|--------|---------|-------------|-------------|
| **2g** | 64 Kbps | 32 Kbps | 300ms | 1% | Legacy 2G mobile |
| **3g** | 1.6 Mbps | 384 Kbps | 150ms | 0.5% | Standard 3G |
| **4g** | 50 Mbps | 20 Mbps | 50ms | 0.1% | 4G LTE |
| **wifi** | 100 Mbps | 100 Mbps | 5ms | 0% | Good WiFi |
| **satellite** | 25 Mbps | 3 Mbps | 600ms | 0.5% | Satellite internet |

## Error Codes

### Common gRPC Status Codes

| Code | Status | Description |
|------|--------|-------------|
| 0 | OK | Request successful |
| 3 | INVALID_ARGUMENT | Invalid parameters |
| 5 | NOT_FOUND | Resource not found |
| 13 | INTERNAL | System error |
| 14 | UNAVAILABLE | Service unavailable |

### Custom Error Details

NetTestLab provides detailed error information in the gRPC status details:

```json
{
  "error_code": "INTERFACE_NOT_FOUND",
  "message": "Network interface 'eth1' not found",
  "details": {
    "interface": "eth1",
    "available_interfaces": ["eth0", "wifi", "br-lan"]
  }
}
```

## Rate Limiting

- **Apply Operations**: Maximum 10 requests per minute per interface
- **Query Operations**: No limits
- **Profile Operations**: Maximum 100 requests per minute

## Examples

### Apply Custom Conditions

```bash
grpcurl -plaintext -d '{
  "interface": "wifi",
  "conditions": {
    "latency": {"delay_ms": 100, "enabled": true},
    "bandwidth": {"download_bps": 10000000, "upload_bps": 5000000, "enabled": true},
    "packet_loss": {"percentage": 0.1, "enabled": true}
  }
}' router:8080 nettestlab.v1.NetworkControlService/ApplyNetworkConditions
```

### Create Custom Profile

```bash
grpcurl -plaintext -d '{
  "profile": {
    "name": "slow_wifi",
    "description": "Simulates congested WiFi",
    "conditions": {
      "latency": {"delay_ms": 50, "enabled": true},
      "bandwidth": {"download_bps": 5000000, "upload_bps": 1000000, "enabled": true}
    }
  }
}' router:8080 nettestlab.v1.ProfileService/CreateProfile
```

## Protocol Buffer Definitions

The complete API definitions are available in the `proto/` directory:

- **[network.proto](https://github.com/Eitol/NetTestLab/blob/main/proto/nettestlab/v1/network.proto)** - Network control operations
- **[profiles.proto](https://github.com/Eitol/NetTestLab/blob/main/proto/nettestlab/v1/profiles.proto)** - Profile management
- **[monitoring.proto](https://github.com/Eitol/NetTestLab/blob/main/proto/nettestlab/v1/monitoring.proto)** - System monitoring

## Interactive API Explorer

For development and testing, you can use tools like:

- **grpcurl** - Command-line gRPC client
- **BloomRPC** - GUI gRPC client
- **Postman** - With gRPC support
- **buf studio** - Web-based Protocol Buffer explorer

---

**Next:** [Network Control API →](network-control.md) | [Profile Management API →](profile-management.md) | [Monitoring API →](monitoring.md)