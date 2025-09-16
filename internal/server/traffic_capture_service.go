package server

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	nettestlabv1 "github.com/Eitol/NetTestLab/api/nettestlab/v1"
	"github.com/Eitol/NetTestLab/internal/device"
)

// TrafficCaptureService implements the Connect traffic capture service
type TrafficCaptureService struct {
	deviceManager *device.Manager
}

// NewTrafficCaptureService creates a new traffic capture service
func NewTrafficCaptureService(dataDir string) (*TrafficCaptureService, error) {
	deviceManager, err := device.NewManager(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create device manager: %w", err)
	}

	return &TrafficCaptureService{
		deviceManager: deviceManager,
	}, nil
}

// Close closes the service and its resources
func (s *TrafficCaptureService) Close() error {
	if s.deviceManager != nil {
		return s.deviceManager.Close()
	}
	return nil
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