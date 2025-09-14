package network

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	nettestlabv1 "github.com/nettestlab/nettestlab/api/nettestlab/v1"
)

// Controller manages network traffic control
type Controller struct {
	mu               sync.RWMutex
	activeConditions map[string]*nettestlabv1.NetworkConditions
	interfaces       map[string]*InterfaceInfo
}

// InterfaceInfo holds information about a network interface
type InterfaceInfo struct {
	Name        string
	Type        nettestlabv1.InterfaceType
	IsUp        bool
	IPAddresses []string
}

// NewController creates a new network controller
func NewController() (*Controller, error) {
	c := &Controller{
		activeConditions: make(map[string]*nettestlabv1.NetworkConditions),
		interfaces:       make(map[string]*InterfaceInfo),
	}

	// Initialize interface information
	if err := c.refreshInterfaces(); err != nil {
		// On non-Linux systems, create mock interfaces for testing
		if runtime.GOOS != "linux" {
			c.createMockInterfaces()
		} else {
			return nil, fmt.Errorf("failed to initialize interfaces: %w", err)
		}
	}

	return c, nil
}

// ApplyConditions applies network conditions to an interface
// ApplyConditions applies network conditions to an interface (supports "wifi" as interface name)
func (c *Controller) ApplyConditions(iface string, conditions *nettestlabv1.NetworkConditions, direction nettestlabv1.TrafficDirection) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Resolve interface name(s) - handles "wifi" special case
	interfaces, err := c.resolveInterfaceName(iface)
	if err != nil {
		return err
	}

	// Apply conditions to all resolved interfaces
	var lastError error
	successCount := 0

	for _, resolvedIface := range interfaces {
		// On non-Linux systems, just store the conditions without applying tc rules
		if runtime.GOOS != "linux" {
			c.activeConditions[resolvedIface] = conditions
			successCount++
			continue
		}

		// Reset existing conditions first
		if err := c.resetConditionsUnsafe(resolvedIface, direction); err != nil {
			lastError = fmt.Errorf("failed to reset existing conditions on %s: %w", resolvedIface, err)
			continue
		}

		// Apply new conditions
		if err := c.applyConditionsUnsafe(resolvedIface, conditions, direction); err != nil {
			lastError = fmt.Errorf("failed to apply conditions to %s: %w", resolvedIface, err)
			continue
		}

		// Store active conditions
		c.activeConditions[resolvedIface] = conditions
		successCount++
	}

	// If we succeeded on at least one interface, consider it a success
	if successCount > 0 {
		return nil
	}

	// If all failed, return the last error
	if lastError != nil {
		return lastError
	}

	return fmt.Errorf("no interfaces processed")
}

// ResetConditions removes all conditions from an interface (supports "wifi" as interface name)
func (c *Controller) ResetConditions(iface string, direction nettestlabv1.TrafficDirection) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Resolve interface name(s) - handles "wifi" special case
	interfaces, err := c.resolveInterfaceName(iface)
	if err != nil {
		return err
	}

	// Reset conditions on all resolved interfaces
	var lastError error
	successCount := 0

	for _, resolvedIface := range interfaces {
		if err := c.resetConditionsUnsafe(resolvedIface, direction); err != nil {
			lastError = fmt.Errorf("failed to reset conditions on %s: %w", resolvedIface, err)
		} else {
			successCount++
		}
	}

	// If we succeeded on at least one interface, consider it a success
	if successCount > 0 {
		return nil
	}

	// If all failed, return the last error
	if lastError != nil {
		return lastError
	}

	return fmt.Errorf("no interfaces processed")
}

// GetConditions returns current conditions for an interface
func (c *Controller) GetConditions(iface string) (*nettestlabv1.NetworkConditions, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	conditions, exists := c.activeConditions[iface]
	return conditions, exists
}

// GetInterfaces returns all available network interfaces
func (c *Controller) GetInterfaces() map[string]*InterfaceInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]*InterfaceInfo)
	for name, info := range c.interfaces {
		result[name] = info
	}
	return result
}

// RefreshInterfaces updates the interface information
func (c *Controller) RefreshInterfaces() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.refreshInterfaces()
}

