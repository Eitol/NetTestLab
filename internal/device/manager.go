package device

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	nettestlabv1 "github.com/Eitol/NetTestLab/api/nettestlab/v1"
)

// Manager handles device operations and coordinates discovery with storage
type Manager struct {
	db        *Database
	discovery *Discovery
}

// NewManager creates a new device manager
func NewManager(dataDir string) (*Manager, error) {
	db, err := NewDatabase(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	discovery := NewDiscovery()

	manager := &Manager{
		db:        db,
		discovery: discovery,
	}

	// Start periodic device discovery
	manager.StartPeriodicDiscovery(30 * time.Second)

	return manager, nil
}

// Close closes the device manager and its resources
func (m *Manager) Close() error {
	return m.db.Close()
}

// StartPeriodicDiscovery starts automatic device discovery
func (m *Manager) StartPeriodicDiscovery(interval time.Duration) {
	m.discovery.StartPeriodicScan(interval, func(detected []*DetectedDevice) {
		err := m.updateDetectedDevices(detected)
		if err != nil {
			fmt.Printf("Error updating detected devices: %v\n", err)
		}
	})
}

// updateDetectedDevices updates the database with newly detected devices
func (m *Manager) updateDetectedDevices(detected []*DetectedDevice) error {
	// First, mark all non-registered devices as disconnected
	err := m.markNonRegisteredDevicesAsDisconnected()
	if err != nil {
		fmt.Printf("Warning: failed to mark devices as disconnected: %v\n", err)
	}

	// Then update/create detected devices as connected
	for _, device := range detected {
		err := m.mergeDetectedDevice(device)
		if err != nil {
			fmt.Printf("Error merging detected device %s: %v\n", device.MacAddress, err)
		}
	}
	return nil
}

// markNonRegisteredDevicesAsDisconnected marks all auto-discovered devices as disconnected
// before updating with current detections
func (m *Manager) markNonRegisteredDevicesAsDisconnected() error {
	// Get all non-registered devices
	nonRegisteredDevices, err := m.db.GetNonRegisteredDevices()
	if err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)

	for _, device := range nonRegisteredDevices {
		device.ConnectionStatus = int(nettestlabv1.DeviceConnectionStatus_DEVICE_CONNECTION_STATUS_DISCONNECTED)
		device.LastSeen = &now

		err := m.db.UpdateDevice(device)
		if err != nil {
			fmt.Printf("Warning: failed to mark device %s as disconnected: %v\n", device.MacAddress, err)
		}
	}

	return nil
}

