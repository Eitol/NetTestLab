package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Eitol/NetTestLab/api/nettestlab/v1"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	RouterIP       string
	RouterUser     string
	RouterPassword string
	OpenWrtSDK     string
	PackagePath    string
	SSHKeyPath     string
}

// Default test configuration - override with environment variables
var defaultConfig = TestConfig{
	RouterIP:       getEnvOrDefault("NETTESTLAB_ROUTER_IP", "192.168.1.1"),
	RouterUser:     getEnvOrDefault("NETTESTLAB_ROUTER_USER", "root"),
	RouterPassword: getEnvOrDefault("NETTESTLAB_ROUTER_PASSWORD", ""),
	OpenWrtSDK:     getEnvOrDefault("OPENWRT_SDK_PATH", ""),
	PackagePath:    getEnvOrDefault("NETTESTLAB_PACKAGE_PATH", ""),
	SSHKeyPath:     getEnvOrDefault("SSH_KEY_PATH", ""),
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestOpenWrtIntegration is the main integration test
func TestOpenWrtIntegration(t *testing.T) {
	config := defaultConfig

	// Skip if router IP is not set to a real value
	if config.RouterIP == "192.168.1.1" {
		t.Skip("NETTESTLAB_ROUTER_IP not set, skipping router integration test")
	}

	t.Run("GRPCConnectivity", func(t *testing.T) {
		testGRPCConnectivity(t, config)
	})

	t.Run("ProfileManagement", func(t *testing.T) {
		testProfileManagement(t, config)
	})

	t.Run("NetworkControl", func(t *testing.T) {
		testNetworkControl(t, config)
	})

	t.Run("SystemMonitoring", func(t *testing.T) {
		testSystemMonitoring(t, config)
	})
}

// testGRPCConnectivity tests basic gRPC connectivity
func testGRPCConnectivity(t *testing.T, config TestConfig) {
	t.Log("Testing gRPC connectivity...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to gRPC server
	conn, err := grpc.DialContext(ctx, config.RouterIP+":8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	// Test NetworkControlService
	networkClient := pb.NewNetworkControlServiceClient(conn)
	status, err := networkClient.GetSystemStatus(ctx, &pb.GetSystemStatusRequest{})
	if err != nil {
		t.Fatalf("Failed to get system status: %v", err)
	}

	if len(status.Interfaces) == 0 {
		t.Fatal("No network interfaces found")
	}

	t.Logf("System status: %d interfaces, version %s", len(status.Interfaces), status.Version)
	t.Log("gRPC connectivity test passed")
}

// testProfileManagement tests profile operations
func testProfileManagement(t *testing.T, config TestConfig) {
	t.Log("Testing profile management...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, config.RouterIP+":8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	profileClient := pb.NewProfileServiceClient(conn)

	// List profiles
	t.Log("Listing profiles...")
	listResp, err := profileClient.ListProfiles(ctx, &pb.ListProfilesRequest{})
	if err != nil {
		t.Fatalf("Failed to list profiles: %v", err)
	}

	if len(listResp.Profiles) == 0 {
		t.Fatal("No profiles found - built-in profiles should be available")
	}

	t.Logf("Found %d profiles", len(listResp.Profiles))

	// Verify built-in profiles exist
	expectedProfiles := []string{"2g", "3g", "4g", "wifi"}
	for _, expected := range expectedProfiles {
		found := false
		for _, profile := range listResp.Profiles {
			if profile.Name == expected {
				found = true
				if !profile.BuiltIn {
					t.Errorf("Profile %s should be marked as built-in", expected)
				}
				break
			}
		}
		if !found {
			t.Errorf("Built-in profile %s not found", expected)
		}
	}

	// Test creating a custom profile
	t.Log("Creating custom profile...")
	customProfile := &pb.NetworkProfile{
		Name:        "test_profile",
		DisplayName: "Test Profile",
		Description: "Test profile for integration testing",
		Conditions: &pb.NetworkConditions{
			Latency: &pb.LatencyConfig{
				DelayMs: 100,
				Enabled: true,
			},
		},
		Type: pb.ProfileType_PROFILE_TYPE_CUSTOM,
		Tags: []string{"test", "integration"},
	}

	createResp, err := profileClient.CreateProfile(ctx, &pb.CreateProfileRequest{
		Profile: customProfile,
	})
	if err != nil {
		t.Fatalf("Failed to create custom profile: %v", err)
	}

	if !createResp.Success {
		t.Fatalf("Profile creation failed: %s", createResp.ErrorMessage)
	}

	// Verify profile was created
	getResp, err := profileClient.GetProfile(ctx, &pb.GetProfileRequest{
		Name: "test_profile",
	})
	if err != nil {
		t.Fatalf("Failed to get created profile: %v", err)
	}

	if getResp.Profile.Name != "test_profile" {
		t.Errorf("Retrieved profile name mismatch: got %s, want test_profile", getResp.Profile.Name)
	}

	// Clean up - delete test profile
	_, err = profileClient.DeleteProfile(ctx, &pb.DeleteProfileRequest{
		Name: "test_profile",
	})
	if err != nil {
		t.Logf("Warning: failed to clean up test profile: %v", err)
	}

	t.Log("Profile management test passed")
}

// testNetworkControl tests network condition control
func testNetworkControl(t *testing.T, config TestConfig) {
	t.Log("Testing network control...")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, config.RouterIP+":8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	networkClient := pb.NewNetworkControlServiceClient(conn)
	profileClient := pb.NewProfileServiceClient(conn)

	// Get available interfaces
	t.Log("Getting system status...")
	statusResp, err := networkClient.GetSystemStatus(ctx, &pb.GetSystemStatusRequest{})
	if err != nil {
		t.Fatalf("Failed to get system status: %v", err)
	}

	if len(statusResp.Interfaces) == 0 {
		t.Fatal("No available interfaces found")
	}

	// Use first available interface for testing
	testInterface := statusResp.Interfaces[0].Name
	t.Logf("Using interface %s for testing", testInterface)

	// Test applying custom conditions
	t.Log("Applying custom network conditions...")
	applyResp, err := networkClient.ApplyNetworkConditions(ctx, &pb.ApplyNetworkConditionsRequest{
		Interface: testInterface,
		Conditions: &pb.NetworkConditions{
			Latency: &pb.LatencyConfig{
				DelayMs: 50,
				Enabled: true,
			},
			PacketLoss: &pb.PacketLossConfig{
				Percentage: 1.0,
				Enabled:    true,
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to apply network conditions: %v", err)
	}

	if !applyResp.Success {
		t.Fatalf("Failed to apply conditions: %s", applyResp.ErrorMessage)
	}

	t.Log("Network conditions applied successfully")

	// Verify conditions are applied
	time.Sleep(2 * time.Second)
	getResp, err := networkClient.GetNetworkConditions(ctx, &pb.GetNetworkConditionsRequest{
		Interface: testInterface,
	})
	if err != nil {
		t.Fatalf("Failed to get network conditions: %v", err)
	}

	if getResp.Conditions.Latency == nil || getResp.Conditions.Latency.DelayMs != 50 {
		t.Error("Latency condition not applied correctly")
	}

	// Test applying a profile
	t.Log("Applying 3G profile...")
	profileResp, err := profileClient.ApplyProfile(ctx, &pb.ApplyProfileRequest{
		Interface:   testInterface,
		ProfileName: "3g",
	})
	if err != nil {
		t.Fatalf("Failed to apply 3G profile: %v", err)
	}

	if !profileResp.Success {
		t.Fatalf("Failed to apply 3G profile: %s", profileResp.ErrorMessage)
	}

	t.Log("3G profile applied successfully")

	// Reset conditions
	t.Log("Resetting network conditions...")
	resetResp, err := networkClient.ResetNetworkConditions(ctx, &pb.ResetNetworkConditionsRequest{
		Interface: testInterface,
	})
	if err != nil {
		t.Fatalf("Failed to reset network conditions: %v", err)
	}

	if !resetResp.Success {
		t.Fatalf("Failed to reset conditions: %s", resetResp.ErrorMessage)
	}

	t.Log("Network control test passed")
}

// testSystemMonitoring tests monitoring capabilities
func testSystemMonitoring(t *testing.T, config TestConfig) {
	t.Log("Testing system monitoring...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, config.RouterIP+":8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	monitoringClient := pb.NewMonitoringServiceClient(conn)
	networkClient := pb.NewNetworkControlServiceClient(conn)

	// Get system health
	t.Log("Getting system health...")
	healthResp, err := monitoringClient.GetHealth(ctx, &pb.GetHealthRequest{})
	if err != nil {
		t.Fatalf("Failed to get system health: %v", err)
	}

	if healthResp.Status != pb.HealthStatus_HEALTH_STATUS_HEALTHY {
		t.Logf("Warning: System health status: %v", healthResp.Status)
	}

	t.Logf("System health: %v, version %s", healthResp.Status, healthResp.Version)

	// Get system metrics
	t.Log("Getting system metrics...")
	metricsResp, err := monitoringClient.GetMetrics(ctx, &pb.GetMetricsRequest{})
	if err != nil {
		t.Fatalf("Failed to get system metrics: %v", err)
	}

	if metricsResp.System != nil {
		cpu := metricsResp.System.CpuUsage
		mem := metricsResp.System.MemoryUsage
		
		if cpu < 0 || cpu > 100 {
			t.Errorf("Invalid CPU usage: %f", cpu)
		}

		if mem < 0 || mem > 100 {
			t.Errorf("Invalid memory usage: %f", mem)
		}

		t.Logf("System metrics: CPU %.1f%%, Memory %.1f%%", cpu, mem)
	}

	// Get system status to see interfaces
	statusResp, err := networkClient.GetSystemStatus(ctx, &pb.GetSystemStatusRequest{})
	if err != nil {
		t.Fatalf("Failed to get system status: %v", err)
	}

	if len(statusResp.Interfaces) == 0 {
		t.Error("No interfaces found")
	}

	for _, iface := range statusResp.Interfaces {
		t.Logf("Interface %s: type %v, up %v", iface.Name, iface.Type, iface.IsUp)
	}

	t.Log("System monitoring test passed")
}

// Helper functions

func getProjectRoot() (string, error) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Look for go.mod file to find project root
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			return "", fmt.Errorf("project root not found")
		}
		wd = parent
	}
}

func findPackageFile(packagesDir, packageName string) (string, error) {
	var packageFile string

	err := filepath.Walk(packagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.Contains(info.Name(), packageName) && strings.HasSuffix(info.Name(), ".ipk") {
			packageFile = path
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if packageFile == "" {
		return "", fmt.Errorf("package file not found")
	}

	return packageFile, nil
}

func scpFile(localPath, remoteHost, remotePath string, config TestConfig) error {
	var cmd *exec.Cmd

	if config.SSHKeyPath != "" {
		cmd = exec.Command("scp", "-i", config.SSHKeyPath, localPath, 
			fmt.Sprintf("%s@%s:%s", config.RouterUser, remoteHost, remotePath))
	} else {
		cmd = exec.Command("scp", localPath, 
			fmt.Sprintf("%s@%s:%s", config.RouterUser, remoteHost, remotePath))
	}

	if config.RouterPassword != "" {
		// Use sshpass if password is provided
		cmd = exec.Command("sshpass", "-p", config.RouterPassword, "scp", localPath,
			fmt.Sprintf("%s@%s:%s", config.RouterUser, remoteHost, remotePath))
	}

	return cmd.Run()
}

func runSSHCommand(config TestConfig, command string) (string, error) {
	var cmd *exec.Cmd

	if config.SSHKeyPath != "" {
		cmd = exec.Command("ssh", "-i", config.SSHKeyPath,
			fmt.Sprintf("%s@%s", config.RouterUser, config.RouterIP), command)
	} else {
		cmd = exec.Command("ssh", 
			fmt.Sprintf("%s@%s", config.RouterUser, config.RouterIP), command)
	}

	if config.RouterPassword != "" {
		// Use sshpass if password is provided
		cmd = exec.Command("sshpass", "-p", config.RouterPassword, "ssh",
			fmt.Sprintf("%s@%s", config.RouterUser, config.RouterIP), command)
	}

	output, err := cmd.CombinedOutput()
	return string(output), err
}