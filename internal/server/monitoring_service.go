package server

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	nettestlabv1 "github.com/Eitol/NetTestLab/api/nettestlab/v1"
	"github.com/Eitol/NetTestLab/internal/network"
)

var startTime = time.Now()

// MonitoringService implements the Connect monitoring service
type MonitoringService struct {
	controller *network.Controller
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(controller *network.Controller) *MonitoringService {
	return &MonitoringService{
		controller: controller,
	}
}

// GetHealth returns system health status
func (s *MonitoringService) GetHealth(ctx context.Context, req *connect.Request[nettestlabv1.GetHealthRequest]) (*connect.Response[nettestlabv1.GetHealthResponse], error) {
	uptime := time.Since(startTime)

	components := []*nettestlabv1.ComponentHealth{
		{
			Name:      "network_controller",
			Status:    nettestlabv1.HealthStatus_HEALTH_STATUS_HEALTHY,
			LastCheck: timestamppb.Now(),
		},
		{
			Name:      "profile_manager",
			Status:    nettestlabv1.HealthStatus_HEALTH_STATUS_HEALTHY,
			LastCheck: timestamppb.Now(),
		},
	}

	return connect.NewResponse(&nettestlabv1.GetHealthResponse{
		Status:     nettestlabv1.HealthStatus_HEALTH_STATUS_HEALTHY,
		Timestamp:  timestamppb.Now(),
		Components: components,
		Uptime:     durationpb.New(uptime),
		Version:    "1.0.0",
	}), nil
}

// GetMetrics returns system metrics
func (s *MonitoringService) GetMetrics(ctx context.Context, req *connect.Request[nettestlabv1.GetMetricsRequest]) (*connect.Response[nettestlabv1.GetMetricsResponse], error) {
	var systemMetrics *nettestlabv1.SystemMetrics

	// Check if we should use fake data (non-router systems)
	if !s.controller.IsRouter() {
		// Use fake system metrics for non-router systems
		if fakeMetrics, isFake := s.controller.GetFakeSystemMetrics(); isFake {
			systemMetrics = fakeMetrics
		} else {
			// Fallback to real metrics calculation
			systemMetrics = s.getRealSystemMetrics()
		}
	} else {
		// Use real system metrics for router systems
		systemMetrics = s.getRealSystemMetrics()
	}

	// Get interface metrics
	var interfaceMetrics []*nettestlabv1.InterfaceMetrics
	interfaces := s.controller.GetInterfaces()

	for name := range interfaces {
		metrics := s.controller.GetInterfaceMetrics(name)
		if metrics != nil {
			interfaceMetrics = append(interfaceMetrics, metrics)
		}
	}

	netTestLabMetrics := &nettestlabv1.NetTestLabMetrics{
		ActiveConditions:    uint32(len(s.controller.GetInterfaces())), // Simplified
		TotalRequests:       100,                                       // Placeholder
		FailedRequests:      5,                                         // Placeholder
		ActiveConnections:   1,                                         // Placeholder
		ProfileApplications: 50,                                        // Placeholder
	}

	return connect.NewResponse(&nettestlabv1.GetMetricsResponse{
		System:     systemMetrics,
		Interfaces: interfaceMetrics,
		Nettestlab: netTestLabMetrics,
		Timestamp:  timestamppb.Now(),
	}), nil
}

// getRealSystemMetrics gets real system metrics (original implementation)
func (s *MonitoringService) getRealSystemMetrics() *nettestlabv1.SystemMetrics {
	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &nettestlabv1.SystemMetrics{
		CpuUsage:           getCPUUsage(),
		MemoryUsage:        getMemoryUsage(),
		TotalMemory:        memStats.Sys,
		AvailableMemory:    memStats.Sys - memStats.Alloc,
		DiskUsage:          getDiskUsage(),
		NetworkConnections: 1, // Placeholder
		LoadAverage: &nettestlabv1.LoadAverage{
			OneMinute:      0.1, // Placeholder
			FiveMinutes:    0.2, // Placeholder
			FifteenMinutes: 0.1, // Placeholder
		},
	}
}

// GetInterfaceStats returns interface-specific statistics
func (s *MonitoringService) GetInterfaceStats(ctx context.Context, req *connect.Request[nettestlabv1.GetInterfaceStatsRequest]) (*connect.Response[nettestlabv1.GetInterfaceStatsResponse], error) {
	if req.Msg.Interface == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("interface name is required"))
	}

	// Use controller's method to get traffic stats (which handles fake data automatically)
	stats := s.controller.GetTrafficStats(req.Msg.Interface)

	return connect.NewResponse(&nettestlabv1.GetInterfaceStatsResponse{
		Interface: req.Msg.Interface,
		Stats:     stats,
	}), nil
}

