package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"connectrpc.com/grpcreflect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/Eitol/NetTestLab/api/nettestlab/v1/nettestlabv1connect"
	"github.com/Eitol/NetTestLab/internal/network"
	"github.com/Eitol/NetTestLab/internal/profiles"
	"github.com/Eitol/NetTestLab/internal/server"
)

var (
	port        = flag.Int("port", 8080, "Server port")
	host        = flag.String("host", "0.0.0.0", "Server host")
	profilesDir = flag.String("profiles-dir", "./data/profiles", "Directory to store profiles")
	dataDir     = flag.String("data-dir", "./data", "Base data directory")
	webDir      = flag.String("web-dir", "web", "Web interface directory")
)

func main() {
	flag.Parse()

	// Create data directories if they don't exist
	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}
	if err := os.MkdirAll(*profilesDir, 0755); err != nil {
		log.Fatalf("Failed to create profiles directory: %v", err)
	}

	// Initialize network controller
	networkController, err := network.NewController()
	if err != nil {
		log.Fatalf("Failed to initialize network controller: %v", err)
	}

	// Initialize profile manager with custom directory
	profileManager := profiles.NewManagerWithProfilesDir(*profilesDir)

	// Load built-in profiles
	if err := profileManager.LoadBuiltInProfiles(); err != nil {
		log.Printf("Warning: Failed to load built-in profiles: %v", err)
	}

	// Create Connect-compatible services (now implementing Connect interfaces directly)
	networkService := server.NewNetworkControlService(networkController)
	profileService := server.NewProfileService(profileManager, networkController)
	monitoringService := server.NewMonitoringService(networkController)

	// Create traffic capture service
	trafficCaptureService, err := server.NewTrafficCaptureService(*dataDir)
	if err != nil {
		log.Fatalf("Failed to initialize traffic capture service: %v", err)
	}
	defer trafficCaptureService.Close()

	// Create HTTP mux
	mux := http.NewServeMux()

	// Setup static files and web interface
	setupWebInterface(mux)

	// Create Connect handlers (services now implement Connect interfaces directly)
	networkPath, networkHandler := nettestlabv1connect.NewNetworkControlServiceHandler(networkService)
	profilePath, profileHandler := nettestlabv1connect.NewProfileServiceHandler(profileService)
	monitoringPath, monitoringHandler := nettestlabv1connect.NewMonitoringServiceHandler(monitoringService)
	trafficCapturePath, trafficCaptureHandler := nettestlabv1connect.NewTrafficCaptureServiceHandler(trafficCaptureService)

	// Mount Connect handlers
	mux.Handle(networkPath, networkHandler)
	mux.Handle(profilePath, profileHandler)
	mux.Handle(monitoringPath, monitoringHandler)
	mux.Handle(trafficCapturePath, trafficCaptureHandler)

	// Add reflection for easier debugging
	reflector := grpcreflect.NewStaticReflector(
		nettestlabv1connect.NetworkControlServiceName,
		nettestlabv1connect.ProfileServiceName,
		nettestlabv1connect.MonitoringServiceName,
		nettestlabv1connect.TrafficCaptureServiceName,
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// Create server with HTTP/2 support (h2c for unencrypted HTTP/2)
	addr := fmt.Sprintf("%s:%d", *host, *port)
	server := &http.Server{
		Addr:         addr,
		Handler:      h2c.NewHandler(mux, &http2.Server{}),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("NetTestLab Connect server starting on %s", addr)
	log.Printf("Connect services available at: %s", addr)
	log.Printf("Web interface available at: http://%s/", addr)
	log.Printf("Health check: curl -X POST %s/nettestlab.v1.MonitoringService/GetHealth", addr)

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down server...")

	// Shutdown server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

func setupWebInterface(mux *http.ServeMux) {
	// Serve static files with no-cache headers
	staticDir := filepath.Join(*webDir, "static")
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir)))
	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		// Add aggressive no-cache headers for all static files, especially JS
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
		w.Header().Set("Last-Modified", "Thu, 01 Jan 1970 00:00:00 GMT")
		w.Header().Set("ETag", "")

		// Specific headers for JS files
		if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		}

		staticHandler.ServeHTTP(w, r)
	})

	// Serve index.html for root and fallback
	indexPath := filepath.Join(*webDir, "index.html")
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Only serve index.html for root path, let Connect handle API routes
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		http.ServeFile(w, r, indexPath)
	})
}
