package device

import (
	"bufio"
	"fmt"
	"net"
	"os"
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
	LastSeen   time.Time
}

// ScanConnectedDevices performs a full scan of connected devices
func (d *Discovery) ScanConnectedDevices() ([]*DetectedDevice, error) {
	// Get devices from ARP table
	arpDevices, err := d.scanARPTable()
	if err != nil {
		return nil, fmt.Errorf("failed to scan ARP table: %w", err)
	}

	// Get devices from DHCP leases
	dhcpDevices, err := d.scanDHCPLeases()
	if err != nil {
		// DHCP scan failure is not critical, log but continue
		fmt.Printf("Warning: failed to scan DHCP leases: %v\n", err)
	}

	// Merge ARP and DHCP data
	deviceMap := make(map[string]*DetectedDevice)

	// Add ARP devices
	for _, device := range arpDevices {
		deviceMap[device.MacAddress] = device
	}

	// Merge DHCP data
	for _, dhcpDevice := range dhcpDevices {
		if existing, exists := deviceMap[dhcpDevice.MacAddress]; exists {
			// Update hostname if available from DHCP
			if dhcpDevice.Hostname != "" {
				existing.Hostname = dhcpDevice.Hostname
			}
		} else {
			deviceMap[dhcpDevice.MacAddress] = dhcpDevice
		}
	}

	// Convert map to slice
	var devices []*DetectedDevice
	for _, device := range deviceMap {
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

// scanARPTable reads the ARP table to find active devices
func (d *Discovery) scanARPTable() ([]*DetectedDevice, error) {
	file, err := os.Open(d.arpTablePath)
	if err != nil {
		// Return empty list instead of error for non-Linux systems (development)
		return []*DetectedDevice{}, nil
	}
	defer file.Close()

	var devices []*DetectedDevice
	scanner := bufio.NewScanner(file)
	
	// Skip header line
	if scanner.Scan() {
		// Header: IP address HW type Flags HW address Mask Device
	}

	for scanner.Scan() {
		line := scanner.Text()
		device := d.parseARPLine(line)
		if device != nil {
			devices = append(devices, device)
		}
	}

	return devices, scanner.Err()
}

// parseARPLine parses a single line from /proc/net/arp
func (d *Discovery) parseARPLine(line string) *DetectedDevice {
	// Example line: "192.168.1.100 0x1 0x2 aa:bb:cc:dd:ee:ff * br-lan"
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return nil
	}

	ipAddress := fields[0]
	macAddress := fields[3]

	// Skip incomplete entries (marked with 00:00:00:00:00:00)
	if macAddress == "00:00:00:00:00:00" {
		return nil
	}

	// Validate MAC address format
	if !d.isValidMACAddress(macAddress) {
		return nil
	}

	return &DetectedDevice{
		MacAddress: strings.ToLower(macAddress),
		IPAddress:  ipAddress,
		LastSeen:   time.Now(),
	}
}

// scanDHCPLeases reads DHCP lease file to get hostname information
func (d *Discovery) scanDHCPLeases() ([]*DetectedDevice, error) {
	file, err := os.Open(d.dhcpLeasePath)
	if err != nil {
		// Return empty list instead of error for non-OpenWRT systems (development)
		return []*DetectedDevice{}, nil
	}
	defer file.Close()

	var devices []*DetectedDevice
	scanner := bufio.NewScanner(file)

	var currentLease *DetectedDevice
	leaseRegex := regexp.MustCompile(`^lease\s+(\S+)\s+{`)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Start of a new lease
		if matches := leaseRegex.FindStringSubmatch(line); matches != nil {
			currentLease = &DetectedDevice{
				IPAddress: matches[1],
				LastSeen:  time.Now(),
			}
			continue
		}
		
		if currentLease == nil {
			continue
		}
		
		// Parse lease properties
		if strings.HasPrefix(line, "hardware ethernet") {
			mac := strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(line, "hardware ethernet")), ";")
			currentLease.MacAddress = strings.ToLower(mac)
		} else if strings.HasPrefix(line, "client-hostname") {
			hostname := strings.Trim(strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(line, "client-hostname")), ";"), "\"")
			currentLease.Hostname = hostname
		} else if line == "}" && currentLease.MacAddress != "" {
			// End of lease, add to devices if valid
			if d.isValidMACAddress(currentLease.MacAddress) {
				devices = append(devices, currentLease)
			}
			currentLease = nil
		}
	}

	return devices, scanner.Err()
}

// resolveHostname attempts to resolve hostname via reverse DNS
func (d *Discovery) resolveHostname(ipAddress string) string {
	names, err := net.LookupAddr(ipAddress)
	if err != nil || len(names) == 0 {
		return ""
	}
	
	hostname := names[0]
	// Remove trailing dot if present
	if strings.HasSuffix(hostname, ".") {
		hostname = hostname[:len(hostname)-1]
	}
	
	return hostname
}

// isValidMACAddress validates MAC address format
func (d *Discovery) isValidMACAddress(mac string) bool {
	// Check if it matches standard MAC format
	macRegex := regexp.MustCompile(`^([0-9a-f]{2}[:-]){5}([0-9a-f]{2})$`)
	return macRegex.MatchString(strings.ToLower(mac))
}

// StartPeriodicScan starts a goroutine that scans for devices periodically
func (d *Discovery) StartPeriodicScan(interval time.Duration, callback func([]*DetectedDevice)) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				devices, err := d.ScanConnectedDevices()
				if err != nil {
					fmt.Printf("Error scanning devices: %v\n", err)
					continue
				}
				if callback != nil {
					callback(devices)
				}
			}
		}
	}()
}