package profiles

import (
	"os"
	"path/filepath"
	"testing"

	nettestlabv1 "github.com/Eitol/NetTestLab/api/nettestlab/v1"
)

func TestProfilePersistence(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "nettestlab_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	manager := NewManagerWithProfilesDir(tempDir)

	// Load built-in profiles
	if err := manager.LoadBuiltInProfiles(); err != nil {
		t.Fatalf("Failed to load built-in profiles: %v", err)
	}

	// Create a custom profile
	customProfile := &nettestlabv1.NetworkProfile{
		Name:        "test-profile",
		DisplayName: "Test Profile",
		Description: "A test profile for persistence testing",
		Type:        nettestlabv1.ProfileType_PROFILE_TYPE_CUSTOM,
		Tags:        []string{"test", "custom"},
		Conditions: &nettestlabv1.NetworkConditions{
			Latency: &nettestlabv1.LatencyConfig{
				DelayMs: 100,
				Enabled: true,
			},
		},
	}

	// Add the custom profile
	if err := manager.CreateProfile(customProfile); err != nil {
		t.Fatalf("Failed to create custom profile: %v", err)
	}

	// Verify the profile exists
	retrieved, exists := manager.GetProfile("test-profile")
	if !exists {
		t.Fatal("Custom profile not found after creation")
	}

	if retrieved.Name != "test-profile" {
		t.Errorf("Expected profile name 'test-profile', got '%s'", retrieved.Name)
	}

	// Verify persistence file was created
	profileFile := filepath.Join(tempDir, "test-profile.json")
	if _, err := os.Stat(profileFile); os.IsNotExist(err) {
		t.Fatal("Profile file was not created")
	}

	// Create a new manager with the same directory
	manager2 := NewManagerWithProfilesDir(tempDir)
	if err := manager2.LoadBuiltInProfiles(); err != nil {
		t.Fatalf("Failed to load profiles in second manager: %v", err)
	}

	// Verify the custom profile was loaded
	retrieved2, exists2 := manager2.GetProfile("test-profile")
	if !exists2 {
		t.Fatal("Custom profile not found after reload")
	}

	if retrieved2.Name != "test-profile" {
		t.Errorf("Expected profile name 'test-profile' after reload, got '%s'", retrieved2.Name)
	}

	if retrieved2.DisplayName != "Test Profile" {
		t.Errorf("Expected display name 'Test Profile' after reload, got '%s'", retrieved2.DisplayName)
	}

	// Verify built-in profiles are not persisted
	builtInProfile, exists := manager2.GetProfile("4g")
	if !exists {
		t.Fatal("Built-in profile '4g' not found")
	}
	if !builtInProfile.BuiltIn {
		t.Error("Built-in profile should have BuiltIn=true")
	}
}

func TestProfileUpdate(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "nettestlab_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager
	manager := NewManagerWithProfilesDir(tempDir)
	if err := manager.LoadBuiltInProfiles(); err != nil {
		t.Fatalf("Failed to load built-in profiles: %v", err)
	}

	// Create a custom profile
	originalProfile := &nettestlabv1.NetworkProfile{
		Name:        "update-test",
		DisplayName: "Original Profile",
		Description: "Original description",
		Type:        nettestlabv1.ProfileType_PROFILE_TYPE_CUSTOM,
		Conditions: &nettestlabv1.NetworkConditions{
			Latency: &nettestlabv1.LatencyConfig{
				DelayMs: 50,
				Enabled: true,
			},
		},
	}

	if err := manager.CreateProfile(originalProfile); err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// Update the profile
	updatedProfile := &nettestlabv1.NetworkProfile{
		Name:        "update-test",
		DisplayName: "Updated Profile",
		Description: "Updated description",
		Type:        nettestlabv1.ProfileType_PROFILE_TYPE_CUSTOM,
		Conditions: &nettestlabv1.NetworkConditions{
			Latency: &nettestlabv1.LatencyConfig{
				DelayMs: 100,
				Enabled: true,
			},
		},
	}

	if err := manager.UpdateProfile("update-test", updatedProfile); err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}

	// Verify the update
	retrieved, exists := manager.GetProfile("update-test")
	if !exists {
		t.Fatal("Profile not found after update")
	}

	if retrieved.DisplayName != "Updated Profile" {
		t.Errorf("Expected updated display name, got '%s'", retrieved.DisplayName)
	}

	if retrieved.Conditions.Latency.DelayMs != 100 {
		t.Errorf("Expected updated latency 100ms, got %d", retrieved.Conditions.Latency.DelayMs)
	}

	// Verify persistence
	manager2 := NewManagerWithProfilesDir(tempDir)
	if err := manager2.LoadBuiltInProfiles(); err != nil {
		t.Fatalf("Failed to load profiles in second manager: %v", err)
	}

	retrieved2, exists2 := manager2.GetProfile("update-test")
	if !exists2 {
		t.Fatal("Updated profile not found after reload")
	}

	if retrieved2.DisplayName != "Updated Profile" {
		t.Errorf("Expected updated display name after reload, got '%s'", retrieved2.DisplayName)
	}
}