// createMockInterfaces creates mock interfaces for testing on non-Linux systems
func (c *Controller) createMockInterfaces() {
	c.interfaces = map[string]*InterfaceInfo{
		"eth0": {
			Name:        "eth0",
			Type:        nettestlabv1.InterfaceType_INTERFACE_TYPE_ETHERNET,
			IsUp:        true,
			IPAddresses: []string{"192.168.1.100"},
		},
		"wlan0": {
			Name:        "wlan0",
			Type:        nettestlabv1.InterfaceType_INTERFACE_TYPE_WIRELESS,
			IsUp:        true,
			IPAddresses: []string{"192.168.1.101"},
		},
		"lo": {
			Name:        "lo",
			Type:        nettestlabv1.InterfaceType_INTERFACE_TYPE_LOOPBACK,
			IsUp:        true,
			IPAddresses: []string{"127.0.0.1"},
		},
	}
}

// applyConditionsUnsafe applies conditions without locking (internal use)
func (c *Controller) applyConditionsUnsafe(iface string, conditions *nettestlabv1.NetworkConditions, direction nettestlabv1.TrafficDirection) error {
	// Skip actual tc commands on non-Linux systems
	if runtime.GOOS != "linux" {
		// Store fake conditions for testing
		c.activeConditions[iface] = conditions
		return nil
	}

	// First, clear any existing qdisc on the interface
	c.resetConditionsUnsafe(iface, direction)

	var cmds [][]string

	// Build tc commands based on conditions
	if conditions.Latency != nil && conditions.Latency.Enabled {
		cmds = append(cmds, c.buildLatencyCommand(iface, conditions.Latency, direction)...)
	}

	if conditions.PacketLoss != nil && conditions.PacketLoss.Enabled {
		cmds = append(cmds, c.buildPacketLossCommand(iface, conditions.PacketLoss, direction)...)
	}

	if conditions.Bandwidth != nil && conditions.Bandwidth.Enabled {
		cmds = append(cmds, c.buildBandwidthCommand(iface, conditions.Bandwidth, direction)...)
	}

	if conditions.Jitter != nil && conditions.Jitter.Enabled {
		cmds = append(cmds, c.buildJitterCommand(iface, conditions.Jitter, direction)...)
	}

	if conditions.Corruption != nil && conditions.Corruption.Enabled {
		cmds = append(cmds, c.buildCorruptionCommand(iface, conditions.Corruption, direction)...)
	}

	// Execute all commands
	for _, cmd := range cmds {
		if err := c.executeCommand(cmd); err != nil {
			return fmt.Errorf("failed to execute command %v: %w", cmd, err)
		}
	}

	// Store active conditions
	c.activeConditions[iface] = conditions

	return nil
}

// resetConditionsUnsafe removes conditions without locking (internal use)
func (c *Controller) resetConditionsUnsafe(iface string, direction nettestlabv1.TrafficDirection) error {
	// Skip actual tc commands on non-Linux systems
	if runtime.GOOS != "linux" {
		delete(c.activeConditions, iface)
		return nil
	}

	// First try to remove all qdiscs completely (this handles any existing qdiscs)
	// We try multiple times to ensure complete cleanup
	for attempt := 0; attempt < 3; attempt++ {
		cmds := [][]string{
			{"tc", "qdisc", "del", "dev", iface, "root"},    // Remove all root qdiscs
			{"tc", "qdisc", "del", "dev", iface, "ingress"}, // Remove ingress qdisc
		}

		for _, cmd := range cmds {
			if err := c.executeCommand(cmd); err != nil {
				// Log but don't fail on expected errors (like "qdisc not found")
				if !strings.Contains(err.Error(), "RTNETLINK answers: No such file or directory") &&
					!strings.Contains(err.Error(), "RTNETLINK answers: Invalid argument") {
					fmt.Printf("Warning: qdisc cleanup failed for %s: %v\n", iface, err)
				}
			}
		}

		// Check if cleanup was successful by verifying no custom qdiscs remain
		checkCmd := exec.Command("tc", "qdisc", "show", "dev", iface)
		output, err := checkCmd.Output()
		if err == nil {
			outputStr := string(output)
			// If only noqueue or default qdiscs remain, we're done
			if !strings.Contains(outputStr, "netem") &&
				!strings.Contains(outputStr, "tbf") &&
				!strings.Contains(outputStr, "handle 1:") {
				break
			}
		}

		// Wait briefly before retry
		time.Sleep(100 * time.Millisecond)
	}

	// Remove from active conditions
	delete(c.activeConditions, iface)

	return nil
}

