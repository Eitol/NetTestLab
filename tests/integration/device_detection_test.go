package integration

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Eitol/NetTestLab/api/nettestlab/v1"
)

func TestDeviceDetection(t *testing.T) {
	// Router IP from environment or default
	routerIP := "192.168.1.4"

	t.Logf("Testing device detection with router at %s", routerIP)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to gRPC server
	conn, err := grpc.NewClient(routerIP+":8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to router: %v", err)
	}
	defer conn.Close()

	trafficCaptureClient := pb.NewTrafficCaptureServiceClient(conn)

	t.Run("ListDevices", func(t *testing.T) {
		t.Log("Listing detected devices...")

		// Wait a moment for potential device discovery
		time.Sleep(2 * time.Second)

		// List all devices
		listResp, err := trafficCaptureClient.ListDevices(ctx, &pb.ListDevicesRequest{
			Filter:   pb.DeviceFilter_DEVICE_FILTER_ALL,
			PageSize: 100,
		})
		if err != nil {
			t.Fatalf("Failed to list devices: %v", err)
		}

		t.Logf("Found %d devices total", listResp.TotalCount)

		// Check for specific Android device mentioned in conversation
		var androidDevice *pb.Device
		var connectedDevices []*pb.Device

		for _, device := range listResp.Devices {
			t.Logf("Device: MAC=%s, IP=%s, Status=%s, Registered=%v, Hostname=%s",
				device.MacAddress, device.IpAddress,
				device.ConnectionStatus.String(),
				device.Registered,
				device.Hostname)

			if device.ConnectionStatus == pb.DeviceConnectionStatus_DEVICE_CONNECTION_STATUS_CONNECTED {
				connectedDevices = append(connectedDevices, device)
			}

			// Look for the specific Android device mentioned
			if device.MacAddress == "0c:c6:fd:17:f5:74" {
				androidDevice = device
			}
		}

		t.Logf("Found %d connected devices", len(connectedDevices))

		// Verify the Android device from conversation is detected correctly
		if androidDevice != nil {
			t.Logf("✅ Found Android device: MAC=%s, IP=%s, Status=%s",
				androidDevice.MacAddress, androidDevice.IpAddress,
				androidDevice.ConnectionStatus.String())

			if androidDevice.ConnectionStatus != pb.DeviceConnectionStatus_DEVICE_CONNECTION_STATUS_CONNECTED {
				t.Errorf("❌ Android device should be connected but shows as %s",
					androidDevice.ConnectionStatus.String())
			}

			if androidDevice.IpAddress != "192.168.1.86" {
				t.Errorf("❌ Android device IP mismatch: expected 192.168.1.86, got %s",
					androidDevice.IpAddress)
			}
		} else {
			t.Log("⚠️  Android device (0c:c6:fd:17:f5:74) not found - may not be connected")
		}

		// Log some sample connected devices
		for i, device := range connectedDevices {
			if i < 3 { // Show first 3 connected devices
				t.Logf("Connected device %d: %s (%s) - %s",
					i+1, device.MacAddress, device.IpAddress, device.Hostname)
			}
		}
	})

	t.Run("WiFiOnlyDetection", func(t *testing.T) {
		t.Log("Verifying WiFi-only detection (no unwanted ARP devices)...")

		// List devices again
		listResp, err := trafficCaptureClient.ListDevices(ctx, &pb.ListDevicesRequest{
			Filter:   pb.DeviceFilter_DEVICE_FILTER_CONNECTED,
			PageSize: 100,
		})
		if err != nil {
			t.Fatalf("Failed to list connected devices: %v", err)
		}

		// Check for unwanted devices mentioned in conversation
		unwantedIPs := []string{"192.168.1.1", "192.168.1.97", "192.168.1.33"}

		for _, device := range listResp.Devices {
			for _, unwantedIP := range unwantedIPs {
				if device.IpAddress == unwantedIP {
					t.Errorf("❌ Found unwanted device from ARP table: %s (%s) - should be filtered out",
						device.IpAddress, device.MacAddress)
				}
			}
		}

		t.Logf("✅ Verified no unwanted ARP devices in connected list")
	})
}