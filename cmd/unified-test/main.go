package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Eitol/NetTestLab/api/nettestlab/v1"
)

func main() {
	serverAddr := flag.String("server", "192.168.1.4:8080", "Server address (can be localhost:8080 for local testing)")
	quickTest := flag.Bool("quick", false, "Run only quick connectivity tests")
	verbose := flag.Bool("v", false, "Verbose output")
	flag.Parse()

	fmt.Printf("🚀 NetTestLab Unified Test Suite\n")
	fmt.Printf("================================\n")
	fmt.Printf("🔗 Server: %s\n", *serverAddr)
	fmt.Printf("⚡ Mode: %s\n", func() string {
		if *quickTest {
			return "Quick Test"
		}
		return "Full Test Suite"
	}())
	fmt.Println()

	// Setup connection
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Connect to server
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("❌ Failed to connect to %s: %v", *serverAddr, err)
	}
	defer conn.Close()

	// Create clients
	networkClient := pb.NewNetworkControlServiceClient(conn)
	profileClient := pb.NewProfileServiceClient(conn)
	monitoringClient := pb.NewMonitoringServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	fmt.Println("✅ Connected successfully!")

	// Test counter
	testCount := 0
	passedTests := 0

	// Helper function to run a test
	runTest := func(name string, testFunc func() error) {
		testCount++
		fmt.Printf("\n📋 Test %d: %s\n", testCount, name)
		if err := testFunc(); err != nil {
			fmt.Printf("❌ FAILED: %v\n", err)
			if !*quickTest {
				log.Printf("Test failed but continuing...")
			}
		} else {
			fmt.Printf("✅ PASSED\n")
			passedTests++
		}
	}

	// Test 1: Basic Connectivity & System Status
	runTest("Basic Connectivity & System Status", func() error {
		status, err := networkClient.GetSystemStatus(ctx, &pb.GetSystemStatusRequest{})
		if err != nil {
			return fmt.Errorf("failed to get system status: %v", err)
		}

		if *verbose {
			fmt.Printf("   📊 System Version: %s\n", status.Version)
			fmt.Printf("   ⏰ Uptime: %s\n", status.Uptime.AsDuration())
			fmt.Printf("   🌐 Interfaces found: %d\n", len(status.Interfaces))

			for _, iface := range status.Interfaces {
				fmt.Printf("     - %s (%s) - Up: %v\n", iface.Name, iface.Type, iface.IsUp)
			}
		}

		if len(status.Interfaces) == 0 {
			return fmt.Errorf("no network interfaces found")
		}

		return nil
	})

	// Test 2: Health Check
	runTest("Health Check", func() error {
		health, err := monitoringClient.GetHealth(ctx, &pb.GetHealthRequest{})
		if err != nil {
			return fmt.Errorf("failed to get health status: %v", err)
		}

		if *verbose {
			fmt.Printf("   💚 Status: %s\n", health.Status)
			if health.Uptime != nil {
				fmt.Printf("   ⏰ Uptime: %s\n", health.Uptime.AsDuration())
			}
		}

		return nil
	})

	// Test 3: Profile Management
	runTest("Profile Management", func() error {
		profiles, err := profileClient.ListProfiles(ctx, &pb.ListProfilesRequest{})
		if err != nil {
			return fmt.Errorf("failed to list profiles: %v", err)
		}

		if *verbose {
			fmt.Printf("   📱 Available profiles: %d\n", len(profiles.Profiles))
			for _, profile := range profiles.Profiles {
				fmt.Printf("     - %s: %s\n", profile.Name, profile.Description)
			}
		}

		if len(profiles.Profiles) == 0 {
			return fmt.Errorf("no profiles available")
		}

		return nil
	})

	if *quickTest {
		fmt.Printf("\n🎯 Quick test completed: %d/%d tests passed\n", passedTests, testCount)
		if passedTests == testCount {
			fmt.Println("✅ All quick tests passed!")
			os.Exit(0)
		} else {
			fmt.Println("❌ Some quick tests failed!")
			os.Exit(1)
		}
	}

	// Test 4: Simple Packet Loss Application
	runTest("Simple Packet Loss on WiFi", func() error {
		// Apply simple packet loss
		applyReq := &pb.ApplyNetworkConditionsRequest{
			Interface: "wifi", // Use auto-discovery
			Conditions: &pb.NetworkConditions{
				PacketLoss: &pb.PacketLossConfig{
					Percentage: 2.0,
					Enabled:    true,
				},
			},
		}

		applyResp, err := networkClient.ApplyNetworkConditions(ctx, applyReq)
		if err != nil {
			return fmt.Errorf("failed to apply packet loss: %v", err)
		}

		if !applyResp.Success {
			return fmt.Errorf("apply failed: %s", applyResp.ErrorMessage)
		}

		if *verbose {
			fmt.Printf("   📉 Applied: 2%% packet loss to WiFi\n")
		}

		// Wait briefly
		time.Sleep(1 * time.Second)

		// Reset conditions
		resetReq := &pb.ResetNetworkConditionsRequest{
			Interface: "wifi",
		}

		resetResp, err := networkClient.ResetNetworkConditions(ctx, resetReq)
		if err != nil {
			return fmt.Errorf("failed to reset conditions: %v", err)
		}

		if !resetResp.Success {
			return fmt.Errorf("reset failed: %s", resetResp.ErrorMessage)
		}

		if *verbose {
			fmt.Printf("   🧹 Conditions reset\n")
		}

		return nil
	})

	// Test 5: Mobile Network Scenarios
	runTest("Mobile Network Scenarios", func() error {
		scenarios := []struct {
			name       string
			conditions *pb.NetworkConditions
		}{
			{
				name: "2G Network",
				conditions: &pb.NetworkConditions{
					Latency: &pb.LatencyConfig{
						DelayMs: 300,
						Enabled: true,
					},
					PacketLoss: &pb.PacketLossConfig{
						Percentage: 5.0,
						Enabled:    true,
					},
				},
			},
			{
				name: "3G Network",
				conditions: &pb.NetworkConditions{
					Latency: &pb.LatencyConfig{
						DelayMs: 150,
						Enabled: true,
					},
					PacketLoss: &pb.PacketLossConfig{
						Percentage: 2.0,
						Enabled:    true,
					},
				},
			},
			{
				name: "4G/LTE Network",
				conditions: &pb.NetworkConditions{
					Latency: &pb.LatencyConfig{
						DelayMs: 50,
						Enabled: true,
					},
					PacketLoss: &pb.PacketLossConfig{
						Percentage: 0.5,
						Enabled:    true,
					},
				},
			},
		}

		for i, scenario := range scenarios {
			if *verbose {
				fmt.Printf("   📱 Scenario %d: %s\n", i+1, scenario.name)
			}

			// Apply scenario
			applyReq := &pb.ApplyNetworkConditionsRequest{
				Interface:  "wifi",
				Conditions: scenario.conditions,
			}

			applyResp, err := networkClient.ApplyNetworkConditions(ctx, applyReq)
			if err != nil {
				return fmt.Errorf("failed to apply %s: %v", scenario.name, err)
			}

			if !applyResp.Success {
				return fmt.Errorf("%s apply failed: %s", scenario.name, applyResp.ErrorMessage)
			}

			if *verbose {
				if applyResp.AppliedConditions != nil {
					cond := applyResp.AppliedConditions
					if cond.Latency != nil && cond.Latency.Enabled {
						fmt.Printf("     ⏱️  Latency: %dms\n", cond.Latency.DelayMs)
					}
					if cond.PacketLoss != nil && cond.PacketLoss.Enabled {
						fmt.Printf("     📉 Packet Loss: %.1f%%\n", cond.PacketLoss.Percentage)
					}
				}
			}

			// Brief wait
			time.Sleep(500 * time.Millisecond)

			// Reset after each scenario
			resetReq := &pb.ResetNetworkConditionsRequest{
				Interface: "wifi",
			}

			resetResp, err := networkClient.ResetNetworkConditions(ctx, resetReq)
			if err != nil {
				return fmt.Errorf("failed to reset after %s: %v", scenario.name, err)
			}

			if !resetResp.Success {
				return fmt.Errorf("%s reset failed: %s", scenario.name, resetResp.ErrorMessage)
			}
		}

		return nil
	})

	// Test 6: Profile Application
	runTest("Profile Application to WiFi", func() error {
		// Get available profiles first
		profiles, err := profileClient.ListProfiles(ctx, &pb.ListProfilesRequest{})
		if err != nil {
			return fmt.Errorf("failed to list profiles: %v", err)
		}

		if len(profiles.Profiles) == 0 {
			return fmt.Errorf("no profiles available for testing")
		}

		// Test with first available profile
		testProfile := profiles.Profiles[0]

		if *verbose {
			fmt.Printf("   📱 Testing profile: %s\n", testProfile.Name)
		}

		// Get profile details first
		profileReq := &pb.GetProfileRequest{
			Name: testProfile.Name,
		}

		profileResp, err := profileClient.GetProfile(ctx, profileReq)
		if err != nil {
			return fmt.Errorf("failed to get profile %s: %v", testProfile.Name, err)
		}

		// Apply profile conditions
		applyReq := &pb.ApplyNetworkConditionsRequest{
			Interface:  "wifi",
			Conditions: profileResp.Profile.Conditions,
		}

		applyResp, err := networkClient.ApplyNetworkConditions(ctx, applyReq)
		if err != nil {
			return fmt.Errorf("failed to apply profile %s: %v", testProfile.Name, err)
		}

		if !applyResp.Success {
			return fmt.Errorf("profile %s apply failed: %s", testProfile.Name, applyResp.ErrorMessage)
		}

		if *verbose {
			fmt.Printf("   ✅ Profile %s applied\n", testProfile.Name)
		}

		// Wait briefly
		time.Sleep(1 * time.Second)

		// Reset
		resetReq := &pb.ResetNetworkConditionsRequest{
			Interface: "wifi",
		}

		resetResp, err := networkClient.ResetNetworkConditions(ctx, resetReq)
		if err != nil {
			return fmt.Errorf("failed to reset profile conditions: %v", err)
		}

		if !resetResp.Success {
			return fmt.Errorf("profile reset failed: %s", resetResp.ErrorMessage)
		}

		return nil
	})

	// Test 7: Auto-discovery with multiple scenarios
	runTest("Auto-discovery WiFi Interface Test", func() error {
		scenarios := []struct {
			name       string
			conditions *pb.NetworkConditions
		}{
			{
				name: "Poor 3G",
				conditions: &pb.NetworkConditions{
					Latency: &pb.LatencyConfig{
						DelayMs: 200,
						Enabled: true,
					},
					PacketLoss: &pb.PacketLossConfig{
						Percentage: 3.0,
						Enabled:    true,
					},
				},
			},
			{
				name: "Good 4G",
				conditions: &pb.NetworkConditions{
					Latency: &pb.LatencyConfig{
						DelayMs: 30,
						Enabled: true,
					},
					PacketLoss: &pb.PacketLossConfig{
						Percentage: 0.5,
						Enabled:    true,
					},
				},
			},
		}

		for _, scenario := range scenarios {
			if *verbose {
				fmt.Printf("   📶 Auto-discovery test: %s\n", scenario.name)
			}

			// Apply using "wifi" keyword for auto-discovery
			applyReq := &pb.ApplyNetworkConditionsRequest{
				Interface:  "wifi",
				Conditions: scenario.conditions,
			}

			applyResp, err := networkClient.ApplyNetworkConditions(ctx, applyReq)
			if err != nil {
				return fmt.Errorf("auto-discovery failed for %s: %v", scenario.name, err)
			}

			if !applyResp.Success {
				return fmt.Errorf("auto-discovery %s failed: %s", scenario.name, applyResp.ErrorMessage)
			}

			// Brief wait
			time.Sleep(500 * time.Millisecond)
		}

		// Final cleanup
		resetReq := &pb.ResetNetworkConditionsRequest{
			Interface: "wifi",
		}

		resetResp, err := networkClient.ResetNetworkConditions(ctx, resetReq)
		if err != nil {
			return fmt.Errorf("final auto-discovery cleanup failed: %v", err)
		}

		if !resetResp.Success {
			return fmt.Errorf("final cleanup failed: %s", resetResp.ErrorMessage)
		}

		return nil
	})

	// Final Summary
	fmt.Printf("\n🎯 Test Summary\n")
	fmt.Printf("===============\n")
	fmt.Printf("✅ Passed: %d/%d tests\n", passedTests, testCount)

	if passedTests == testCount {
		fmt.Printf("🎉 All tests passed! NetTestLab is working correctly.\n")
		fmt.Printf("💡 The system is ready for mobile device testing.\n")
		os.Exit(0)
	} else {
		fmt.Printf("❌ %d tests failed. Please check the system.\n", testCount-passedTests)
		os.Exit(1)
	}
}
