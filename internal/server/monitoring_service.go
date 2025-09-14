package server

import (
	"context"
	"runtime"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	nettestlabv1 "github.com/Eitol/NetTestLab/api/nettestlab/v1"
	"github.com/Eitol/NetTestLab/internal/network"
)

var startTime = time.Now()

// MonitoringService implements the gRPC monitoring service
type MonitoringService struct {
	nettestlabv1.UnimplementedMonitoringServiceServer
	controller *network.Controller
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(controller *network.Controller) *MonitoringService {
	return &MonitoringService{
		controller: controller,
	}
}

// GetHealth returns system health status
func (s *MonitoringService) GetHealth(ctx context.Context, req *nettestlabv1.GetHealthRequest) (*nettestlabv1.GetHealthResponse, error) {
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

	return &nettestlabv1.GetHealthResponse{
		Status:     nettestlabv1.HealthStatus_HEALTH_STATUS_HEALTHY,
		Timestamp:  timestamppb.Now(),
		Components: components,
		Uptime:     durationpb.New(uptime),
		Version:    "1.0.0",
	}, nil
}

// GetMetrics returns system metrics
func (s *MonitoringService) GetMetrics(ctx context.Context, req *nettestlabv1.GetMetricsRequest) (*nettestlabv1.GetMetricsResponse, error) {
	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	systemMetrics := &nettestlabv1.SystemMetrics{
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

	// Get interface metrics
	var interfaceMetrics []*nettestlabv1.InterfaceMetrics
	interfaces := s.controller.GetInterfaces()

	for name := range interfaces {
		metrics := getInterfaceMetrics(name)
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

	return &nettestlabv1.GetMetricsResponse{
		System:     systemMetrics,
		Interfaces: interfaceMetrics,
		Nettestlab: netTestLabMetrics,
		Timestamp:  timestamppb.Now(),
	}, nil
}

// GetInterfaceStats returns interface-specific statistics
func (s *MonitoringService) GetInterfaceStats(ctx context.Context, req *nettestlabv1.GetInterfaceStatsRequest) (*nettestlabv1.GetInterfaceStatsResponse, error) {
	if req.Interface == "" {
		return nil, nil
	}

	stats := &nettestlabv1.TrafficStats{
		TotalBytes:      1000000, // Placeholder
		TotalPackets:    10000,   // Placeholder
		AffectedPackets: 500,     // Placeholder
		AvgLatencyMs:    10.5,    // Placeholder
		LossRate:        0.1,     // Placeholder
	}

	return &nettestlabv1.GetInterfaceStatsResponse{
		Interface: req.Interface,
		Stats:     stats,
	}, nil
}

// StreamMetrics streams real-time metrics (placeholder implementation)
func (s *MonitoringService) StreamMetrics(req *nettestlabv1.StreamMetricsRequest, stream nettestlabv1.MonitoringService_StreamMetricsServer) error {
	ticker := time.NewTicker(5 * time.Second) // Default 5 second interval
	if req.Interval != nil {
		ticker = time.NewTicker(req.Interval.AsDuration())
	}
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
			// Get current metrics (simplified)
			update := &nettestlabv1.MetricsUpdate{
				Timestamp: timestamppb.Now(),
				System: &nettestlabv1.SystemMetrics{
					CpuUsage:    getCPUUsage(),
					MemoryUsage: getMemoryUsage(),
				},
			}

			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
}

// GetHistoricalMetrics returns historical metrics (placeholder)
func (s *MonitoringService) GetHistoricalMetrics(ctx context.Context, req *nettestlabv1.GetHistoricalMetricsRequest) (*nettestlabv1.GetHistoricalMetricsResponse, error) {
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

	return &nettestlabv1.GetHistoricalMetricsResponse{
		DataPoints: dataPoints,
		Metadata:   metadata,
	}, nil
}

// Helper functions for system metrics (placeholders)
func getCPUUsage() float32 {
	// Placeholder - in real implementation, read from /proc/stat
	return 15.5
}

func getMemoryUsage() float32 {
	// Placeholder - in real implementation, read from /proc/meminfo
	return 45.2
}

func getDiskUsage() float32 {
	// Placeholder - in real implementation, use syscall.Statfs
	return 60.0
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
