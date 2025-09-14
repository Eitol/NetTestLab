package profiles

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	nettestlabv1 "github.com/Eitol/NetTestLab/api/nettestlab/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Manager handles network profiles
type Manager struct {
	mu          sync.RWMutex
	profiles    map[string]*nettestlabv1.NetworkProfile
	profilesDir string
}

// NewManager creates a new profile manager
func NewManager() *Manager {
	return NewManagerWithProfilesDir("./data/profiles")
}

// NewManagerWithProfilesDir creates a new profile manager with custom profiles directory
func NewManagerWithProfilesDir(profilesDir string) *Manager {
	return &Manager{
		profiles:    make(map[string]*nettestlabv1.NetworkProfile),
		profilesDir: profilesDir,
	}
}

// ensureProfilesDir creates the profiles directory if it doesn't exist
func (m *Manager) ensureProfilesDir() error {
	return os.MkdirAll(m.profilesDir, 0755)
}

// getProfileFilePath returns the file path for a given profile name
func (m *Manager) getProfileFilePath(profileName string) string {
	return filepath.Join(m.profilesDir, profileName+".json")
}

// saveProfile saves a single profile to its JSON file
func (m *Manager) saveProfile(profile *nettestlabv1.NetworkProfile) error {
	if err := m.ensureProfilesDir(); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}

	profileJSON, err := protojson.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile %s: %w", profile.Name, err)
	}

	filePath := m.getProfileFilePath(profile.Name)
	if err := os.WriteFile(filePath, profileJSON, 0644); err != nil {
		return fmt.Errorf("failed to write profile file %s: %w", filePath, err)
	}

	return nil
}

// loadProfile loads a single profile from its JSON file
func (m *Manager) loadProfile(profileName string) (*nettestlabv1.NetworkProfile, error) {
	filePath := m.getProfileFilePath(profileName)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("profile file %s not found", filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile file %s: %w", filePath, err)
	}

	profile := &nettestlabv1.NetworkProfile{}
	if err := protojson.Unmarshal(data, profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile %s: %w", profileName, err)
	}

	return profile, nil
}

// loadAllProfiles loads all profiles from the profiles directory
func (m *Manager) loadAllProfiles() error {
	// Check if directory exists
	if _, err := os.Stat(m.profilesDir); os.IsNotExist(err) {
		// Directory doesn't exist, which is fine for first run
		return nil
	}

	entries, err := os.ReadDir(m.profilesDir)
	if err != nil {
		return fmt.Errorf("failed to read profiles directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// Extract profile name from filename
		profileName := strings.TrimSuffix(entry.Name(), ".json")

		profile, err := m.loadProfile(profileName)
		if err != nil {
			// Log error but continue with other profiles
			fmt.Printf("Warning: failed to load profile %s: %v\n", profileName, err)
			continue
		}

		m.profiles[profileName] = profile
	}

	return nil
}

// deleteProfile removes a profile file from disk
func (m *Manager) deleteProfileFile(profileName string) error {
	filePath := m.getProfileFilePath(profileName)

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete profile file %s: %w", filePath, err)
	}

	return nil
}