func TestProfileDeletion(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "nettestlab_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager
	manager := NewManagerWithProfilesDir(tempDir)
	if err := manager.LoadBuiltInProfiles(); err != nil {
		t.Fatalf("Failed to load built-in profiles: %v", err)
	}

	// Create a custom profile
	customProfile := &nettestlabv1.NetworkProfile{
		Name:        "delete-test",
		DisplayName: "To Be Deleted",
		Description: "This profile will be deleted",
		Type:        nettestlabv1.ProfileType_PROFILE_TYPE_CUSTOM,
	}

	if err := manager.CreateProfile(customProfile); err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// Verify it exists
	_, exists := manager.GetProfile("delete-test")
	if !exists {
		t.Fatal("Profile not found after creation")
	}

	// Delete the profile
	if err := manager.DeleteProfile("delete-test"); err != nil {
		t.Fatalf("Failed to delete profile: %v", err)
	}

	// Verify it's gone
	_, exists = manager.GetProfile("delete-test")
	if exists {
		t.Fatal("Profile still exists after deletion")
	}

	// Verify persistence
	manager2 := NewManagerWithProfilesDir(tempDir)
	if err := manager2.LoadBuiltInProfiles(); err != nil {
		t.Fatalf("Failed to load profiles in second manager: %v", err)
	}

	_, exists2 := manager2.GetProfile("delete-test")
	if exists2 {
		t.Fatal("Deleted profile found after reload")
	}
}

func TestBuiltInProfileProtection(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "nettestlab_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager
	manager := NewManagerWithProfilesDir(tempDir)
	if err := manager.LoadBuiltInProfiles(); err != nil {
		t.Fatalf("Failed to load built-in profiles: %v", err)
	}

	// Built-in profiles can now be updated (since they're file-based)
	// Test that we can update a built-in profile
	updatedProfile := &nettestlabv1.NetworkProfile{
		Name:        "4g",
		DisplayName: "Modified 4G",
		Description: "This should work now",
		Type:        nettestlabv1.ProfileType_PROFILE_TYPE_MOBILE,
		Conditions: &nettestlabv1.NetworkConditions{
			Latency: &nettestlabv1.LatencyConfig{
				DelayMs: 25,
				Enabled: true,
			},
		},
	}

	err = manager.UpdateProfile("4g", updatedProfile)
	if err != nil {
		t.Fatalf("Failed to update built-in profile: %v", err)
	}

	// Verify the update worked
	retrieved, exists := manager.GetProfile("4g")
	if !exists {
		t.Fatal("Built-in profile not found after update")
	}

	if retrieved.DisplayName != "Modified 4G" {
		t.Errorf("Expected updated display name, got '%s'", retrieved.DisplayName)
	}

	// Built-in profiles can also be deleted now (they're just files)
	err = manager.DeleteProfile("4g")
	if err != nil {
		t.Fatalf("Failed to delete built-in profile: %v", err)
	}

	// Verify deletion
	_, exists = manager.GetProfile("4g")
	if exists {
		t.Fatal("Built-in profile still exists after deletion")
	}
}
