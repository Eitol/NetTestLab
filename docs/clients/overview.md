# Overview

NetTestLab provides multiple ways to interact with the network simulation engine. Choose the client that best fits your development environment and use case.

## Available Clients

| Language | Status | Use Case |
|----------|--------|----------|
| **Go** | ✅ Ready | Backend services, CLI tools |
| **Python** | ✅ Ready | Testing automation, data analysis |
| **JavaScript/TypeScript** | ✅ Ready | Web applications, Node.js services |
| **Java** | ✅ Ready | Enterprise applications, Android |
| **Dart** | ✅ Ready | Flutter mobile applications |

## Quick Start

All clients are generated from the same Protocol Buffer definitions, ensuring consistent APIs across languages.

### Go Client

```go
import "github.com/Eitol/NetTestLab/clients/go"

client, err := nettestlab.NewClient("router-ip:8080")
if err != nil {
    log.Fatal(err)
}

// Apply 3G profile
err = client.ApplyProfile(ctx, "wifi", "3g")
```

### Python Client

```python
import nettestlab

client = nettestlab.Client("router-ip:8080")

# Apply 3G profile
client.apply_profile("wifi", "3g")
```

### JavaScript/TypeScript Client

```typescript
import { NetTestLabClient } from '@nettestlab/client';

const client = new NetTestLabClient('http://router-ip:8080');

// Apply 3G profile
await client.applyProfile('wifi', '3g');
```

## Authentication

Currently, NetTestLab operates without authentication for simplicity in development environments. For production deployments, consider:

- Running on isolated networks
- Using VPN access
- Implementing firewall rules
- Adding reverse proxy with authentication

## Error Handling

All clients implement consistent error handling patterns:

- **Connection errors**: Network connectivity issues
- **Validation errors**: Invalid parameters or requests
- **System errors**: Router configuration or resource issues

## Best Practices

### Connection Management

- Reuse client connections when possible
- Implement connection pooling for high-throughput scenarios
- Handle connection timeouts gracefully

### Error Recovery

- Implement retry logic with exponential backoff
- Reset network conditions on application exit
- Monitor system status before applying conditions

### Testing Integration

- Use profiles for consistent test scenarios
- Reset conditions between test cases
- Monitor network metrics during tests

## Next Steps

- **[Go Client →](go.md)** - Detailed Go client documentation
- **[Python Client →](python.md)** - Python client examples and patterns
- **[JavaScript Client →](javascript.md)** - Web and Node.js integration
- **[Java Client →](java.md)** - Enterprise and Android development
- **[Dart Client →](dart.md)** - Flutter mobile applications