// createDefaultProfiles creates the default built-in profiles if they don't exist
func (m *Manager) createDefaultProfiles() error {
	builtInProfiles := []*nettestlabv1.NetworkProfile{
		// 2G Network Profile
		{
			Name:        "2g",
			DisplayName: "2G Mobile Network",
			Description: "Simulates 2G mobile network conditions with high latency and low bandwidth",
			Type:        nettestlabv1.ProfileType_PROFILE_TYPE_MOBILE,
			BuiltIn:     true,
			Tags:        []string{"mobile", "2g", "slow"},
			CreatedAt:   timestamppb.Now(),
			UpdatedAt:   timestamppb.Now(),
			Version:     1,
			Conditions: &nettestlabv1.NetworkConditions{
				Latency: &nettestlabv1.LatencyConfig{
					DelayMs: 500,
					Enabled: true,
				},
				PacketLoss: &nettestlabv1.PacketLossConfig{
					Percentage: 2.0,
					Enabled:    true,
					Pattern:    nettestlabv1.LossPattern_LOSS_PATTERN_RANDOM,
				},
				Bandwidth: &nettestlabv1.BandwidthConfig{
					DownloadBps: 56000, // 56 Kbps
					UploadBps:   28000, // 28 Kbps
					Enabled:     true,
				},
			},
		},
		// 3G Network Profile
		{
			Name:        "3g",
			DisplayName: "3G Mobile Network",
			Description: "Simulates 3G mobile network conditions with moderate latency and bandwidth",
			Type:        nettestlabv1.ProfileType_PROFILE_TYPE_MOBILE,
			BuiltIn:     true,
			Tags:        []string{"mobile", "3g", "moderate"},
			CreatedAt:   timestamppb.Now(),
			UpdatedAt:   timestamppb.Now(),
			Version:     1,
			Conditions: &nettestlabv1.NetworkConditions{
				Latency: &nettestlabv1.LatencyConfig{
					DelayMs: 150,
					Enabled: true,
				},
				PacketLoss: &nettestlabv1.PacketLossConfig{
					Percentage: 0.5,
					Enabled:    true,
					Pattern:    nettestlabv1.LossPattern_LOSS_PATTERN_RANDOM,
				},
				Bandwidth: &nettestlabv1.BandwidthConfig{
					DownloadBps: 1600000, // 1.6 Mbps
					UploadBps:   384000,  // 384 Kbps
					Enabled:     true,
				},
			},
		},
		// 4G Network Profile
		{
			Name:        "4g",
			DisplayName: "4G/LTE Mobile Network",
			Description: "Simulates 4G/LTE mobile network conditions with low latency and high bandwidth",
			Type:        nettestlabv1.ProfileType_PROFILE_TYPE_MOBILE,
			BuiltIn:     true,
			Tags:        []string{"mobile", "4g", "lte", "fast"},
			CreatedAt:   timestamppb.Now(),
			UpdatedAt:   timestamppb.Now(),
			Version:     1,
			Conditions: &nettestlabv1.NetworkConditions{
				Latency: &nettestlabv1.LatencyConfig{
					DelayMs: 50,
					Enabled: true,
				},
				PacketLoss: &nettestlabv1.PacketLossConfig{
					Percentage: 0.1,
					Enabled:    true,
					Pattern:    nettestlabv1.LossPattern_LOSS_PATTERN_RANDOM,
				},
				Bandwidth: &nettestlabv1.BandwidthConfig{
					DownloadBps: 50000000, // 50 Mbps
					UploadBps:   10000000, // 10 Mbps
					Enabled:     true,
				},
			},
		},
		// 5G Network Profile
		{
			Name:        "5g",
			DisplayName: "5G Mobile Network",
			Description: "Simulates 5G mobile network conditions with very low latency and very high bandwidth",
			Type:        nettestlabv1.ProfileType_PROFILE_TYPE_MOBILE,
			BuiltIn:     true,
			Tags:        []string{"mobile", "5g", "ultra-fast"},
			CreatedAt:   timestamppb.Now(),
			UpdatedAt:   timestamppb.Now(),
			Version:     1,
			Conditions: &nettestlabv1.NetworkConditions{
				Latency: &nettestlabv1.LatencyConfig{
					DelayMs: 10,
					Enabled: true,
				},
				PacketLoss: &nettestlabv1.PacketLossConfig{
					Percentage: 0.01,
					Enabled:    true,
					Pattern:    nettestlabv1.LossPattern_LOSS_PATTERN_RANDOM,
				},
				Bandwidth: &nettestlabv1.BandwidthConfig{
					DownloadBps: 1000000000, // 1 Gbps
					UploadBps:   100000000,  // 100 Mbps
					Enabled:     true,
				},
			},
		},
		// WiFi Network Profile
		{
			Name:        "wifi",
			DisplayName: "WiFi Network",
			Description: "Simulates typical WiFi network conditions with low latency and high bandwidth",
			Type:        nettestlabv1.ProfileType_PROFILE_TYPE_WIFI,
			BuiltIn:     true,
			Tags:        []string{"wifi", "broadband", "fast"},
			CreatedAt:   timestamppb.Now(),
			UpdatedAt:   timestamppb.Now(),
			Version:     1,
			Conditions: &nettestlabv1.NetworkConditions{
				Latency: &nettestlabv1.LatencyConfig{
					DelayMs: 5,
					Enabled: true,
				},
				PacketLoss: &nettestlabv1.PacketLossConfig{
					Percentage: 0.0,
					Enabled:    false,
				},
				Bandwidth: &nettestlabv1.BandwidthConfig{
					DownloadBps: 100000000, // 100 Mbps
					UploadBps:   100000000, // 100 Mbps
					Enabled:     true,
				},
			},
		},
		// Satellite Network Profile
		{
			Name:        "satellite",
			DisplayName: "Satellite Network",
			Description: "Simulates satellite network conditions with very high latency",
			Type:        nettestlabv1.ProfileType_PROFILE_TYPE_SATELLITE,
			BuiltIn:     true,
			Tags:        []string{"satellite", "high-latency"},
			CreatedAt:   timestamppb.Now(),
			UpdatedAt:   timestamppb.Now(),
			Version:     1,
			Conditions: &nettestlabv1.NetworkConditions{
				Latency: &nettestlabv1.LatencyConfig{
					DelayMs: 600,
					Enabled: true,
				},
				PacketLoss: &nettestlabv1.PacketLossConfig{
					Percentage: 1.0,
					Enabled:    true,
					Pattern:    nettestlabv1.LossPattern_LOSS_PATTERN_BURST,
				},
				Bandwidth: &nettestlabv1.BandwidthConfig{
					DownloadBps: 25000000, // 25 Mbps
					UploadBps:   3000000,  // 3 Mbps
					Enabled:     true,
				},
				Jitter: &nettestlabv1.JitterConfig{
					VariationMs:  50,
					Enabled:      true,
					Distribution: nettestlabv1.JitterDistribution_JITTER_DISTRIBUTION_UNIFORM,
				},
			},
		},
	}

	// Create each profile file if it doesn't exist
	for _, profile := range builtInProfiles {
		filePath := m.getProfileFilePath(profile.Name)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			if err := m.saveProfile(profile); err != nil {
				return fmt.Errorf("failed to create default profile %s: %w", profile.Name, err)
			}
		}
	}

	return nil
}

