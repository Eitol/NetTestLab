package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	nettestlabv1 "github.com/Eitol/NetTestLab/api/nettestlab/v1"
	"github.com/Eitol/NetTestLab/internal/network"
	"github.com/Eitol/NetTestLab/internal/profiles"
	"github.com/Eitol/NetTestLab/internal/server"
)

var (
	port   = flag.Int("port", 8080, "Server port (gRPC and HTTP)")
	host   = flag.String("host", "0.0.0.0", "Server host")
	webDir = flag.String("web-dir", "web", "Web interface directory")
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

	// Setup combined server
	addr := fmt.Sprintf("%s:%d", *host, *port)
	combinedServer := setupCombinedServer(grpcServer)

	log.Printf("NetTestLab server starting on %s", addr)
	log.Printf("gRPC services available at: %s", addr)
	log.Printf("Web interface available at: http://%s/web", addr)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start combined server
	go func() {
		if err := combinedServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve combined server: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down server...")

	// Shutdown server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := combinedServer.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

func setupWebServer(addr string) *http.Server {
	mux := http.NewServeMux()

	// Serve static files
	staticDir := filepath.Join(*webDir, "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Serve index.html for all other routes (SPA)
	indexPath := filepath.Join(*webDir, "index.html")
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Set security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		
		// Serve index.html
		http.ServeFile(w, r, indexPath)
	})

	// API endpoints for web interface
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/version", handleVersion)

	return &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "nettestlab"}`))
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"version": "1.0.0", "service": "nettestlab"}`))
}
