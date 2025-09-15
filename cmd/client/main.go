package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Eitol/NetTestLab/api/nettestlab/v1"
)

func main() {
	serverAddr := flag.String("server", "192.168.1.4:8080", "Server address")
	useTLS := flag.Bool("tls", false, "Use TLS connection")
	quickTest := flag.Bool("test", false, "Run quick connectivity test only")
	flag.Parse()

	fmt.Printf("🔗 Connecting to NetTestLab server at %s\n", *serverAddr)

	// Setup connection options
	var opts []grpc.DialOption
	if *useTLS {
		config := &tls.Config{
			ServerName: "192.168.1.4", // For testing
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Connect to server
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create clients
	networkClient := pb.NewNetworkControlServiceClient(conn)
	profileClient := pb.NewProfileServiceClient(conn)
	monitoringClient := pb.NewMonitoringServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("✅ Connected successfully!")

	// If quick test mode, just do connectivity check and exit
	if *quickTest {
		fmt.Println("\n🚀 Quick Test Mode: Basic connectivity check")
		_, err := networkClient.GetSystemStatus(ctx, &pb.GetSystemStatusRequest{})
		if err != nil {
			log.Fatalf("❌ Quick test failed: %v", err)
		}
		fmt.Println("✅ Quick test passed: Server is responding")
		return
	}

	// Test 1: Get system status (shows interfaces)
	fmt.Println("\n📋 Testing: Get System Status")
	status, err := networkClient.GetSystemStatus(ctx, &pb.GetSystemStatusRequest{})
	if err != nil {
		log.Printf("Failed to get system status: %v", err)
	} else {
		fmt.Printf("System Version: %s\n", status.Version)
		fmt.Printf("Uptime: %s\n", status.Uptime.AsDuration())
		fmt.Printf("Found %d interfaces:\n", len(status.Interfaces))
		for _, iface := range status.Interfaces {
			fmt.Printf("  - %s (%s) - Up: %v, Has Conditions: %v\n",
				iface.Name, iface.Type, iface.IsUp, iface.HasConditions)
		}
		if status.Load != nil {
			fmt.Printf("System Load - CPU: %.1f%%, Memory: %.1f%%\n",
				status.Load.CpuUsage, status.Load.MemoryUsage)
		}
	}

	// Test 2: List available profiles
	fmt.Println("\n📱 Testing: List Network Profiles")
	profiles, err := profileClient.ListProfiles(ctx, &pb.ListProfilesRequest{})
	if err != nil {
		log.Printf("Failed to list profiles: %v", err)
	} else {
		fmt.Printf("Found %d profiles:\n", len(profiles.Profiles))
		for _, profile := range profiles.Profiles {
			fmt.Printf("  - %s: %s\n", profile.Name, profile.Description)
		}
	}

	// Test 3: Get health status
	fmt.Println("\n📊 Testing: Health Check")
	health, err := monitoringClient.GetHealth(ctx, &pb.GetHealthRequest{})
	if err != nil {
		log.Printf("Failed to get health: %v", err)
	} else {
		fmt.Printf("Health Status: %s\n", health.Status)
		fmt.Printf("Version: %s\n", health.Version)
		fmt.Printf("Uptime: %s\n", health.Uptime.AsDuration())
		if len(health.Components) > 0 {
			fmt.Printf("Component Health:\n")
			for _, comp := range health.Components {
				fmt.Printf("  - %s: %s\n", comp.Name, comp.Status)
			}
		}
	}

	// Test 4: Apply a network condition (2G simulation on WiFi interface)
	fmt.Println("\n📱 Testing: Apply Mobile Network Condition on WiFi Interface")

	// Apply custom network conditions to WiFi interface used by mobile devices
	applyReq := &pb.ApplyNetworkConditionsRequest{
		Interface: "wl1-ap0", // WiFi interface for mobile device connections
		Conditions: &pb.NetworkConditions{
			Latency: &pb.LatencyConfig{
				DelayMs: 250,
				Enabled: true,
			},
			PacketLoss: &pb.PacketLossConfig{
				Percentage: 5.0, // 5% packet loss
				Enabled:    true,
			},
			Bandwidth: &pb.BandwidthConfig{
				DownloadBps: 56000, // 56 kbps (2G speed)
				UploadBps:   28000, // 28 kbps (2G speed)
				Enabled:     true,
			},
			Jitter: &pb.JitterConfig{
				VariationMs: 50,
				Enabled:     true,
			},
		},
	}

	applyResp, err := networkClient.ApplyNetworkConditions(ctx, applyReq)
	if err != nil {
		log.Printf("Failed to apply conditions: %v", err)
	} else if applyResp.Success {
		fmt.Println("Network conditions applied successfully!")
		if applyResp.AppliedConditions != nil {
			fmt.Printf("Applied conditions:\n")
			if applyResp.AppliedConditions.Latency != nil && applyResp.AppliedConditions.Latency.Enabled {
				fmt.Printf("  - Latency: %dms\n", applyResp.AppliedConditions.Latency.DelayMs)
			}
			if applyResp.AppliedConditions.PacketLoss != nil && applyResp.AppliedConditions.PacketLoss.Enabled {
				fmt.Printf("  - Packet Loss: %.1f%%\n", applyResp.AppliedConditions.PacketLoss.Percentage)
			}
			if applyResp.AppliedConditions.Bandwidth != nil && applyResp.AppliedConditions.Bandwidth.Enabled {
				fmt.Printf("  - Bandwidth: %d/%d bps\n",
					applyResp.AppliedConditions.Bandwidth.DownloadBps,
					applyResp.AppliedConditions.Bandwidth.UploadBps)
			}
		}

		// Wait a moment, then reset conditions
		fmt.Println("\n🧹 Testing: Reset WiFi Network Conditions")
		time.Sleep(3 * time.Second)

		resetReq := &pb.ResetNetworkConditionsRequest{
			Interface: "wl1-ap0",
		}

		resetResp, err := networkClient.ResetNetworkConditions(ctx, resetReq)
		if err != nil {
			log.Printf("Failed to reset conditions: %v", err)
		} else if resetResp.Success {
			fmt.Println("Conditions reset successfully!")
		} else {
			fmt.Printf("Reset failed: %s\n", resetResp.ErrorMessage)
		}
	} else {
		fmt.Printf("Failed to apply conditions: %s\n", applyResp.ErrorMessage)
	}

	fmt.Println("\n🎉 All tests completed!")
}