// LoadBuiltInProfiles loads all profiles from disk and creates default ones if they don't exist
func (m *Manager) LoadBuiltInProfiles() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// First, create default profiles if they don't exist
	if err := m.createDefaultProfiles(); err != nil {
		return fmt.Errorf("failed to create default profiles: %w", err)
	}

	// Then, load all profiles from disk
	if err := m.loadAllProfiles(); err != nil {
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	return nil
}

// ListProfiles returns all profiles matching the filter
func (m *Manager) ListProfiles(profileType nettestlabv1.ProfileType, builtInOnly bool) []*nettestlabv1.NetworkProfile {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*nettestlabv1.NetworkProfile

	for _, profile := range m.profiles {
		// Filter by type if specified
		if profileType != nettestlabv1.ProfileType_PROFILE_TYPE_UNSPECIFIED && profile.Type != profileType {
			continue
		}

		// Filter by built-in if specified
		if builtInOnly && !profile.BuiltIn {
			continue
		}

		result = append(result, profile)
	}

	return result
}

// GetProfile returns a specific profile by name
func (m *Manager) GetProfile(name string) (*nettestlabv1.NetworkProfile, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	profile, exists := m.profiles[name]
	return profile, exists
}

// CreateProfile creates a new custom profile
func (m *Manager) CreateProfile(profile *nettestlabv1.NetworkProfile) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if profile already exists
	if _, exists := m.profiles[profile.Name]; exists {
		return fmt.Errorf("profile %s already exists", profile.Name)
	}

	// Set metadata
	now := timestamppb.Now()
	profile.CreatedAt = now
	profile.UpdatedAt = now
	profile.Version = 1
	profile.BuiltIn = false

	// Store profile in memory
	m.profiles[profile.Name] = profile

	// Save to disk
	if err := m.saveProfile(profile); err != nil {
		// Remove from memory if save failed
		delete(m.profiles, profile.Name)
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// UpdateProfile updates an existing profile
func (m *Manager) UpdateProfile(name string, profile *nettestlabv1.NetworkProfile) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, exists := m.profiles[name]
	if !exists {
		return fmt.Errorf("profile %s not found", name)
	}

	// Store original for rollback
	original := m.profiles[name]

	// Update metadata - preserve BuiltIn status and timestamps for existing profiles
	profile.CreatedAt = existing.CreatedAt
	profile.UpdatedAt = timestamppb.Now()
	profile.Version = existing.Version + 1
	profile.BuiltIn = existing.BuiltIn // Preserve built-in status
	profile.Name = name                // Ensure name consistency

	// Store updated profile in memory
	m.profiles[name] = profile

	// Save to disk
	if err := m.saveProfile(profile); err != nil {
		// Rollback on failure
		m.profiles[name] = original
		return fmt.Errorf("failed to save updated profile: %w", err)
	}

	return nil
}

// DeleteProfile deletes a profile
func (m *Manager) DeleteProfile(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.profiles[name]
	if !exists {
		return fmt.Errorf("profile %s not found", name)
	}

	// Store original for rollback
	original := m.profiles[name]

	// Delete from memory
	delete(m.profiles, name)

	// Delete from disk
	if err := m.deleteProfileFile(name); err != nil {
		// Rollback on failure
		m.profiles[name] = original
		return fmt.Errorf("failed to delete profile file: %w", err)
	}

	return nil
}
