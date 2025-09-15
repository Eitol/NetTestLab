package network

import (
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	nettestlabv1 "github.com/Eitol/NetTestLab/api/nettestlab/v1"
)

// EnvironmentDetector handles detection of the running environment
type EnvironmentDetector struct {
	isRouter    bool
	fakeDataGen *FakeDataGenerator
}

// FakeDataGenerator generates realistic fake data for non-router environments
type FakeDataGenerator struct {
	rand           *rand.Rand
	startTime      time.Time
	interfaceStats map[string]*InterfaceStats
}

// InterfaceStats holds fake statistics for an interface
type InterfaceStats struct {
	BytesReceived      uint64
	BytesTransmitted   uint64
	PacketsReceived    uint64
	PacketsTransmitted uint64
	LastUpdate         time.Time
	BaseRate           uint64 // bytes per second baseline
}

// NewEnvironmentDetector creates a new environment detector
func NewEnvironmentDetector() *EnvironmentDetector {
	detector := &EnvironmentDetector{
		isRouter: detectRouter(),
		fakeDataGen: &FakeDataGenerator{
			rand:           rand.New(rand.NewSource(time.Now().UnixNano())),
			startTime:      time.Now(),
			interfaceStats: make(map[string]*InterfaceStats),
		},
	}

	// Initialize fake stats for common interfaces
	if !detector.isRouter {
		detector.fakeDataGen.initializeFakeStats()
	}

	return detector
}

// IsRouter returns true if running on a router
func (e *EnvironmentDetector) IsRouter() bool {
	return e.isRouter
}

// detectRouter attempts to determine if we're running on an OpenWrt router
func detectRouter() bool {
	// Method 1: Check for OpenWrt specific files
	openWrtFiles := []string{
		"/etc/openwrt_release",
		"/etc/openwrt_version",
		"/usr/lib/opkg",
	}

	for _, file := range openWrtFiles {
		if _, err := os.Stat(file); err == nil {
			return true
		}
	}

	// Method 2: Check for UCI config system
	if _, err := os.Stat("/sbin/uci"); err == nil {
		return true
	}

	// Method 3: Check for typical router directory structure
	routerDirs := []string{
		"/overlay",
		"/rom",
	}

	routerDirCount := 0
	for _, dir := range routerDirs {
		if _, err := os.Stat(dir); err == nil {
			routerDirCount++
		}
	}

	// If we find multiple router-specific directories, likely a router
	if routerDirCount >= 2 {
		return true
	}

	// Method 4: Check kernel version and architecture patterns common in routers
	if runtime.GOOS == "linux" {
		if data, err := os.ReadFile("/proc/version"); err == nil {
			version := string(data)
			// Look for common router firmware indicators
			routerIndicators := []string{
				"OpenWrt",
				"LEDE",
				"mips",
				"arm-openwrt",
				"bcm",
			}

			for _, indicator := range routerIndicators {
				if strings.Contains(version, indicator) {
					return true
				}
			}
		}
	}

	// Method 5: Check for router-specific network interfaces
	if runtime.GOOS == "linux" {
		// Look for common router interface patterns
		routerInterfacePatterns := []string{
			"br-lan",
			"wl0",
			"wl1",
			"ath0",
			"ath1",
		}

		for _, pattern := range routerInterfacePatterns {
			if _, err := os.Stat("/sys/class/net/" + pattern); err == nil {
				return true
			}
		}
	}

	// Default: assume not a router
	return false
}

// initializeFakeStats initializes fake statistics for common interfaces
func (f *FakeDataGenerator) initializeFakeStats() {
	interfaces := []struct {
		name string
		rate uint64 // bytes per second baseline
	}{
		{"eth0", 10000000},    // 10 MB/s
		{"wlan0", 5000000},    // 5 MB/s
		{"lo", 1000000},       // 1 MB/s
		{"lo0", 1000000},      // 1 MB/s (macOS)
		{"en0", 8000000},      // 8 MB/s (macOS Ethernet)
		{"en1", 6000000},      // 6 MB/s (macOS WiFi)
		{"Wi-Fi", 6000000},    // 6 MB/s (Windows WiFi)
		{"Ethernet", 8000000}, // 8 MB/s (Windows Ethernet)
		{"wifi0", 6000000},    // 6 MB/s
		{"docker0", 2000000},  // 2 MB/s (Docker bridge)
		{"br-lan", 15000000},  // 15 MB/s (Router LAN bridge)
		{"wl0", 5000000},      // 5 MB/s (Router WiFi)
		{"wl1", 8000000},      // 8 MB/s (Router WiFi 5GHz)
	}

	for _, iface := range interfaces {
		f.interfaceStats[iface.name] = &InterfaceStats{
			BytesReceived:      f.randomInitialBytes(),
			BytesTransmitted:   f.randomInitialBytes(),
			PacketsReceived:    f.randomInitialPackets(),
			PacketsTransmitted: f.randomInitialPackets(),
			LastUpdate:         f.startTime,
			BaseRate:           iface.rate,
		}
	}
}

