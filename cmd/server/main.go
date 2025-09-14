package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	nettestlabv1 "github.com/nettestlab/nettestlab/api/nettestlab/v1"
	"github.com/nettestlab/nettestlab/internal/network"
	"github.com/nettestlab/nettestlab/internal/profiles"
	"github.com/nettestlab/nettestlab/internal/server"
)

var (
	port = flag.Int("port", 8080, "gRPC server port")
	host = flag.String("host", "0.0.0.0", "gRPC server host")
)

func main() {
	flag.Parse()

	// Initialize network controller
	networkController, err := network.NewController()
	if err != nil {
		log.Fatalf("Failed to initialize network controller: %v", err)
	}

	// Initialize profile manager
	profileManager := profiles.NewManager()

	// Load built-in profiles
	if err := profileManager.LoadBuiltInProfiles(); err != nil {
		log.Printf("Warning: Failed to load built-in profiles: %v", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	networkService := server.NewNetworkControlService(networkController)
	profileService := server.NewProfileService(profileManager, networkController)
	monitoringService := server.NewMonitoringService(networkController)

	nettestlabv1.RegisterNetworkControlServiceServer(grpcServer, networkService)
	nettestlabv1.RegisterProfileServiceServer(grpcServer, profileService)
	nettestlabv1.RegisterMonitoringServiceServer(grpcServer, monitoringService)

	// Enable reflection for easier testing
	reflection.Register(grpcServer)

	// Setup listener
	addr := fmt.Sprintf("%s:%d", *host, *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	log.Printf("NetTestLab gRPC server starting on %s", addr)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop()
	}()

	// Start server
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}

	log.Println("Server stopped")
}
