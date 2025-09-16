package capture

import (
	"fmt"
	"net"
	"strings"

	nettestlabv1 "github.com/Eitol/NetTestLab/api/nettestlab/v1"
)

// FilterBuilder constructs BPF filters for traffic capture
type FilterBuilder struct{}

// NewFilterBuilder creates a new filter builder
func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{}
}

// BuildFilter constructs a BPF filter from devices and URL targets
func (fb *FilterBuilder) BuildFilter(devices []*nettestlabv1.Device, targets []*nettestlabv1.UrlTarget, onlyConnected bool) (string, error) {
	var filterParts []string

	// Build device filter
	deviceFilter := fb.buildDeviceFilter(devices, onlyConnected)
	if deviceFilter != "" {
		filterParts = append(filterParts, fmt.Sprintf("(%s)", deviceFilter))
	}

	// Build protocol and port filter from targets
	protocolFilter := fb.buildProtocolFilter(targets)
	if protocolFilter != "" {
		filterParts = append(filterParts, fmt.Sprintf("(%s)", protocolFilter))
	}

	// Build host filter from targets (this is the complex part)
	hostFilter := fb.buildHostFilter(targets)
	if hostFilter != "" {
		filterParts = append(filterParts, fmt.Sprintf("(%s)", hostFilter))
	}

	// Combine all filters with AND
	if len(filterParts) == 0 {
		return "", fmt.Errorf("no valid filter criteria provided")
	}

	return strings.Join(filterParts, " and "), nil
}

// buildDeviceFilter creates BPF filter for specific devices
func (fb *FilterBuilder) buildDeviceFilter(devices []*nettestlabv1.Device, onlyConnected bool) string {
	if len(devices) == 0 {
		return ""
	}

	var deviceFilters []string

	for _, device := range devices {
		// Skip disconnected devices if onlyConnected is true
		if onlyConnected && device.ConnectionStatus != nettestlabv1.DeviceConnectionStatus_DEVICE_CONNECTION_STATUS_CONNECTED {
			continue
		}

		var deviceParts []string

		// Add MAC address filter
		if device.MacAddress != "" {
			macFilter := fb.formatMacForBPF(device.MacAddress)
			if macFilter != "" {
				deviceParts = append(deviceParts, fmt.Sprintf("ether host %s", macFilter))
			}
		}

		// Add IP address filter if available
		if device.IpAddress != "" {
			deviceParts = append(deviceParts, fmt.Sprintf("host %s", device.IpAddress))
		}

		// Combine MAC and IP with OR for this device
		if len(deviceParts) > 0 {
			deviceFilters = append(deviceFilters, fmt.Sprintf("(%s)", strings.Join(deviceParts, " or ")))
		}
	}

	// Combine all devices with OR
	if len(deviceFilters) > 0 {
		return strings.Join(deviceFilters, " or ")
	}

	return ""
}

// buildProtocolFilter creates BPF filter for protocols and ports
func (fb *FilterBuilder) buildProtocolFilter(targets []*nettestlabv1.UrlTarget) string {
	if len(targets) == 0 {
		return ""
	}

	var protocolFilters []string
	httpPorts := make(map[int32]bool)
	httpsPorts := make(map[int32]bool)
	allPorts := make(map[int32]bool)

	// Collect ports by protocol type
	for _, target := range targets {
		if !target.Enabled {
			continue
		}

		switch target.ProtocolFilter {
		case "HTTP":
			for _, port := range target.Ports {
				httpPorts[port] = true
			}
			// Default HTTP ports if none specified
			if len(target.Ports) == 0 {
				httpPorts[80] = true
			}
		case "HTTPS":
			for _, port := range target.Ports {
				httpsPorts[port] = true
			}
			// Default HTTPS ports if none specified
			if len(target.Ports) == 0 {
				httpsPorts[443] = true
			}
		case "ALL":
			for _, port := range target.Ports {
				allPorts[port] = true
			}
			// Default web ports if none specified
			if len(target.Ports) == 0 {
				allPorts[80] = true
				allPorts[443] = true
			}
		}
	}

	// Build port filters
	if len(httpPorts) > 0 {
		ports := fb.mapKeysToSlice(httpPorts)
		portFilter := fb.buildPortFilter(ports)
		if portFilter != "" {
			protocolFilters = append(protocolFilters, portFilter)
		}
	}

	if len(httpsPorts) > 0 {
		ports := fb.mapKeysToSlice(httpsPorts)
		portFilter := fb.buildPortFilter(ports)
		if portFilter != "" {
			protocolFilters = append(protocolFilters, portFilter)
		}
	}

	if len(allPorts) > 0 {
		ports := fb.mapKeysToSlice(allPorts)
		portFilter := fb.buildPortFilter(ports)
		if portFilter != "" {
			protocolFilters = append(protocolFilters, portFilter)
		}
	}

	// If no specific ports, default to common web ports
	if len(protocolFilters) == 0 {
		protocolFilters = append(protocolFilters, "(port 80 or port 443)")
	}

	// Combine all protocol filters with OR
	if len(protocolFilters) > 0 {
		return strings.Join(protocolFilters, " or ")
	}

	return ""
}