// GetFakeInterfaceMetrics generates fake but realistic interface metrics
func (e *EnvironmentDetector) GetFakeInterfaceMetrics(iface string) *nettestlabv1.InterfaceMetrics {
	if e.isRouter {
		return nil // Use real data on routers
	}

	return e.fakeDataGen.generateInterfaceMetrics(iface)
}

// generateInterfaceMetrics creates realistic fake metrics for an interface
func (f *FakeDataGenerator) generateInterfaceMetrics(iface string) *nettestlabv1.InterfaceMetrics {
	stats, exists := f.interfaceStats[iface]
	if !exists {
		// Create new stats for unknown interface
		stats = &InterfaceStats{
			BytesReceived:      f.randomInitialBytes(),
			BytesTransmitted:   f.randomInitialBytes(),
			PacketsReceived:    f.randomInitialPackets(),
			PacketsTransmitted: f.randomInitialPackets(),
			LastUpdate:         f.startTime,
			BaseRate:           uint64(f.rand.Intn(10000000) + 1000000), // 1-10 MB/s
		}
		f.interfaceStats[iface] = stats
	}

	// Update stats based on time elapsed
	f.updateStats(stats)

	// Calculate current bandwidth utilization
	rxBps := f.calculateCurrentRate(stats, true)
	txBps := f.calculateCurrentRate(stats, false)
	maxBandwidth := stats.BaseRate
	utilization := float32(rxBps+txBps) / float32(maxBandwidth*2) * 100

	return &nettestlabv1.InterfaceMetrics{
		Interface:          iface,
		BytesReceived:      stats.BytesReceived,
		BytesTransmitted:   stats.BytesTransmitted,
		PacketsReceived:    stats.PacketsReceived,
		PacketsTransmitted: stats.PacketsTransmitted,
		Bandwidth: &nettestlabv1.BandwidthUtilization{
			RxBps:              rxBps,
			TxBps:              txBps,
			UtilizationPercent: utilization,
		},
	}
}

// updateStats updates the fake statistics based on elapsed time
func (f *FakeDataGenerator) updateStats(stats *InterfaceStats) {
	now := time.Now()
	elapsed := now.Sub(stats.LastUpdate)

	if elapsed < time.Second {
		return // Don't update too frequently
	}

	// Calculate bytes to add based on base rate and some randomness
	variation := 0.3 // 30% variation
	multiplier := 1.0 + (f.rand.Float64()-0.5)*variation
	bytesToAdd := uint64(float64(stats.BaseRate) * elapsed.Seconds() * multiplier)

	// Update received bytes (slightly higher than transmitted for typical usage)
	rxMultiplier := 1.2 + f.rand.Float64()*0.3 // 1.2x to 1.5x
	stats.BytesReceived += uint64(float64(bytesToAdd) * rxMultiplier)
	stats.BytesTransmitted += bytesToAdd

	// Update packet counts (assume average packet size of 1200 bytes)
	avgPacketSize := uint64(1200)
	stats.PacketsReceived += uint64(float64(bytesToAdd)*rxMultiplier) / avgPacketSize
	stats.PacketsTransmitted += bytesToAdd / avgPacketSize

	stats.LastUpdate = now
}

// calculateCurrentRate calculates the current transfer rate
func (f *FakeDataGenerator) calculateCurrentRate(stats *InterfaceStats, isRx bool) uint64 {
	// Simulate variable network activity
	baseRate := stats.BaseRate

	// Add some randomness to simulate real network activity
	variation := f.rand.Float64() * 0.4  // 0-40% of base rate
	currentMultiplier := 0.1 + variation // 10-50% of base rate typically

	// Occasionally spike the activity
	if f.rand.Intn(20) == 0 { // 5% chance of high activity
		currentMultiplier = 0.8 + f.rand.Float64()*0.2 // 80-100% of base rate
	}

	rate := uint64(float64(baseRate) * currentMultiplier)

	if isRx {
		// RX is typically higher than TX for client devices
		rate = uint64(float64(rate) * (1.2 + f.rand.Float64()*0.3))
	}

	return rate
}

