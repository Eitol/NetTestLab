package device

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// Discovery handles automatic device detection via ARP and DHCP
type Discovery struct {
	arpTablePath  string
	dhcpLeasePath string
	vendorLookup  *VendorLookup
}

// NewDiscovery creates a new device discovery instance
func NewDiscovery() *Discovery {
	return &Discovery{
		arpTablePath:  "/proc/net/arp",
		dhcpLeasePath: "/var/lib/dhcp/dhcpd.leases",
		vendorLookup:  NewVendorLookup(),
	}
}

// DetectedDevice represents a device found through network discovery
type DetectedDevice struct {
	MacAddress string
	IPAddress  string
	Hostname   string
	Vendor     string
	Interface  string
	LastSeen   time.Time
}

// ScanConnectedDevices performs a full scan of connected devices
func (d *Discovery) ScanConnectedDevices() ([]*DetectedDevice, error) {
	// Get devices from WiFi stations only (like LuCI Associated Stations)
	wifiDevices, err := d.scanWiFiStations()
	if err != nil {
		return nil, fmt.Errorf("failed to scan WiFi stations: %w", err)
	}

	// Return only WiFi devices with vendor and hostname resolution
	var devices []*DetectedDevice
	for _, device := range wifiDevices {
		// Lookup vendor information
		device.Vendor = d.vendorLookup.LookupVendor(device.MacAddress)

		// Try to resolve hostname if not available
		if device.Hostname == "" {
			device.Hostname = d.resolveHostname(device.IPAddress)
		}

		devices = append(devices, device)
	}

	return devices, nil
}

// resolveHostname attempts to resolve hostname via reverse DNS
func (d *Discovery) resolveHostname(ipAddress string) string {
	if ipAddress == "" {
		return ""
	}
	
	names, err := net.LookupAddr(ipAddress)
	if err != nil || len(names) == 0 {
		return ""
	}

	hostname := names[0]
	// Remove trailing dot if present
	hostname = strings.TrimSuffix(hostname, ".")

	return hostname
}

// isValidMACAddress validates MAC address format
func (d *Discovery) isValidMACAddress(mac string) bool {
	// Check if it matches standard MAC format
	macRegex := regexp.MustCompile(`^([0-9a-f]{2}[:-]){5}([0-9a-f]{2})$`)
	return macRegex.MatchString(strings.ToLower(mac))
}

// scanWiFiStations scans for WiFi connected devices using iw command
func (d *Discovery) scanWiFiStations() ([]*DetectedDevice, error) {
	var devices []*DetectedDevice

	// Get WiFi interfaces
	interfaces, err := d.getWiFiInterfaces()
	if err != nil {
		return devices, err
	}

	// Scan each WiFi interface for connected stations
	for _, iface := range interfaces {
		stationDevices, err := d.scanWiFiInterface(iface)
		if err != nil {
			fmt.Printf("Warning: failed to scan WiFi interface %s: %v\n", iface, err)
			continue
		}
		devices = append(devices, stationDevices...)
	}

	return devices, nil
}

// getWiFiInterfaces gets list of WiFi interfaces that are in AP mode
func (d *Discovery) getWiFiInterfaces() ([]string, error) {
	// Run "iw dev" to get interface list
	cmd := exec.Command("iw", "dev")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run iw dev: %w", err)
	}

	var interfaces []string
	lines := strings.Split(string(output), "\n")
	var currentInterface string
	isAP := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Look for interface name
		if strings.HasPrefix(line, "Interface ") {
			currentInterface = strings.TrimPrefix(line, "Interface ")
			isAP = false
		}
		
		// Check if interface is in AP mode
		if strings.Contains(line, "type AP") {
			isAP = true
		}
		
		// If we found an AP interface, add it to the list
		if currentInterface != "" && isAP && !strings.Contains(line, "Interface ") {
			// Check if this interface is already added
			found := false
			for _, existing := range interfaces {
				if existing == currentInterface {
					found = true
					break
				}
			}
			if !found {
				interfaces = append(interfaces, currentInterface)
			}
		}
	}

	return interfaces, nil
}

// scanWiFiInterface scans a specific WiFi interface for connected stations
func (d *Discovery) scanWiFiInterface(interfaceName string) ([]*DetectedDevice, error) {
	// Run "iw dev <interface> station dump"
	cmd := exec.Command("iw", "dev", interfaceName, "station", "dump")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run iw station dump: %w", err)
	}

	var devices []*DetectedDevice
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Look for station MAC addresses
		if strings.HasPrefix(line, "Station ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				macAddress := parts[1]
				
				// Validate MAC address format
				if d.isValidMACAddress(macAddress) {
					// Try to get IP from ARP table for this specific MAC
					ipAddress := d.getIPFromARP(macAddress)
					
					device := &DetectedDevice{
						MacAddress: strings.ToLower(macAddress),
						IPAddress:  ipAddress,
						Interface:  "WiFi", // Mark as WiFi interface
						LastSeen:   time.Now(),
					}
					devices = append(devices, device)
				}
			}
		}
	}

	return devices, nil
}

// getIPFromARP tries to find IP address for a MAC address from ARP table
func (d *Discovery) getIPFromARP(macAddress string) string {
	file, err := os.Open(d.arpTablePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	
	// Skip header line
	if scanner.Scan() {
		// Header: IP address HW type Flags HW address Mask Device
	}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			arpMac := strings.ToLower(fields[3])
			if arpMac == strings.ToLower(macAddress) {
				return fields[0] // IP address
			}
		}
	}

	return ""
}

// StartPeriodicScan starts a goroutine that scans for devices periodically
func (d *Discovery) StartPeriodicScan(interval time.Duration, callback func([]*DetectedDevice)) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			devices, err := d.ScanConnectedDevices()
			if err != nil {
				fmt.Printf("Error scanning devices: %v\n", err)
				continue
			}
			if callback != nil {
				callback(devices)
			}
		}
	}()
}
