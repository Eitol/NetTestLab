package server

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	nettestlabv1 "github.com/Eitol/NetTestLab/api/nettestlab/v1"
	"github.com/Eitol/NetTestLab/internal/network"
)

// NetworkControlService implements the Connect network control service
type NetworkControlService struct {
	controller *network.Controller
}

// NewNetworkControlService creates a new network control service
func NewNetworkControlService(controller *network.Controller) *NetworkControlService {
	return &NetworkControlService{
		controller: controller,
	}
}

// ApplyNetworkConditions applies network conditions to an interface
func (s *NetworkControlService) ApplyNetworkConditions(ctx context.Context, req *connect.Request[nettestlabv1.ApplyNetworkConditionsRequest]) (*connect.Response[nettestlabv1.ApplyNetworkConditionsResponse], error) {
	// Validate request
	if req.Msg.Interface == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("interface name is required"))
	}
	if req.Msg.Conditions == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("network conditions are required"))
	}

	// Set default direction if not specified
	direction := req.Msg.Direction
	if direction == nettestlabv1.TrafficDirection_TRAFFIC_DIRECTION_UNSPECIFIED {
		direction = nettestlabv1.TrafficDirection_TRAFFIC_DIRECTION_BOTH
	}

	// Apply conditions
	err := s.controller.ApplyConditions(req.Msg.Interface, req.Msg.Conditions, direction)
	if err != nil {
		response := &nettestlabv1.ApplyNetworkConditionsResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}
		return connect.NewResponse(response), nil
	}

	response := &nettestlabv1.ApplyNetworkConditionsResponse{
		Success:           true,
		AppliedConditions: req.Msg.Conditions,
	}
	return connect.NewResponse(response), nil
}

// ResetNetworkConditions removes all network conditions from an interface
func (s *NetworkControlService) ResetNetworkConditions(ctx context.Context, req *connect.Request[nettestlabv1.ResetNetworkConditionsRequest]) (*connect.Response[nettestlabv1.ResetNetworkConditionsResponse], error) {
	// Validate request
	if req.Msg.Interface == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("interface name is required"))
	}

	// Set default direction if not specified
	direction := req.Msg.Direction
	if direction == nettestlabv1.TrafficDirection_TRAFFIC_DIRECTION_UNSPECIFIED {
		direction = nettestlabv1.TrafficDirection_TRAFFIC_DIRECTION_BOTH
	}

	// Reset conditions
	err := s.controller.ResetConditions(req.Msg.Interface, direction)
	if err != nil {
		response := &nettestlabv1.ResetNetworkConditionsResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}
		return connect.NewResponse(response), nil
	}

	response := &nettestlabv1.ResetNetworkConditionsResponse{
		Success: true,
	}
	return connect.NewResponse(response), nil
}

// GetNetworkConditions returns current network conditions for an interface
func (s *NetworkControlService) GetNetworkConditions(ctx context.Context, req *connect.Request[nettestlabv1.GetNetworkConditionsRequest]) (*connect.Response[nettestlabv1.GetNetworkConditionsResponse], error) {
	// Validate request
	if req.Msg.Interface == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("interface name is required"))
	}

	// Get conditions
	conditions, active := s.controller.GetConditions(req.Msg.Interface)
	if !active {
		// Return empty conditions if none are active
		conditions = &nettestlabv1.NetworkConditions{}
	}

	response := &nettestlabv1.GetNetworkConditionsResponse{
		Conditions:  conditions,
		Active:      active,
		LastApplied: timestamppb.Now(), // For now, use current time
	}
	return connect.NewResponse(response), nil
}

// GetSystemStatus returns system status and available interfaces
func (s *NetworkControlService) GetSystemStatus(ctx context.Context, req *connect.Request[nettestlabv1.GetSystemStatusRequest]) (*connect.Response[nettestlabv1.GetSystemStatusResponse], error) {
	// Refresh interface information
	if err := s.controller.RefreshInterfaces(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to refresh interfaces: %v", err))
	}

	// Get interfaces
	interfaceInfos := s.controller.GetInterfaces()
	var interfaces []*nettestlabv1.NetworkInterface

	for _, info := range interfaceInfos {
		// Check if interface has active conditions
		_, hasConditions := s.controller.GetConditions(info.Name)

		// Get applied profile name
		appliedProfile, _ := s.controller.GetAppliedProfile(info.Name)

		interfaces = append(interfaces, &nettestlabv1.NetworkInterface{
			Name:           info.Name,
			Type:           info.Type,
			IsUp:           info.IsUp,
			IpAddresses:    info.IPAddresses,
			HasConditions:  hasConditions,
			AppliedProfile: appliedProfile,
		})
	}

	response := &nettestlabv1.GetSystemStatusResponse{
		Interfaces: interfaces,
		Version:    "1.0.0",
		Load: &nettestlabv1.SystemLoad{
			CpuUsage:          10.0, // Placeholder
			MemoryUsage:       30.0, // Placeholder
			ActiveConnections: 1,    // Placeholder
		},
	}
	return connect.NewResponse(response), nil
}