// mergeDetectedDevice merges a detected device with existing database records
func (m *Manager) mergeDetectedDevice(detected *DetectedDevice) error {
	// Check if device already exists by MAC address
	existing, err := m.db.GetDeviceByMAC(detected.MacAddress)
	if err != nil {
		return err
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	if existing == nil {
		// Create new device record
		deviceRow := &DeviceRow{
			ID:               uuid.New().String(),
			MacAddress:       detected.MacAddress,
			IPAddress:        &detected.IPAddress,
			Hostname:         &detected.Hostname,
			ConnectionStatus: int(nettestlabv1.DeviceConnectionStatus_DEVICE_CONNECTION_STATUS_CONNECTED),
			Registered:       false,
			FirstSeen:        &nowStr,
			LastSeen:         &nowStr,
			Vendor:           &detected.Vendor,
		}

		return m.db.InsertDevice(deviceRow)
	} else {
		// Update existing device
		// Update IP if changed
		if detected.IPAddress != "" && (existing.IPAddress == nil || *existing.IPAddress != detected.IPAddress) {
			// Add previous IP to history
			err := m.addIPToHistory(existing, detected.IPAddress)
			if err != nil {
				fmt.Printf("Warning: failed to update IP history: %v\n", err)
			}
			existing.IPAddress = &detected.IPAddress
		}

		// Update hostname if available and different
		if detected.Hostname != "" && (existing.Hostname == nil || *existing.Hostname != detected.Hostname) {
			existing.Hostname = &detected.Hostname
		}

		// Update vendor if available and different
		if detected.Vendor != "" && (existing.Vendor == nil || *existing.Vendor != detected.Vendor) {
			existing.Vendor = &detected.Vendor
		}

		// Update connection status and last seen
		existing.ConnectionStatus = int(nettestlabv1.DeviceConnectionStatus_DEVICE_CONNECTION_STATUS_CONNECTED)
		nowStr := now.Format(time.RFC3339)
		existing.LastSeen = &nowStr

		return m.db.UpdateDevice(existing)
	}
}

// addIPToHistory adds the current IP to the previous IPs list
func (m *Manager) addIPToHistory(device *DeviceRow, newIP string) error {
	var previousIPs []string

	if device.PreviousIPs != nil && *device.PreviousIPs != "" {
		err := json.Unmarshal([]byte(*device.PreviousIPs), &previousIPs)
		if err != nil {
			// If unmarshal fails, start with empty list
			previousIPs = []string{}
		}
	}

	// Add current IP to history if it's different from the new one
	if device.IPAddress != nil && *device.IPAddress != newIP {
		// Check if IP is already in history
		found := false
		for _, ip := range previousIPs {
			if ip == *device.IPAddress {
				found = true
				break
			}
		}

		if !found {
			previousIPs = append(previousIPs, *device.IPAddress)

			// Keep only last 10 IPs
			if len(previousIPs) > 10 {
				previousIPs = previousIPs[len(previousIPs)-10:]
			}
		}
	}

	// Update the JSON field
	jsonData, err := json.Marshal(previousIPs)
	if err != nil {
		return err
	}

	jsonStr := string(jsonData)
	device.PreviousIPs = &jsonStr

	return nil
}

// ListDevices returns devices based on filter criteria
func (m *Manager) ListDevices(filter nettestlabv1.DeviceFilter, pageSize int, pageToken string) ([]*nettestlabv1.Device, string, int, error) {
	// Convert proto filter to internal filter
	dbFilter := m.convertFilterToDBFilter(filter)

	// Calculate offset from page token
	offset := 0
	if pageToken != "" {
		// Simple integer-based pagination
		// In production, you might want more sophisticated pagination
		if _, err := fmt.Sscanf(pageToken, "%d", &offset); err != nil {
			offset = 0
		}
	}

	// Get devices from database
	deviceRows, err := m.db.ListDevices(dbFilter, pageSize, offset)
	if err != nil {
		return nil, "", 0, err
	}

	// Convert to protobuf devices
	var devices []*nettestlabv1.Device
	for _, row := range deviceRows {
		device := m.convertRowToProtoDevice(row)
		devices = append(devices, device)
	}

	// Calculate next page token
	nextPageToken := ""
	if len(devices) == pageSize {
		nextPageToken = fmt.Sprintf("%d", offset+pageSize)
	}

	// Get total count
	totalCount, err := m.db.CountDevices(dbFilter)
	if err != nil {
		return nil, "", 0, err
	}

	return devices, nextPageToken, totalCount, nil
}

// RegisterDevice registers a device manually
func (m *Manager) RegisterDevice(macAddress, deviceName, deviceModel, osVersion, appVersion string) (*nettestlabv1.Device, bool, error) {
	// Normalize MAC address
	macAddress = strings.ToLower(macAddress)

	// Validate MAC address format
	if !m.discovery.isValidMACAddress(macAddress) {
		return nil, false, fmt.Errorf("invalid MAC address format")
	}

	// Check if device already exists
	existing, err := m.db.GetDeviceByMAC(macAddress)
	if err != nil {
		return nil, false, err
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)
	created := false

	if existing == nil {
		// Create new device
		deviceRow := &DeviceRow{
			ID:           uuid.New().String(),
			MacAddress:   macAddress,
			DeviceName:   &deviceName,
			DeviceModel:  &deviceModel,
			OSVersion:    &osVersion,
			AppVersion:   &appVersion,
			Registered:   true,
			RegisteredAt: &nowStr,
			FirstSeen:    &nowStr,
			LastSeen:     &nowStr,
		}

		// Try to get vendor info
		vendor := m.discovery.vendorLookup.LookupVendor(macAddress)
		if vendor != "Unknown" {
			deviceRow.Vendor = &vendor
		}

		err = m.db.InsertDevice(deviceRow)
		if err != nil {
			return nil, false, err
		}

		created = true
		existing = deviceRow
	} else {
		// Update existing device
		if deviceName != "" {
			existing.DeviceName = &deviceName
		}
		if deviceModel != "" {
			existing.DeviceModel = &deviceModel
		}
		if osVersion != "" {
			existing.OSVersion = &osVersion
		}
		if appVersion != "" {
			existing.AppVersion = &appVersion
		}

		if !existing.Registered {
			existing.Registered = true
			existing.RegisteredAt = &nowStr
		}

		err = m.db.UpdateDevice(existing)
		if err != nil {
			return nil, false, err
		}
	}

	// Convert to proto device
	device := m.convertRowToProtoDevice(existing)
	return device, created, nil
}

// UpdateDevice updates device information
func (m *Manager) UpdateDevice(deviceID, deviceName, deviceModel, osVersion, appVersion string) (*nettestlabv1.Device, error) {
	// Get existing device
	existing, err := m.db.GetDeviceByID(deviceID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("device not found")
	}

	// Update fields
	if deviceName != "" {
		existing.DeviceName = &deviceName
	}
	if deviceModel != "" {
		existing.DeviceModel = &deviceModel
	}
	if osVersion != "" {
		existing.OSVersion = &osVersion
	}
	if appVersion != "" {
		existing.AppVersion = &appVersion
	}

	err = m.db.UpdateDevice(existing)
	if err != nil {
		return nil, err
	}

	// Convert to proto device
	device := m.convertRowToProtoDevice(existing)
	return device, nil
}

// DeleteDevice removes a device from the database
func (m *Manager) DeleteDevice(deviceID string) error {
	return m.db.DeleteDevice(deviceID)
}

// convertFilterToDBFilter converts protobuf filter to database filter
func (m *Manager) convertFilterToDBFilter(filter nettestlabv1.DeviceFilter) DeviceFilter {
	switch filter {
	case nettestlabv1.DeviceFilter_DEVICE_FILTER_CONNECTED:
		return DeviceFilterConnected
	case nettestlabv1.DeviceFilter_DEVICE_FILTER_REGISTERED:
		return DeviceFilterRegistered
	case nettestlabv1.DeviceFilter_DEVICE_FILTER_CONNECTED_REGISTERED:
		return DeviceFilterConnectedRegistered
	case nettestlabv1.DeviceFilter_DEVICE_FILTER_CONNECTED_UNREGISTERED:
		return DeviceFilterConnectedUnregistered
	default:
		return DeviceFilterAll
	}
}

// convertRowToProtoDevice converts database row to protobuf device
func (m *Manager) convertRowToProtoDevice(row *DeviceRow) *nettestlabv1.Device {
	device := &nettestlabv1.Device{
		Id:               row.ID,
		MacAddress:       row.MacAddress,
		ConnectionStatus: nettestlabv1.DeviceConnectionStatus(row.ConnectionStatus),
		Registered:       row.Registered,
	}

	// Optional fields
	if row.IPAddress != nil {
		device.IpAddress = *row.IPAddress
	}
	if row.Hostname != nil {
		device.Hostname = *row.Hostname
	}
	if row.DeviceName != nil {
		device.DeviceName = *row.DeviceName
	}
	if row.DeviceModel != nil {
		device.DeviceModel = *row.DeviceModel
	}
	if row.OSVersion != nil {
		device.OsVersion = *row.OSVersion
	}
	if row.AppVersion != nil {
		device.AppVersion = *row.AppVersion
	}
	if row.Vendor != nil {
		device.Vendor = *row.Vendor
	}

	// Parse timestamps
	if row.FirstSeen != nil {
		if t, err := time.Parse(time.RFC3339, *row.FirstSeen); err == nil {
			device.FirstSeen = timestamppb.New(t)
		}
	}
	if row.LastSeen != nil {
		if t, err := time.Parse(time.RFC3339, *row.LastSeen); err == nil {
			device.LastSeen = timestamppb.New(t)
		}
	}
	if row.RegisteredAt != nil {
		if t, err := time.Parse(time.RFC3339, *row.RegisteredAt); err == nil {
			device.RegisteredAt = timestamppb.New(t)
		}
	}

	// Parse previous IPs
	if row.PreviousIPs != nil && *row.PreviousIPs != "" {
		var previousIPs []string
		if err := json.Unmarshal([]byte(*row.PreviousIPs), &previousIPs); err == nil {
			device.PreviousIps = previousIPs
		}
	}

	return device
}
