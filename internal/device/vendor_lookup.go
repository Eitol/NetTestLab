package device

import (
	"strings"
)

// VendorLookup provides MAC address to vendor mapping
type VendorLookup struct {
	vendorMap map[string]string
}

// NewVendorLookup creates a new vendor lookup instance with common vendors
func NewVendorLookup() *VendorLookup {
	return &VendorLookup{
		vendorMap: getCommonVendors(),
	}
}

// LookupVendor returns the vendor name for a given MAC address
func (v *VendorLookup) LookupVendor(macAddress string) string {
	if macAddress == "" {
		return "Unknown"
	}

	// Normalize MAC address and extract OUI (first 3 octets)
	mac := strings.ToLower(strings.ReplaceAll(macAddress, ":", ""))
	if len(mac) < 6 {
		return "Unknown"
	}

	oui := mac[:6]

	// Check direct OUI match
	if vendor, exists := v.vendorMap[oui]; exists {
		return vendor
	}

	return "Unknown"
}

// getCommonVendors returns a map of common MAC OUI prefixes to vendor names
// This is a subset of the most common vendors found in home/office networks
func getCommonVendors() map[string]string {
	return map[string]string{
		// Apple
		"001cb3": "Apple",
		"0017f2": "Apple",
		"0019e3": "Apple",
		"001ec2": "Apple",
		"0023df": "Apple",
		"002332": "Apple",
		"002436": "Apple",
		"0025bc": "Apple",
		"002608": "Apple",
		"0026bb": "Apple",
		"002711": "Apple",
		"0050e4": "Apple",
		"5c95ae": "Apple",
		"704d7b": "Apple",
		"7c04d0": "Apple",
		"bc6778": "Apple",
		"f45c89": "Apple",
		"f8e61a": "Apple",

		// Samsung
		"0015b9": "Samsung",
		"001632": "Samsung",
		"0018af": "Samsung",
		"001d25": "Samsung",
		"002454": "Samsung",
		"0026cc": "Samsung",
		"002566": "Samsung",
		"0027e3": "Samsung",
		"34e2fd": "Samsung",
		"40b0fa": "Samsung",
		"5c0a5b": "Samsung",
		"88c663": "Samsung",
		"c8214e": "Samsung",
		"ec59e7": "Samsung",

		// Huawei
		"001e10": "Huawei",
		"002264": "Huawei",
		"0025df": "Huawei",
		"0027c6": "Huawei",
		"4ccc6a": "Huawei",
		"54259c": "Huawei",
		"648b96": "Huawei",
		"843a4b": "Huawei",
		"ac853d": "Huawei",
		"c83ae6": "Huawei",

		// Xiaomi
		"28ff3c": "Xiaomi",
		"341c85": "Xiaomi",
		"3c28d2": "Xiaomi",
		"50d2f5": "Xiaomi",
		"64b473": "Xiaomi",
		"68df7a": "Xiaomi",
		"78442e": "Xiaomi",
		"7c1dd9": "Xiaomi",
		"8c85ac": "Xiaomi",
		"8cf5a3": "Xiaomi",
		"a0e4cb": "Xiaomi",
		"a8be27": "Xiaomi",
		"f8a45f": "Xiaomi",

		// OnePlus
		"ac37e4": "OnePlus",
		"d0876b": "OnePlus",

		// Google
		"5e7b8a": "Google",
		"f4f5e8": "Google",

		// LG
		"001c62": "LG Electronics",
		"0021fe": "LG Electronics",
		"002140": "LG Electronics",
		"f8bb7b": "LG Electronics",

		// Sony
		"001599": "Sony",
		"001a8a": "Sony",
		"001b59": "Sony",
		"0023f1": "Sony",

		// Intel (WiFi cards)
		"001cc0": "Intel",
		"0024d7": "Intel",
		"7085c2": "Intel",
		"9cade5": "Intel",

		// Broadcom (WiFi cards)
		"001018": "Broadcom",
		"002269": "Broadcom",
		"bc1485": "Broadcom",

		// Qualcomm (WiFi cards)
		"008048": "Qualcomm",
		"20087a": "Qualcomm",

		// TP-Link
		"001d0f": "TP-Link",
		"14cf92": "TP-Link",
		"98da4a": "TP-Link",
		"a42bb0": "TP-Link",
		"c4e9ba": "TP-Link",
		"f4ec38": "TP-Link",

		// Netgear
		"001b2f": "Netgear",
		"0026f2": "Netgear",
		"28c68e": "Netgear",
		"44947a": "Netgear",
		"84d6d0": "Netgear",

		// D-Link
		"001195": "D-Link",
		"001346": "D-Link",
		"0015e9": "D-Link",
		"141dd7": "D-Link",
		"1c7ee5": "D-Link",

		// ASUS
		"001e8c": "ASUSTek Computer",
		"0026ab": "ASUSTek Computer",
		"107b44": "ASUSTek Computer",
		"bc5ff4": "ASUSTek Computer",

		// Microsoft
		"0050f2": "Microsoft",
		"7c1e52": "Microsoft",
		"e83935": "Microsoft",

		// Amazon
		"44650d": "Amazon Technologies",
		"747548": "Amazon Technologies",
		"f0d1a9": "Amazon Technologies",

		// Nintendo
		"0009bf": "Nintendo",
		"001656": "Nintendo",
		"0019fd": "Nintendo",
		"0025a0": "Nintendo",
		"64b5c6": "Nintendo",

		// Sony PlayStation
		"0002c7": "Sony Interactive Entertainment",
		"c89e43": "Sony Interactive Entertainment",

		// Generic/Unknown patterns
		"000000": "Unknown",
		"ffffff": "Unknown",
	}
}

// AddVendor adds a new vendor mapping to the lookup table
func (v *VendorLookup) AddVendor(oui, vendor string) {
	oui = strings.ToLower(strings.ReplaceAll(oui, ":", ""))
	if len(oui) == 6 {
		v.vendorMap[oui] = vendor
	}
}

// GetAllVendors returns all known vendor mappings
func (v *VendorLookup) GetAllVendors() map[string]string {
	result := make(map[string]string)
	for oui, vendor := range v.vendorMap {
		result[oui] = vendor
	}
	return result
}