// buildLatencyCommand creates tc commands for latency
func (c *Controller) buildLatencyCommand(iface string, config *nettestlabv1.LatencyConfig, direction nettestlabv1.TrafficDirection) [][]string {
	delay := fmt.Sprintf("%dms", config.DelayMs)

	return [][]string{
		{"tc", "qdisc", "add", "dev", iface, "root", "handle", "1:", "netem", "delay", delay},
	}
}

// buildPacketLossCommand creates tc commands for packet loss
func (c *Controller) buildPacketLossCommand(iface string, config *nettestlabv1.PacketLossConfig, direction nettestlabv1.TrafficDirection) [][]string {
	loss := fmt.Sprintf("%.2f%%", config.Percentage)

	return [][]string{
		{"tc", "qdisc", "add", "dev", iface, "root", "handle", "1:", "netem", "loss", loss},
	}
}

// buildBandwidthCommand creates tc commands for bandwidth limiting
func (c *Controller) buildBandwidthCommand(iface string, config *nettestlabv1.BandwidthConfig, direction nettestlabv1.TrafficDirection) [][]string {
	// Convert bps to appropriate unit for tc
	downloadRate := formatBandwidth(config.DownloadBps)

	return [][]string{
		{"tc", "qdisc", "add", "dev", iface, "root", "handle", "1:", "tbf", "rate", downloadRate, "burst", "32kbit", "latency", "400ms"},
	}
}

// buildJitterCommand creates tc commands for jitter
func (c *Controller) buildJitterCommand(iface string, config *nettestlabv1.JitterConfig, direction nettestlabv1.TrafficDirection) [][]string {
	variation := fmt.Sprintf("%dms", config.VariationMs)

	return [][]string{
		{"tc", "qdisc", "add", "dev", iface, "root", "handle", "1:", "netem", "delay", "0ms", variation},
	}
}

// buildCorruptionCommand creates tc commands for packet corruption
func (c *Controller) buildCorruptionCommand(iface string, config *nettestlabv1.CorruptionConfig, direction nettestlabv1.TrafficDirection) [][]string {
	corrupt := fmt.Sprintf("%.2f%%", config.Percentage)

	return [][]string{
		{"tc", "qdisc", "add", "dev", iface, "root", "handle", "1:", "netem", "corrupt", corrupt},
	}
}

// executeCommand runs a shell command
func (c *Controller) executeCommand(cmd []string) error {
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %s, output: %s", err, string(output))
	}
	return nil
}

// refreshInterfaces updates the internal interface list
func (c *Controller) refreshInterfaces() error {
	// Use different commands based on OS
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("ifconfig")
	} else {
		cmd = exec.Command("ip", "link", "show")
	}

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get interface list: %w", err)
	}

	var interfaces map[string]*InterfaceInfo
	if runtime.GOOS == "darwin" {
		interfaces = parseIfconfigOutput(string(output))
	} else {
		interfaces = parseInterfaceList(string(output))
	}

	c.interfaces = interfaces
	return nil
}

// parseIfconfigOutput parses ifconfig output on macOS
func parseIfconfigOutput(output string) map[string]*InterfaceInfo {
	interfaces := make(map[string]*InterfaceInfo)
	lines := strings.Split(output, "\n")

	var currentInterface *InterfaceInfo
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// New interface starts with interface name followed by ':'
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			if len(parts) > 0 {
				name := parts[0]
				interfaceType := determineInterfaceType(name)
				isUp := strings.Contains(line, "UP")

				currentInterface = &InterfaceInfo{
					Name:        name,
					Type:        interfaceType,
					IsUp:        isUp,
					IPAddresses: []string{},
				}
				interfaces[name] = currentInterface
			}
		} else if currentInterface != nil && strings.HasPrefix(line, "inet ") {
			// Extract IP address
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ip := parts[1]
				currentInterface.IPAddresses = append(currentInterface.IPAddresses, ip)
			}
		}
	}

	return interfaces
}

