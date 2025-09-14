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

	fmt.Printf("🔬 Simple WiFi Test - Single Condition\n")
	fmt.Printf("🔗 Connecting to server at %s\n", *serverAddr)

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

	// Apply ONLY packet loss (simplest condition)
	fmt.Println("\n📱 Step 1: Apply simple packet loss to WiFi")

	applyReq := &pb.ApplyNetworkConditionsRequest{
		Interface: "wl1-ap0",
		Conditions: &pb.NetworkConditions{
			PacketLoss: &pb.PacketLossConfig{
				Percentage: 2.0,
				Enabled:    true,
			},
		},
	}

	applyResp, err := networkClient.ApplyNetworkConditions(ctx, applyReq)
	if err != nil {
		log.Printf("❌ Failed to apply conditions: %v", err)
		return
	}

	if applyResp.Success {
		fmt.Println("✅ Packet loss applied successfully!")

		// Check what was actually applied
		fmt.Println("\n🔍 Step 2: Manual verification")
		time.Sleep(1 * time.Second)

		// Now reset
		fmt.Println("\n🧹 Step 3: Reset conditions")

		resetReq := &pb.ResetNetworkConditionsRequest{
			Interface: "wl1-ap0",
		}

		resetResp, err := networkClient.ResetNetworkConditions(ctx, resetReq)
		if err != nil {
			log.Printf("❌ Failed to reset conditions: %v", err)
		} else if resetResp.Success {
			fmt.Println("✅ Conditions reset successfully!")
		} else {
			fmt.Printf("❌ Reset failed: %s\n", resetResp.ErrorMessage)
		}
	} else {
		fmt.Printf("❌ Failed to apply conditions: %s\n", applyResp.ErrorMessage)
	}

	fmt.Println("\n🎉 Simple test completed!")
}
