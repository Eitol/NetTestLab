package server

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	nettestlabv1 "github.com/Eitol/NetTestLab/api/nettestlab/v1"
	"github.com/Eitol/NetTestLab/internal/device"
	"github.com/Eitol/NetTestLab/internal/target"
)

// TrafficCaptureService implements the Connect traffic capture service
type TrafficCaptureService struct {
	deviceManager *device.Manager
	targetManager *target.Manager
}

// NewTrafficCaptureService creates a new traffic capture service
func NewTrafficCaptureService(dataDir string) (*TrafficCaptureService, error) {
	deviceManager, err := device.NewManager(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create device manager: %w", err)
	}

	targetManager, err := target.NewManager(dataDir)
	if err != nil {
		deviceManager.Close()
		return nil, fmt.Errorf("failed to create target manager: %w", err)
	}

	return &TrafficCaptureService{
		deviceManager: deviceManager,
		targetManager: targetManager,
	}, nil
}

// Close closes the service and its resources
func (s *TrafficCaptureService) Close() error {
	var err1, err2 error
	if s.deviceManager != nil {
		err1 = s.deviceManager.Close()
	}
	if s.targetManager != nil {
		err2 = s.targetManager.Close()
	}
	
	// Return first error encountered
	if err1 != nil {
		return err1
	}
	return err2
}

// ListDevices returns a list of devices based on filter criteria
func (s *TrafficCaptureService) ListDevices(ctx context.Context, req *connect.Request[nettestlabv1.ListDevicesRequest]) (*connect.Response[nettestlabv1.ListDevicesResponse], error) {
	// Set default page size if not specified
	pageSize := req.Msg.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 50 // Default page size
	}

	// Get devices from manager
	devices, nextPageToken, totalCount, err := s.deviceManager.ListDevices(
		req.Msg.Filter,
		int(pageSize),
		req.Msg.PageToken,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list devices: %w", err))
	}

	return connect.NewResponse(&nettestlabv1.ListDevicesResponse{
		Devices:       devices,
		NextPageToken: nextPageToken,
		TotalCount:    int32(totalCount),
	}), nil
}

// RegisterDevice registers a new device or updates an existing one
func (s *TrafficCaptureService) RegisterDevice(ctx context.Context, req *connect.Request[nettestlabv1.RegisterDeviceRequest]) (*connect.Response[nettestlabv1.RegisterDeviceResponse], error) {
	if req.Msg.MacAddress == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("MAC address is required"))
	}

	// Register device
	device, created, err := s.deviceManager.RegisterDevice(
		req.Msg.MacAddress,
		req.Msg.DeviceName,
		req.Msg.DeviceModel,
		req.Msg.OsVersion,
		req.Msg.AppVersion,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to register device: %w", err))
	}

	return connect.NewResponse(&nettestlabv1.RegisterDeviceResponse{
		Device:  device,
		Created: created,
	}), nil
}

// UpdateDevice updates device information
func (s *TrafficCaptureService) UpdateDevice(ctx context.Context, req *connect.Request[nettestlabv1.UpdateDeviceRequest]) (*connect.Response[nettestlabv1.UpdateDeviceResponse], error) {
	if req.Msg.DeviceId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device ID is required"))
	}

	// Update device
	device, err := s.deviceManager.UpdateDevice(
		req.Msg.DeviceId,
		req.Msg.DeviceName,
		req.Msg.DeviceModel,
		req.Msg.OsVersion,
		req.Msg.AppVersion,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update device: %w", err))
	}

	return connect.NewResponse(&nettestlabv1.UpdateDeviceResponse{
		Device: device,
	}), nil
}

// DeleteDevice removes a device from the system
func (s *TrafficCaptureService) DeleteDevice(ctx context.Context, req *connect.Request[nettestlabv1.DeleteDeviceRequest]) (*connect.Response[nettestlabv1.DeleteDeviceResponse], error) {
	if req.Msg.DeviceId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device ID is required"))
	}

	// Delete device
	err := s.deviceManager.DeleteDevice(req.Msg.DeviceId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete device: %w", err))
	}

	return connect.NewResponse(&nettestlabv1.DeleteDeviceResponse{
		Success: true,
		Message: "Device deleted successfully",
	}), nil
}

// === URL TARGET MANAGEMENT APIs ===

// CreateUrlTarget creates a new URL target
func (s *TrafficCaptureService) CreateUrlTarget(ctx context.Context, req *connect.Request[nettestlabv1.CreateUrlTargetRequest]) (*connect.Response[nettestlabv1.CreateUrlTargetResponse], error) {
	if req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("target name is required"))
	}
	if req.Msg.HostRegex == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("host regex is required"))
	}

	// Create target
	target, created, err := s.targetManager.CreateTarget(
		req.Msg.Name,
		req.Msg.Description,
		req.Msg.HostRegex,
		req.Msg.Ports,
		req.Msg.ProtocolFilter,
		req.Msg.Enabled,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create target: %w", err))
	}

	return connect.NewResponse(&nettestlabv1.CreateUrlTargetResponse{
		Target:  target,
		Created: created,
	}), nil
}

// ListUrlTargets returns a list of URL targets
func (s *TrafficCaptureService) ListUrlTargets(ctx context.Context, req *connect.Request[nettestlabv1.ListUrlTargetsRequest]) (*connect.Response[nettestlabv1.ListUrlTargetsResponse], error) {
	// Set default page size if not specified
	pageSize := req.Msg.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 50 // Default page size
	}

	// Get targets from manager
	targets, nextPageToken, totalCount, err := s.targetManager.ListTargets(
		req.Msg.EnabledOnly,
		int(pageSize),
		req.Msg.PageToken,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list targets: %w", err))
	}

	return connect.NewResponse(&nettestlabv1.ListUrlTargetsResponse{
		Targets:       targets,
		NextPageToken: nextPageToken,
		TotalCount:    int32(totalCount),
	}), nil
}

// UpdateUrlTarget updates an existing URL target
func (s *TrafficCaptureService) UpdateUrlTarget(ctx context.Context, req *connect.Request[nettestlabv1.UpdateUrlTargetRequest]) (*connect.Response[nettestlabv1.UpdateUrlTargetResponse], error) {
	if req.Msg.TargetId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("target ID is required"))
	}

	// Update target
	target, err := s.targetManager.UpdateTarget(
		req.Msg.TargetId,
		req.Msg.Name,
		req.Msg.Description,
		req.Msg.HostRegex,
		req.Msg.Ports,
		req.Msg.ProtocolFilter,
		req.Msg.Enabled,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update target: %w", err))
	}

	return connect.NewResponse(&nettestlabv1.UpdateUrlTargetResponse{
		Target: target,
	}), nil
}

// DeleteUrlTarget removes a URL target
func (s *TrafficCaptureService) DeleteUrlTarget(ctx context.Context, req *connect.Request[nettestlabv1.DeleteUrlTargetRequest]) (*connect.Response[nettestlabv1.DeleteUrlTargetResponse], error) {
	if req.Msg.TargetId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("target ID is required"))
	}

	// Delete target
	err := s.targetManager.DeleteTarget(req.Msg.TargetId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete target: %w", err))
	}

	return connect.NewResponse(&nettestlabv1.DeleteUrlTargetResponse{
		Success: true,
		Message: "Target deleted successfully",
	}), nil
}