// randomInitialBytes generates a random initial byte count
func (f *FakeDataGenerator) randomInitialBytes() uint64 {
	// Generate realistic initial values (as if the system has been running for a while)
	base := uint64(f.rand.Intn(1000000000))      // 0-1GB
	return base + uint64(f.rand.Intn(100000000)) // Add up to 100MB more
}

// randomInitialPackets generates a random initial packet count
func (f *FakeDataGenerator) randomInitialPackets() uint64 {
	// Assume average packet size of 1200 bytes
	return f.randomInitialBytes() / 1200
}

// GetFakeSystemMetrics generates fake system metrics
func (e *EnvironmentDetector) GetFakeSystemMetrics() (*nettestlabv1.SystemMetrics, bool) {
	if e.isRouter {
		return nil, false // Use real data on routers
	}

	return e.fakeDataGen.generateSystemMetrics(), true
}

// generateSystemMetrics creates realistic fake system metrics
func (f *FakeDataGenerator) generateSystemMetrics() *nettestlabv1.SystemMetrics {
	// Generate realistic but fake CPU usage (varying over time)
	baseLoad := 15.0 + f.rand.Float64()*20.0        // 15-35% base load
	cpuVariation := (f.rand.Float64() - 0.5) * 10.0 // ±5% variation
	cpuUsage := float32(baseLoad + cpuVariation)
	if cpuUsage < 0 {
		cpuUsage = 0
	}
	if cpuUsage > 100 {
		cpuUsage = 100
	}

	// Generate memory usage (more stable than CPU)
	memUsage := float32(40.0 + f.rand.Float64()*20.0) // 40-60%

	// Generate other metrics
	totalMemory := uint64(8 * 1024 * 1024 * 1024) // 8GB typical
	usedMemory := uint64(float64(totalMemory) * float64(memUsage) / 100.0)
	availableMemory := totalMemory - usedMemory

	return &nettestlabv1.SystemMetrics{
		CpuUsage:           cpuUsage,
		MemoryUsage:        memUsage,
		TotalMemory:        totalMemory,
		AvailableMemory:    availableMemory,
		DiskUsage:          float32(45.0 + f.rand.Float64()*30.0), // 45-75%
		NetworkConnections: uint32(10 + f.rand.Intn(50)),          // 10-60 connections
		LoadAverage: &nettestlabv1.LoadAverage{
			OneMinute:      float32(float64(cpuUsage)/100.0*2.0 + f.rand.Float64()*0.5),
			FiveMinutes:    float32(float64(cpuUsage)/100.0*2.0 + f.rand.Float64()*0.3),
			FifteenMinutes: float32(float64(cpuUsage)/100.0*2.0 + f.rand.Float64()*0.2),
		},
	}
}

// GetFakeTrafficStats generates fake traffic statistics for interface testing
func (e *EnvironmentDetector) GetFakeTrafficStats(iface string) (*nettestlabv1.TrafficStats, bool) {
	if e.isRouter {
		return nil, false // Use real data on routers
	}

	return e.fakeDataGen.generateTrafficStats(iface), true
}

// generateTrafficStats creates realistic fake traffic statistics
func (f *FakeDataGenerator) generateTrafficStats(iface string) *nettestlabv1.TrafficStats {
	stats, exists := f.interfaceStats[iface]
	if !exists {
		stats = &InterfaceStats{
			BytesReceived:      f.randomInitialBytes(),
			BytesTransmitted:   f.randomInitialBytes(),
			PacketsReceived:    f.randomInitialPackets(),
			PacketsTransmitted: f.randomInitialPackets(),
			LastUpdate:         f.startTime,
			BaseRate:           uint64(f.rand.Intn(10000000) + 1000000),
		}
		f.interfaceStats[iface] = stats
	}

	f.updateStats(stats)

	// Calculate some affected packets (simulating traffic control effects)
	totalPackets := stats.PacketsReceived + stats.PacketsTransmitted
	affectedPackets := uint64(float64(totalPackets) * (0.01 + f.rand.Float64()*0.05)) // 1-6% affected

	// Generate realistic latency based on interface type
	avgLatency := 5.0 + f.rand.Float64()*15.0 // 5-20ms base
	if strings.Contains(iface, "wifi") || strings.Contains(iface, "wlan") {
		avgLatency += 10.0 + f.rand.Float64()*10.0 // WiFi adds 10-20ms more
	}

	// Generate loss rate
	lossRate := f.rand.Float64() * 0.5 // 0-0.5% loss

	return &nettestlabv1.TrafficStats{
		TotalBytes:      stats.BytesReceived + stats.BytesTransmitted,
		TotalPackets:    totalPackets,
		AffectedPackets: affectedPackets,
		AvgLatencyMs:    float32(avgLatency),
		LossRate:        float32(lossRate),
	}
}