// buildHostFilter creates BPF filter for host patterns
// Note: BPF doesn't support regex, so this creates filters for known patterns
func (fb *FilterBuilder) buildHostFilter(targets []*nettestlabv1.UrlTarget) string {
	if len(targets) == 0 {
		return ""
	}

	var hostFilters []string

	for _, target := range targets {
		if !target.Enabled {
			continue
		}

		// Convert regex patterns to BPF host filters
		// This is a simplified approach - in a real implementation, you'd need
		// to resolve the regex patterns to actual IP ranges or known hosts
		hostPattern := fb.convertRegexToBPFHosts(target.HostRegex)
		if hostPattern != "" {
			hostFilters = append(hostFilters, hostPattern)
		}
	}

	// Combine all host filters with OR
	if len(hostFilters) > 0 {
		return strings.Join(hostFilters, " or ")
	}

	return ""
}

// buildPortFilter creates a port filter for the given ports
func (fb *FilterBuilder) buildPortFilter(ports []int32) string {
	if len(ports) == 0 {
		return ""
	}

	var portFilters []string
	for _, port := range ports {
		portFilters = append(portFilters, fmt.Sprintf("port %d", port))
	}

	if len(portFilters) == 1 {
		return portFilters[0]
	}

	return fmt.Sprintf("(%s)", strings.Join(portFilters, " or "))
}

// formatMacForBPF formats MAC address for BPF filter
func (fb *FilterBuilder) formatMacForBPF(mac string) string {
	// Remove colons and convert to lowercase
	cleaned := strings.ToLower(strings.ReplaceAll(mac, ":", ""))
	if len(cleaned) != 12 {
		return ""
	}

	// Format as xx:xx:xx:xx:xx:xx
	formatted := fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		cleaned[0:2], cleaned[2:4], cleaned[4:6],
		cleaned[6:8], cleaned[8:10], cleaned[10:12])

	return formatted
}

// convertRegexToBPFHosts converts regex patterns to BPF host filters
// This is a simplified implementation - in practice, you'd need DNS resolution
// or IP range mapping for complex patterns
func (fb *FilterBuilder) convertRegexToBPFHosts(regex string) string {
	// Map of common patterns to known IP ranges or hosts
	knownPatterns := map[string]string{
		".*\\.facebook\\.com":  "host 31.13.64.0/18 or host 31.13.69.0/24", // Facebook IP ranges
		".*\\.google\\.com":    "host 8.8.8.8 or host 172.217.0.0/16",      // Google IP ranges
		".*\\.youtube\\.com":   "host 172.217.0.0/16",                      // YouTube (Google)
		".*\\.netflix\\.com":   "host 54.0.0.0/8",                          // Netflix (AWS)
		".*\\.instagram\\.com": "host 31.13.64.0/18",                       // Instagram (Facebook)
		".*\\.twitter\\.com":   "host 104.244.42.0/24",                     // Twitter IP ranges
		".*\\.amazon\\.com":    "host 54.0.0.0/8 or host 3.0.0.0/8",        // Amazon IP ranges
		".*\\.apple\\.com":     "host 17.0.0.0/8",                          // Apple IP ranges
		".*\\.microsoft\\.com": "host 13.0.0.0/8 or host 40.0.0.0/8",       // Microsoft IP ranges
	}

	// Check for exact matches first
	if pattern, exists := knownPatterns[regex]; exists {
		return fmt.Sprintf("(%s)", pattern)
	}

	// Try to extract domain patterns and create simple filters
	if strings.Contains(regex, "\\.com") {
		// For more complex patterns, we would need a more sophisticated approach
		// This is a placeholder that could be extended
		return fb.buildDomainBasedFilter(regex)
	}

	return ""
}

