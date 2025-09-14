# Mobile Testing Guide

NetTestLab is designed with mobile application testing in mind. This guide shows you how to simulate real-world mobile network conditions to test your apps under various connectivity scenarios.

## Why Test Under Different Network Conditions?

Mobile applications face diverse network environments:

- **Variable bandwidth**: From 2G to 5G and WiFi
- **High latency**: Satellite, congested networks
- **Packet loss**: Poor signal, network switching
- **Intermittent connectivity**: Tunnels, elevators, rural areas

Testing under these conditions helps ensure your app provides a good user experience regardless of network quality.

## Setup for Mobile Testing

### 1. Router Configuration

Ensure your OpenWrt router with NetTestLab is the gateway for your test devices:

```bash
# Verify WiFi interface is available
ip link show wifi

# Check current conditions
grpcurl -plaintext router:8080 nettestlab.v1.NetworkControlService/GetSystemStatus
```

### 2. Device Connection

Connect your mobile test devices to the router's WiFi network. All traffic will pass through NetTestLab's traffic control.

### 3. Baseline Testing

Before applying conditions, establish baseline performance:

```bash
# Ensure no conditions are applied
grpcurl -plaintext -d '{"interface": "wifi"}' \
  router:8080 nettestlab.v1.NetworkControlService/ResetNetworkConditions
```

## Common Mobile Testing Scenarios

### Scenario 1: Progressive Network Degradation

Test how your app handles progressively worse network conditions:

```bash
# Start with good conditions (WiFi)
grpcurl -plaintext -d '{"interface": "wifi", "profile_name": "wifi"}' \
  router:8080 nettestlab.v1.ProfileService/ApplyProfile

# Test your app functionality
# ...

# Degrade to 4G
grpcurl -plaintext -d '{"interface": "wifi", "profile_name": "4g"}' \
  router:8080 nettestlab.v1.ProfileService/ApplyProfile

# Test again
# ...

# Further degrade to 3G
grpcurl -plaintext -d '{"interface": "wifi", "profile_name": "3g"}' \
  router:8080 nettestlab.v1.ProfileService/ApplyProfile

# Test again
# ...

# Finally test on 2G (worst case)
grpcurl -plaintext -d '{"interface": "wifi", "profile_name": "2g"}' \
  router:8080 nettestlab.v1.ProfileService/ApplyProfile
```

### Scenario 2: High Latency Testing

Simulate satellite or long-distance connections:

```bash
# Apply high latency conditions
grpcurl -plaintext -d '{
  "interface": "wifi",
  "conditions": {
    "latency": {"delay_ms": 800, "enabled": true},
    "bandwidth": {"download_bps": 25000000, "upload_bps": 3000000, "enabled": true}
  }
}' router:8080 nettestlab.v1.NetworkControlService/ApplyNetworkConditions
```

Test scenarios:
- Real-time features (chat, video calls)
- Request timeouts
- User interface responsiveness
- Loading states and feedback

### Scenario 3: Packet Loss Simulation

Test resilience to network instability:

```bash
# Apply moderate packet loss
grpcurl -plaintext -d '{
  "interface": "wifi",
  "conditions": {
    "packet_loss": {"percentage": 5.0, "enabled": true},
    "bandwidth": {"download_bps": 10000000, "upload_bps": 5000000, "enabled": true}
  }
}' router:8080 nettestlab.v1.NetworkControlService/ApplyNetworkConditions
```

Monitor:
- Automatic retry mechanisms
- Error handling and user feedback
- Data synchronization
- Media playback quality

### Scenario 4: Bandwidth Limitations

Test with severely limited bandwidth:

```bash
# Very low bandwidth simulation
grpcurl -plaintext -d '{
  "interface": "wifi",
  "conditions": {
    "bandwidth": {"download_bps": 128000, "upload_bps": 64000, "enabled": true}
  }
}' router:8080 nettestlab.v1.NetworkControlService/ApplyNetworkConditions
```

Focus on:
- Image compression and lazy loading
- Video quality adaptation
- Background sync behavior
- Cache effectiveness

## Testing Workflows

### Automated Testing

Integrate NetTestLab into your CI/CD pipeline:

```python
import pytest
import nettestlab

class TestNetworkConditions:
    def setup_method(self):
        self.client = nettestlab.Client("router:8080")
        # Reset conditions before each test
        self.client.reset_network_conditions("wifi")
    
    def teardown_method(self):
        # Clean up after each test
        self.client.reset_network_conditions("wifi")
    
    def test_app_on_3g(self):
        # Apply 3G conditions
        self.client.apply_profile("wifi", "3g")
        
        # Run your app tests here
        result = run_app_tests()
        assert result.success
        
        # Verify reasonable performance under 3G
        assert result.load_time < 10.0  # seconds
    
    def test_app_with_packet_loss(self):
        # Apply packet loss
        self.client.apply_network_conditions("wifi", {
            "packet_loss": {"percentage": 2.0, "enabled": True}
        })
        
        # Test resilience
        result = run_app_tests()
        assert result.success
```

### Manual Testing Checklist

**Initial Setup:**
- [ ] Router and NetTestLab functioning
- [ ] Test devices connected
- [ ] Baseline measurements taken

**For Each Network Profile:**
- [ ] App launches successfully
- [ ] Login/authentication works
- [ ] Core features functional
- [ ] Error handling appropriate
- [ ] Loading states visible
- [ ] Offline capabilities work
- [ ] Data synchronization correct
- [ ] Media playback quality adequate

**Performance Checks:**
- [ ] App remains responsive
- [ ] Memory usage reasonable
- [ ] Battery impact acceptable
- [ ] Network requests efficient

## Testing Different App Types

### Social Media Apps

Focus on:
- Image/video loading strategies
- Feed refresh behavior
- Message delivery and sync
- Story/content caching

Test with: 2G, 3G profiles and high packet loss

### E-commerce Apps

Focus on:
- Product image loading
- Checkout process reliability
- Search responsiveness
- Cart synchronization

Test with: All profiles, especially bandwidth limitations

### Video Streaming Apps

Focus on:
- Adaptive bitrate streaming
- Buffer management
- Quality degradation
- Pause/resume behavior

Test with: Progressive degradation from WiFi to 2G

### Banking/Finance Apps

Focus on:
- Security timeout handling
- Transaction reliability
- Offline mode capabilities
- Sync accuracy

Test with: High latency and packet loss scenarios

## Monitoring and Metrics

### Key Metrics to Track

- **Time to First Byte (TTFB)**
- **Page Load Time**
- **API Response Times**
- **Failed Request Rate**
- **Battery Usage**
- **Memory Consumption**

### Using Built-in Monitoring

```bash
# Monitor system during testing
grpcurl -plaintext router:8080 nettestlab.v1.MonitoringService/GetSystemMetrics

# Get interface statistics
grpcurl -plaintext -d '{"interface": "wifi"}' \
  router:8080 nettestlab.v1.MonitoringService/GetInterfaceInfo
```

## Best Practices

### Test Planning

1. **Start with baseline**: Test without any conditions first
2. **Use realistic profiles**: Based on your user demographics
3. **Test edge cases**: Extreme conditions your users might face
4. **Automate when possible**: Include network tests in CI/CD

### During Testing

1. **Reset between tests**: Ensure clean state
2. **Monitor system resources**: CPU, memory, network usage
3. **Document issues**: Note conditions when problems occur
4. **Test both directions**: Upload and download scenarios

### Issue Investigation

1. **Reproduce conditions**: Use specific profiles/conditions
2. **Isolate variables**: Test individual conditions separately
3. **Check logs**: App logs and NetTestLab system logs
4. **Compare profiles**: Test against multiple network types

## Troubleshooting

### Common Issues

**App doesn't work on slow networks:**
- Check timeout configurations
- Implement proper loading states
- Add retry mechanisms
- Consider caching strategies

**High battery drain on poor networks:**
- Reduce polling frequency
- Implement exponential backoff
- Use push notifications instead of polling
- Optimize background sync

**Poor user experience on high latency:**
- Implement optimistic UI updates
- Add immediate visual feedback
- Cache critical data locally
- Use background prefetching

---

**Related:** [Performance Testing →](performance-testing.md) | [API Reference →](../api/overview.md) | [Client Libraries →](../clients/overview.md)