// parseInterfaceList parses ip link show output
func parseInterfaceList(output string) map[string]*InterfaceInfo {
	interfaces := make(map[string]*InterfaceInfo)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ": ") && !strings.HasPrefix(line, " ") {
			parts := strings.Split(line, ": ")
			if len(parts) >= 2 {
				name := parts[1]
				// Remove state info like <BROADCAST,MULTICAST,UP,LOWER_UP>
				if idx := strings.Index(name, " "); idx != -1 {
					name = name[:idx]
				}

				interfaceType := determineInterfaceType(name)
				isUp := strings.Contains(line, "UP")

				interfaces[name] = &InterfaceInfo{
					Name:        name,
					Type:        interfaceType,
					IsUp:        isUp,
					IPAddresses: getInterfaceIPs(name),
				}
			}
		}
	}

	return interfaces
}

// determineInterfaceType determines the type of network interface
func determineInterfaceType(name string) nettestlabv1.InterfaceType {
	switch {
	case strings.HasPrefix(name, "eth"), strings.HasPrefix(name, "en"):
		return nettestlabv1.InterfaceType_INTERFACE_TYPE_ETHERNET
	case strings.HasPrefix(name, "wlan"), strings.HasPrefix(name, "wifi"), strings.HasPrefix(name, "wl"):
		return nettestlabv1.InterfaceType_INTERFACE_TYPE_WIRELESS
	case strings.HasPrefix(name, "lo"):
		return nettestlabv1.InterfaceType_INTERFACE_TYPE_LOOPBACK
	case strings.HasPrefix(name, "br"):
		return nettestlabv1.InterfaceType_INTERFACE_TYPE_BRIDGE
	default:
		return nettestlabv1.InterfaceType_INTERFACE_TYPE_UNSPECIFIED
	}
}

// getInterfaceIPs gets IP addresses for an interface (Linux only)
func getInterfaceIPs(iface string) []string {
	if runtime.GOOS != "linux" {
		return []string{} // Skip on non-Linux systems
	}

	cmd := exec.Command("ip", "addr", "show", iface)
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var ips []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "inet ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ip := strings.Split(parts[1], "/")[0]
				ips = append(ips, ip)
			}
		}
	}

	return ips
}

// formatBandwidth converts bps to tc-compatible format
func formatBandwidth(bps uint64) string {
	if bps >= 1000000000 {
		return fmt.Sprintf("%.0fgbit", float64(bps)/1000000000)
	} else if bps >= 1000000 {
		return fmt.Sprintf("%.0fmbit", float64(bps)/1000000)
	} else if bps >= 1000 {
		return fmt.Sprintf("%.0fkbit", float64(bps)/1000)
	}
	return fmt.Sprintf("%dbit", bps)
}

// resolveInterfaceName resolves interface names, handling special cases like "wifi"
// Note: This method assumes the caller already holds the mutex
func (c *Controller) resolveInterfaceName(iface string) ([]string, error) {
	// If it's the special "wifi" keyword, find all WiFi interfaces
	if strings.ToLower(iface) == "wifi" {
		return c.findWiFiInterfacesUnsafe()
	}

	// For regular interface names, return as-is (with validation)
	if _, exists := c.interfaces[iface]; !exists {
		return nil, fmt.Errorf("interface '%s' not found", iface)
	}

	return []string{iface}, nil
}

// findWiFiInterfacesUnsafe discovers all active WiFi interfaces (without taking mutex)
func (c *Controller) findWiFiInterfacesUnsafe() ([]string, error) {
	var wifiInterfaces []string

	for name, info := range c.interfaces {
		// Check if it's a WiFi interface and is up
		if info.Type == nettestlabv1.InterfaceType_INTERFACE_TYPE_WIRELESS && info.IsUp {
			wifiInterfaces = append(wifiInterfaces, name)
		}
	}

	if len(wifiInterfaces) == 0 {
		return nil, fmt.Errorf("no active WiFi interfaces found")
	}

	return wifiInterfaces, nil
}