// StreamMetrics streams real-time metrics (placeholder implementation)
func (s *MonitoringService) StreamMetrics(ctx context.Context, req *connect.Request[nettestlabv1.StreamMetricsRequest], stream *connect.ServerStream[nettestlabv1.MetricsUpdate]) error {
	ticker := time.NewTicker(5 * time.Second) // Default 5 second interval
	if req.Msg.Interval != nil {
		ticker = time.NewTicker(req.Msg.Interval.AsDuration())
	}
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			var systemMetrics *nettestlabv1.SystemMetrics

			// Use fake or real metrics based on system type
			if !s.controller.IsRouter() {
				if fakeMetrics, isFake := s.controller.GetFakeSystemMetrics(); isFake {
					systemMetrics = fakeMetrics
				} else {
					systemMetrics = s.getRealSystemMetrics()
				}
			} else {
				systemMetrics = s.getRealSystemMetrics()
			}

			// Get current metrics (simplified)
			update := &nettestlabv1.MetricsUpdate{
				Timestamp: timestamppb.Now(),
				System:    systemMetrics,
			}

			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
}

// GetHistoricalMetrics returns historical metrics (placeholder)
func (s *MonitoringService) GetHistoricalMetrics(ctx context.Context, req *connect.Request[nettestlabv1.GetHistoricalMetricsRequest]) (*connect.Response[nettestlabv1.GetHistoricalMetricsResponse], error) {
	// Placeholder implementation
	dataPoints := []*nettestlabv1.MetricsDataPoint{
		{
			Timestamp: timestamppb.Now(),
			System: &nettestlabv1.SystemMetrics{
				CpuUsage:    getCPUUsage(),
				MemoryUsage: getMemoryUsage(),
			},
		},
	}

	metadata := &nettestlabv1.QueryMetadata{
		TotalPoints:    1,
		ReturnedPoints: 1,
		ExecutionTime:  durationpb.New(10 * time.Millisecond),
	}

	return connect.NewResponse(&nettestlabv1.GetHistoricalMetricsResponse{
		DataPoints: dataPoints,
		Metadata:   metadata,
	}), nil
}

// Helper functions for system metrics (real implementations)
func getCPUUsage() float32 {
	switch runtime.GOOS {
	case "linux":
		return getCPUUsageLinux()
	case "darwin":
		return getCPUUsageMacOS()
	default:
		// Fallback para otros sistemas
		return 15.5
	}
}

// getCPUUsageLinux obtiene el uso de CPU en Linux leyendo /proc/stat
func getCPUUsageLinux() float32 {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0.0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return 0.0
	}

	line := scanner.Text()
	fields := strings.Fields(line)
	if len(fields) < 8 || fields[0] != "cpu" {
		return 0.0
	}

	// Parse CPU times
	user, _ := strconv.ParseUint(fields[1], 10, 64)
	nice, _ := strconv.ParseUint(fields[2], 10, 64)
	system, _ := strconv.ParseUint(fields[3], 10, 64)
	idle, _ := strconv.ParseUint(fields[4], 10, 64)
	iowait, _ := strconv.ParseUint(fields[5], 10, 64)
	irq, _ := strconv.ParseUint(fields[6], 10, 64)
	softirq, _ := strconv.ParseUint(fields[7], 10, 64)

	// Calculate total and idle time
	totalTime := user + nice + system + idle + iowait + irq + softirq
	idleTime := idle + iowait

	if totalTime == 0 {
		return 0.0
	}

	// Calculate CPU usage percentage
	usage := float64(totalTime-idleTime) / float64(totalTime) * 100.0
	return float32(usage)
}

// getCPUUsageMacOS obtiene el uso de CPU en macOS usando el comando top
func getCPUUsageMacOS() float32 {
	cmd := exec.Command("top", "-l", "1", "-n", "0")
	output, err := cmd.Output()
	if err != nil {
		return 0.0
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "CPU usage:") {
			// Buscar el porcentaje de usuario + sistema
			fields := strings.Fields(line)
			for i, field := range fields {
				if strings.Contains(field, "user") && i > 0 {
					userStr := strings.TrimSuffix(fields[i-1], "%")
					user, err := strconv.ParseFloat(userStr, 32)
					if err == nil {
						// Buscar sistema
						for j := i + 1; j < len(fields)-1; j++ {
							if strings.Contains(fields[j+1], "sys") {
								sysStr := strings.TrimSuffix(fields[j], "%")
								sys, err := strconv.ParseFloat(sysStr, 32)
								if err == nil {
									return float32(user + sys)
								}
							}
						}
					}
				}
			}
		}
	}
	return 0.0
}

