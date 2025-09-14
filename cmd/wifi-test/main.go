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

	fmt.Printf("� NetTestLab WiFi Mobile Network Simulator\n")
	fmt.Printf("�🔗 Connecting to server at %s\n", *serverAddr)

	// Setup connection
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Connect to server
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create client
	networkClient := pb.NewNetworkControlServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("✅ Connected successfully!")

	// Test different mobile network scenarios on WiFi interface
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
				Bandwidth: &pb.BandwidthConfig{
					DownloadBps: 56000, // 56 kbps
					UploadBps:   28000, // 28 kbps
					Enabled:     true,
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
				Bandwidth: &pb.BandwidthConfig{
					DownloadBps: 384000, // 384 kbps
					UploadBps:   128000, // 128 kbps
					Enabled:     true,
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
				Bandwidth: &pb.BandwidthConfig{
					DownloadBps: 10000000, // 10 Mbps
					UploadBps:   5000000,  // 5 Mbps
					Enabled:     true,
				},
			},
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n📱 Test %d: Simulating %s on WiFi interface\n", i+1, scenario.name)

		applyReq := &pb.ApplyNetworkConditionsRequest{
			Interface:  "wl1-ap0",
			Conditions: scenario.conditions,
		}

		applyResp, err := networkClient.ApplyNetworkConditions(ctx, applyReq)
		if err != nil {
			log.Printf("❌ Failed to apply %s conditions: %v", scenario.name, err)
			continue
		}

		if applyResp.Success {
			fmt.Printf("✅ %s conditions applied successfully!\n", scenario.name)

			// Show applied conditions
			if applyResp.AppliedConditions != nil {
				cond := applyResp.AppliedConditions
				if cond.Latency != nil && cond.Latency.Enabled {
					fmt.Printf("   📶 Latency: %dms\n", cond.Latency.DelayMs)
				}
				if cond.PacketLoss != nil && cond.PacketLoss.Enabled {
					fmt.Printf("   � Packet Loss: %.1f%%\n", cond.PacketLoss.Percentage)
				}
				if cond.Bandwidth != nil && cond.Bandwidth.Enabled {
					fmt.Printf("   🌐 Bandwidth: %d/%d bps\n",
						cond.Bandwidth.DownloadBps, cond.Bandwidth.UploadBps)
				}
			}

			// Wait to simulate network conditions being active
			fmt.Printf("   ⏱️  Simulating network for 3 seconds...\n")
			time.Sleep(3 * time.Second)

			// Reset conditions
			resetReq := &pb.ResetNetworkConditionsRequest{
				Interface: "wl1-ap0",
			}

			resetResp, err := networkClient.ResetNetworkConditions(ctx, resetReq)
			if err != nil {
				log.Printf("❌ Failed to reset conditions: %v", err)
			} else if resetResp.Success {
				fmt.Printf("   🧹 %s conditions cleared\n", scenario.name)
			}
		} else {
			fmt.Printf("❌ Failed to apply %s conditions: %s\n", scenario.name, applyResp.ErrorMessage)
		}
	}

	fmt.Println("\n🎉 All WiFi mobile network simulations completed!")
	fmt.Println("📱 Your WiFi interface is ready for mobile device testing!")
}
