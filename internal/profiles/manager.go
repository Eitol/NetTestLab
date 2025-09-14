package profiles

import (
	"fmt"
	"sync"

	nettestlabv1 "github.com/nettestlab/nettestlab/api/nettestlab/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Manager handles network profiles
type Manager struct {
	mu       sync.RWMutex
	profiles map[string]*nettestlabv1.NetworkProfile
}

// NewManager creates a new profile manager
func NewManager() *Manager {
	return &Manager{
		profiles: make(map[string]*nettestlabv1.NetworkProfile),
	}
}

// LoadBuiltInProfiles loads predefined network profiles
func (m *Manager) LoadBuiltInProfiles() error {
	m.mu.Lock()
	defer m.mu.Unlock()

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

	// Add all built-in profiles
	for _, profile := range builtInProfiles {
		m.profiles[profile.Name] = profile
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

	// Store profile
	m.profiles[profile.Name] = profile

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

	// Cannot update built-in profiles
	if existing.BuiltIn {
		return fmt.Errorf("cannot update built-in profile %s", name)
	}

	// Update metadata
	profile.CreatedAt = existing.CreatedAt
	profile.UpdatedAt = timestamppb.Now()
	profile.Version = existing.Version + 1
	profile.BuiltIn = false
	profile.Name = name // Ensure name consistency

	// Store updated profile
	m.profiles[name] = profile

	return nil
}

// DeleteProfile deletes a profile
func (m *Manager) DeleteProfile(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	profile, exists := m.profiles[name]
	if !exists {
		return fmt.Errorf("profile %s not found", name)
	}

	// Cannot delete built-in profiles
	if profile.BuiltIn {
		return fmt.Errorf("cannot delete built-in profile %s", name)
	}

	delete(m.profiles, name)
	return nil
}
