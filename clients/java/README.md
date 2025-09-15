# NetTestLab Java Client

A simple and robust Java client for interacting with NetTestLab server. Control network conditions, monitor system metrics, and manage network profiles programmatically.

## Installation

### Maven
```xml
<dependency>
    <groupId>io.github.eitol</groupId>
    <artifactId>nettestlab-client</artifactId>
    <version>1.0.0</version>
</dependency>
```

### Gradle
```groovy
implementation 'io.github.eitol:nettestlab-client:1.0.0'
```

## Quick Start

```java
import io.github.eitol.nettestlab.NetTestLabClient;
import io.github.eitol.nettestlab.utils.NetworkConditionsBuilder;

// Create client
try (NetTestLabClient client = NetTestLabClient.builder()
        .host("localhost")
        .port(8080)
        .build()) {
    
    // Check server health
    var health = client.monitoring().getHealth();
    System.out.println("Server status: " + health.getStatus());
    
    // Apply 3G network conditions
    var conditions = NetworkConditionsBuilder.threeG();
    var response = client.networkControl().applyConditions("eth0", conditions);
    
    if (response.getSuccess()) {
        System.out.println("3G conditions applied successfully");
    }
    
    // Get system metrics
    var metrics = client.monitoring().getMetrics();
    System.out.printf("CPU: %.2f%%, Memory: %.2f%%\n", 
        metrics.getSystem().getCpuUsage(),
        metrics.getSystem().getMemoryUsage());
}
```

## Features

- **Network Control**: Apply latency, packet loss, bandwidth limitations
- **System Monitoring**: Get real-time CPU, memory, and network metrics  
- **Predefined Profiles**: 2G, 3G, 4G, 5G, WiFi, and satellite configurations
- **Custom Conditions**: Build your own network conditions
- **Thread-Safe**: Safe for concurrent use
- **Auto-Closeable**: Automatic resource management

## Network Conditions

### Predefined Profiles

```java
// Mobile networks
NetworkConditions twoG = NetworkConditionsBuilder.twoG();
NetworkConditions threeG = NetworkConditionsBuilder.threeG(); 
NetworkConditions fourG = NetworkConditionsBuilder.fourG();
NetworkConditions fiveG = NetworkConditionsBuilder.fiveG();

// Other connections
NetworkConditions wifi = NetworkConditionsBuilder.wiFi();
NetworkConditions satellite = NetworkConditionsBuilder.satellite();
```

### Custom Conditions

```java
NetworkConditions custom = NetworkConditionsBuilder.create()
    .latency(100)                     // 100ms latency
    .jitter(20)                       // 20ms jitter  
    .bandwidth(5_000_000, 1_000_000)  // 5Mbps down, 1Mbps up
    .packetLoss(0.1f)                 // 0.1% packet loss
    .build();

client.networkControl().applyConditions("eth0", custom);
```

## System Monitoring

```java
// Health check
var health = client.monitoring().getHealth();
boolean isHealthy = health.getStatus() == HealthStatus.HEALTH_STATUS_HEALTHY;

// System metrics
var metrics = client.monitoring().getMetrics();
float cpuUsage = metrics.getSystem().getCpuUsage();
float memoryUsage = metrics.getSystem().getMemoryUsage();

// Network interface stats
metrics.getInterfacesList().forEach(iface -> 
    System.out.printf("%s: ↓%d bytes ↑%d bytes\n",
        iface.getInterface(),
        iface.getBytesReceived(), 
        iface.getBytesTransmitted()));
```

## Error Handling

```java
try {
    var response = client.networkControl().applyConditions("eth0", conditions);
    if (!response.getSuccess()) {
        System.err.println("Failed: " + response.getErrorMessage());
    }
} catch (StatusRuntimeException e) {
    System.err.println("gRPC error: " + e.getStatus());
} catch (Exception e) {
    System.err.println("Unexpected error: " + e.getMessage());
}
```

## Requirements

- Java 11 or higher
- NetTestLab server running

## Building

```bash
mvn clean compile
mvn test
mvn package
```

## License

MIT License - see [LICENSE](../../LICENSE) file for details.