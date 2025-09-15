package server

import (
	"context"
	"fmt"
	"runtime"
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

// Helper functions for system metrics (placeholders)
func getCPUUsage() float32 {
	// TODO Placeholder - in real implementation, read from /proc/stat
	return 15.5
}

func getMemoryUsage() float32 {
	// TODO Placeholder - in real implementation, read from /proc/meminfo
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