func getMemoryUsage() float32 {
	switch runtime.GOOS {
	case "linux":
		return getMemoryUsageLinux()
	case "darwin":
		return getMemoryUsageMacOS()
	default:
		// Fallback usando runtime.MemStats
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		if m.Sys > 0 {
			return float32(m.Alloc) / float32(m.Sys) * 100.0
		}
		return 45.2
	}
}

// getMemoryUsageLinux obtiene el uso de memoria en Linux leyendo /proc/meminfo
func getMemoryUsageLinux() float32 {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0.0
	}
	defer file.Close()

	var memTotal, memAvailable uint64
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			switch fields[0] {
			case "MemTotal:":
				memTotal, _ = strconv.ParseUint(fields[1], 10, 64)
			case "MemAvailable:":
				memAvailable, _ = strconv.ParseUint(fields[1], 10, 64)
			}
		}
	}

	if memTotal > 0 {
		usage := float64(memTotal-memAvailable) / float64(memTotal) * 100.0
		return float32(usage)
	}
	return 0.0
}

// getMemoryUsageMacOS obtiene el uso de memoria en macOS usando vm_stat
func getMemoryUsageMacOS() float32 {
	cmd := exec.Command("vm_stat")
	output, err := cmd.Output()
	if err != nil {
		return 0.0
	}

	lines := strings.Split(string(output), "\n")
	var pageSize, freePages, inactivePages uint64

	// Obtener tamaño de página (típicamente 4096)
	if len(lines) > 0 && strings.Contains(lines[0], "page size") {
		fields := strings.Fields(lines[0])
		for i, field := range fields {
			if field == "size" && i+2 < len(fields) {
				pageSize, _ = strconv.ParseUint(fields[i+2], 10, 64)
				break
			}
		}
	}

	if pageSize == 0 {
		pageSize = 4096 // valor por defecto
	}

	// Parse las páginas
	for _, line := range lines {
		if strings.Contains(line, "Pages free:") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				freeStr := strings.TrimSuffix(fields[2], ".")
				freePages, _ = strconv.ParseUint(freeStr, 10, 64)
			}
		} else if strings.Contains(line, "Pages inactive:") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				inactiveStr := strings.TrimSuffix(fields[2], ".")
				inactivePages, _ = strconv.ParseUint(inactiveStr, 10, 64)
			}
		}
	}

	// Obtener memoria total del sistema usando sysctl
	cmd = exec.Command("sysctl", "-n", "hw.memsize")
	output, err = cmd.Output()
	if err != nil {
		return 0.0
	}

	memTotal, err := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0.0
	}

	// Calcular memoria usada
	freeMemory := (freePages + inactivePages) * pageSize
	usedMemory := memTotal - freeMemory

	if memTotal > 0 {
		usage := float64(usedMemory) / float64(memTotal) * 100.0
		return float32(usage)
	}
	return 0.0
}

func getDiskUsage() float32 {
	switch runtime.GOOS {
	case "linux":
		return getDiskUsageUnix("/")
	case "darwin":
		return getDiskUsageUnix("/")
	default:
		return 60.0 // fallback
	}
}

// getDiskUsageUnix obtiene el uso de disco usando syscall.Statfs
func getDiskUsageUnix(path string) float32 {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0.0
	}

	// Calcular espacio total y disponible
	total := stat.Blocks * uint64(stat.Bsize)
	available := stat.Bavail * uint64(stat.Bsize)
	used := total - available

	if total > 0 {
		usage := float64(used) / float64(total) * 100.0
		return float32(usage)
	}
	return 0.0
}

func getInterfaceMetrics(iface string) *nettestlabv1.InterfaceMetrics {
	// Placeholder - in real implementation, read from /proc/net/dev
	return &nettestlabv1.InterfaceMetrics{
		Interface:          iface,
		BytesReceived:      1000000,
		BytesTransmitted:   500000,
		PacketsReceived:    1000,
		PacketsTransmitted: 500,
		Bandwidth: &nettestlabv1.BandwidthUtilization{
			RxBps:              100000,
			TxBps:              50000,
			UtilizationPercent: 10.5,
		},
	}
}