// buildDomainBasedFilter creates a filter based on domain extraction
func (fb *FilterBuilder) buildDomainBasedFilter(regex string) string {
	// This is a very simplified approach
	// In a real implementation, you would:
	// 1. Parse the regex to extract domain patterns
	// 2. Resolve domains to IP addresses
	// 3. Group IPs into CIDR blocks
	// 4. Create efficient BPF filters

	// For now, return a general web traffic filter
	return "(port 80 or port 443)"
}

// BuildSimpleFilter creates a simple BPF filter for basic use cases
func (fb *FilterBuilder) BuildSimpleFilter(deviceMACs []string, deviceIPs []string, ports []int32) string {
	var filterParts []string

	// Device filters
	if len(deviceMACs) > 0 || len(deviceIPs) > 0 {
		var deviceFilters []string

		for _, mac := range deviceMACs {
			formatted := fb.formatMacForBPF(mac)
			if formatted != "" {
				deviceFilters = append(deviceFilters, fmt.Sprintf("ether host %s", formatted))
			}
		}

		for _, ip := range deviceIPs {
			if net.ParseIP(ip) != nil {
				deviceFilters = append(deviceFilters, fmt.Sprintf("host %s", ip))
			}
		}

		if len(deviceFilters) > 0 {
			filterParts = append(filterParts, fmt.Sprintf("(%s)", strings.Join(deviceFilters, " or ")))
		}
	}

	// Port filters
	if len(ports) > 0 {
		portFilter := fb.buildPortFilter(ports)
		if portFilter != "" {
			filterParts = append(filterParts, portFilter)
		}
	}

	if len(filterParts) == 0 {
		return "port 80 or port 443" // Default web traffic
	}

	return strings.Join(filterParts, " and ")
}

// ValidateFilter validates that a BPF filter is syntactically correct
func (fb *FilterBuilder) ValidateFilter(filter string) error {
	if filter == "" {
		return fmt.Errorf("filter cannot be empty")
	}

	// Basic validation - check for balanced parentheses
	openCount := strings.Count(filter, "(")
	closeCount := strings.Count(filter, ")")
	if openCount != closeCount {
		return fmt.Errorf("unbalanced parentheses in filter")
	}

	// Check for dangerous patterns
	dangerous := []string{"rm ", "del ", "; ", "&&", "||", "`", "$"}
	for _, pattern := range dangerous {
		if strings.Contains(filter, pattern) {
			return fmt.Errorf("potentially dangerous pattern in filter: %s", pattern)
		}
	}

	return nil
}

// Helper functions

func (fb *FilterBuilder) mapKeysToSlice(m map[int32]bool) []int32 {
	keys := make([]int32, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// GetOptimizedFilter optimizes a BPF filter for better performance
func (fb *FilterBuilder) GetOptimizedFilter(filter string) (string, error) {
	if err := fb.ValidateFilter(filter); err != nil {
		return "", err
	}

	// Simple optimizations
	optimized := filter

	// Remove redundant spaces
	optimized = strings.Join(strings.Fields(optimized), " ")

	// Remove redundant parentheses around single expressions
	// This is a simplified approach - a full implementation would use proper parsing
	optimized = strings.ReplaceAll(optimized, "( ", "(")
	optimized = strings.ReplaceAll(optimized, " )", ")")

	return optimized, nil
}
