package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Eitol/NetTestLab/api/nettestlab/v1"
)

func main() {
	serverAddr := flag.String("server", "192.168.1.4:8080", "Server address")
	flag.Parse()

	fmt.Printf("🌐 NetTestLab Auto-WiFi Discovery Test\n")
	fmt.Printf("🔗 Connecting to server at %s\n", *serverAddr)

	// Setup connection
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Connect to server
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create clients
	networkClient := pb.NewNetworkControlServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("✅ Connected successfully!")

	// Test 1: Apply simple conditions using "wifi" as interface name
	fmt.Println("\n🔍 Test 1: Apply simple conditions using 'wifi' interface auto-discovery")

	applyReq := &pb.ApplyNetworkConditionsRequest{
		Interface: "wifi", // Special keyword - should auto-discover WiFi interfaces
		Conditions: &pb.NetworkConditions{
			PacketLoss: &pb.PacketLossConfig{
				Percentage: 2.0,
				Enabled:    true,
			},
		},
	}

	applyResp, err := networkClient.ApplyNetworkConditions(ctx, applyReq)
	if err != nil {
		log.Printf("❌ Failed to apply conditions to WiFi: %v", err)
	} else if applyResp.Success {
		fmt.Println("✅ Conditions applied successfully to WiFi interfaces!")
		fmt.Printf("📊 Applied: 2%% packet loss to WiFi\n")

		// Wait to simulate active conditions
		fmt.Printf("   🔄 Simulating network conditions for 3 seconds...\n")
		time.Sleep(3 * time.Second)

		// Reset using "wifi" keyword
		fmt.Println("\n🧹 Reset conditions using 'wifi' interface auto-discovery")

		resetReq := &pb.ResetNetworkConditionsRequest{
			Interface: "wifi", // Should reset all WiFi interfaces
		}

		resetResp, err := networkClient.ResetNetworkConditions(ctx, resetReq)
		if err != nil {
			log.Printf("❌ Failed to reset WiFi conditions: %v", err)
		} else if resetResp.Success {
			fmt.Println("✅ WiFi conditions reset successfully!")
		} else {
			fmt.Printf("❌ Reset failed: %s\n", resetResp.ErrorMessage)
		}
	} else {
		fmt.Printf("❌ Failed to apply conditions: %s\n", applyResp.ErrorMessage)
	}

	// Test 2: Apply mobile network scenarios using "wifi"
	fmt.Println("\n📱 Test 2: Apply various mobile network scenarios to WiFi")

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

	for i, scenario := range scenarios {
		fmt.Printf("\n   📶 Scenario %d: %s\n", i+1, scenario.name)

		// Apply scenario using "wifi" keyword
		scenarioReq := &pb.ApplyNetworkConditionsRequest{
			Interface:  "wifi", // Auto-discover WiFi interfaces
			Conditions: scenario.conditions,
		}

		scenarioResp, err := networkClient.ApplyNetworkConditions(ctx, scenarioReq)
		if err != nil {
			log.Printf("   ❌ Failed to apply %s: %v", scenario.name, err)
		} else if scenarioResp.Success {
			fmt.Printf("   ✅ %s applied to WiFi interfaces\n", scenario.name)

			// Show applied conditions
			if scenarioResp.AppliedConditions != nil {
				cond := scenarioResp.AppliedConditions
				if cond.Latency != nil && cond.Latency.Enabled {
					fmt.Printf("      ⏱️  Latency: %dms\n", cond.Latency.DelayMs)
				}
				if cond.PacketLoss != nil && cond.PacketLoss.Enabled {
					fmt.Printf("      📉 Packet Loss: %.1f%%\n", cond.PacketLoss.Percentage)
				}
			}

			time.Sleep(2 * time.Second)
		}
	}

	// Final cleanup using "wifi" keyword
	fmt.Println("\n🧹 Final cleanup: Reset all WiFi conditions")
	finalResetReq := &pb.ResetNetworkConditionsRequest{
		Interface: "wifi", // Auto-discover and reset all WiFi interfaces
	}

	finalResetResp, err := networkClient.ResetNetworkConditions(ctx, finalResetReq)
	if err != nil {
		log.Printf("❌ Final cleanup failed: %v", err)
	} else if finalResetResp.Success {
		fmt.Println("✅ All WiFi conditions cleared successfully!")
	}

	fmt.Println("\n🎉 Auto-WiFi discovery test completed!")
	fmt.Println("💡 Users can now use 'wifi' as interface name for automatic WiFi detection!")
	fmt.Println("🌐 This makes the API much more user-friendly for mobile device testing!")
}
