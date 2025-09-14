package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	nettestlabv1 "github.com/nettestlab/nettestlab/api/nettestlab/v1"
	"github.com/nettestlab/nettestlab/internal/network"
)

// NetworkControlService implements the gRPC network control service
type NetworkControlService struct {
	nettestlabv1.UnimplementedNetworkControlServiceServer
	controller *network.Controller
}

// NewNetworkControlService creates a new network control service
func NewNetworkControlService(controller *network.Controller) *NetworkControlService {
	return &NetworkControlService{
		controller: controller,
	}
}

// ApplyNetworkConditions applies network conditions to an interface
func (s *NetworkControlService) ApplyNetworkConditions(ctx context.Context, req *nettestlabv1.ApplyNetworkConditionsRequest) (*nettestlabv1.ApplyNetworkConditionsResponse, error) {
	// Validate request
	if req.Interface == "" {
		return nil, status.Error(codes.InvalidArgument, "interface name is required")
	}
	if req.Conditions == nil {
		return nil, status.Error(codes.InvalidArgument, "network conditions are required")
	}

	// Set default direction if not specified
	direction := req.Direction
	if direction == nettestlabv1.TrafficDirection_TRAFFIC_DIRECTION_UNSPECIFIED {
		direction = nettestlabv1.TrafficDirection_TRAFFIC_DIRECTION_BOTH
	}

	// Apply conditions
	err := s.controller.ApplyConditions(req.Interface, req.Conditions, direction)
	if err != nil {
		return &nettestlabv1.ApplyNetworkConditionsResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &nettestlabv1.ApplyNetworkConditionsResponse{
		Success:           true,
		AppliedConditions: req.Conditions,
	}, nil
}

// ResetNetworkConditions removes all network conditions from an interface
func (s *NetworkControlService) ResetNetworkConditions(ctx context.Context, req *nettestlabv1.ResetNetworkConditionsRequest) (*nettestlabv1.ResetNetworkConditionsResponse, error) {
	// Validate request
	if req.Interface == "" {
		return nil, status.Error(codes.InvalidArgument, "interface name is required")
	}

	// Set default direction if not specified
	direction := req.Direction
	if direction == nettestlabv1.TrafficDirection_TRAFFIC_DIRECTION_UNSPECIFIED {
		direction = nettestlabv1.TrafficDirection_TRAFFIC_DIRECTION_BOTH
	}

	// Reset conditions
	err := s.controller.ResetConditions(req.Interface, direction)
	if err != nil {
		return &nettestlabv1.ResetNetworkConditionsResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &nettestlabv1.ResetNetworkConditionsResponse{
		Success: true,
	}, nil
}

// GetNetworkConditions returns current network conditions for an interface
func (s *NetworkControlService) GetNetworkConditions(ctx context.Context, req *nettestlabv1.GetNetworkConditionsRequest) (*nettestlabv1.GetNetworkConditionsResponse, error) {
	// Validate request
	if req.Interface == "" {
		return nil, status.Error(codes.InvalidArgument, "interface name is required")
	}

	// Get conditions
	conditions, active := s.controller.GetConditions(req.Interface)
	if !active {
		// Return empty conditions if none are active
		conditions = &nettestlabv1.NetworkConditions{}
	}

	return &nettestlabv1.GetNetworkConditionsResponse{
		Conditions:  conditions,
		Active:      active,
		LastApplied: timestamppb.Now(), // For now, use current time
	}, nil
}

// GetSystemStatus returns system status and available interfaces
func (s *NetworkControlService) GetSystemStatus(ctx context.Context, req *nettestlabv1.GetSystemStatusRequest) (*nettestlabv1.GetSystemStatusResponse, error) {
	// Refresh interface information
	if err := s.controller.RefreshInterfaces(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to refresh interfaces: %v", err)
	}

	// Get interfaces
	interfaceInfos := s.controller.GetInterfaces()
	var interfaces []*nettestlabv1.NetworkInterface

	for _, info := range interfaceInfos {
		// Check if interface has active conditions
		_, hasConditions := s.controller.GetConditions(info.Name)

		interfaces = append(interfaces, &nettestlabv1.NetworkInterface{
			Name:          info.Name,
			Type:          info.Type,
			IsUp:          info.IsUp,
			IpAddresses:   info.IPAddresses,
			HasConditions: hasConditions,
		})
	}

	return &nettestlabv1.GetSystemStatusResponse{
		Interfaces: interfaces,
		Version:    "1.0.0",
		Load: &nettestlabv1.SystemLoad{
			CpuUsage:          10.0, // Placeholder
			MemoryUsage:       30.0, // Placeholder
			ActiveConnections: 1,    // Placeholder
		},
	}, nil
}
