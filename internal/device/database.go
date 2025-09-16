package device

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Database handles device storage operations
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new device database instance
func NewDatabase(dataDir string) (*Database, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open SQLite database
	dbPath := filepath.Join(dataDir, "devices.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database := &Database{db: db}

	// Initialize database schema
	if err := database.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// initSchema creates the device table if it doesn't exist
func (d *Database) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS devices (
		id TEXT PRIMARY KEY,
		mac_address TEXT UNIQUE NOT NULL,
		ip_address TEXT,
		hostname TEXT,
		device_name TEXT,
		device_model TEXT,
		os_version TEXT,
		app_version TEXT,
		connection_status INTEGER DEFAULT 0,
		registered BOOLEAN DEFAULT FALSE,
		first_seen DATETIME,
		last_seen DATETIME,
		registered_at DATETIME,
		vendor TEXT,
		previous_ips TEXT, -- JSON array of previous IPs
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_devices_mac ON devices(mac_address);
	CREATE INDEX IF NOT EXISTS idx_devices_connection_status ON devices(connection_status);
	CREATE INDEX IF NOT EXISTS idx_devices_registered ON devices(registered);
	CREATE INDEX IF NOT EXISTS idx_devices_last_seen ON devices(last_seen);
	`

	_, err := d.db.Exec(query)
	return err
}

// DeviceRow represents a device record in the database
type DeviceRow struct {
	ID               string
	MacAddress       string
	IPAddress        *string
	Hostname         *string
	DeviceName       *string
	DeviceModel      *string
	OSVersion        *string
	AppVersion       *string
	ConnectionStatus int
	Registered       bool
	FirstSeen        *string
	LastSeen         *string
	RegisteredAt     *string
	Vendor           *string
	PreviousIPs      *string // JSON string
	CreatedAt        string
	UpdatedAt        string
}

// InsertDevice inserts a new device into the database
func (d *Database) InsertDevice(device *DeviceRow) error {
	query := `
	INSERT INTO devices (
		id, mac_address, ip_address, hostname, device_name, device_model,
		os_version, app_version, connection_status, registered, first_seen,
		last_seen, registered_at, vendor, previous_ips
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := d.db.Exec(query,
		device.ID, device.MacAddress, device.IPAddress, device.Hostname,
		device.DeviceName, device.DeviceModel, device.OSVersion, device.AppVersion,
		device.ConnectionStatus, device.Registered, device.FirstSeen,
		device.LastSeen, device.RegisteredAt, device.Vendor, device.PreviousIPs,
	)
	return err
}

// UpdateDevice updates an existing device in the database
func (d *Database) UpdateDevice(device *DeviceRow) error {
	query := `
	UPDATE devices SET
		ip_address = ?, hostname = ?, device_name = ?, device_model = ?,
		os_version = ?, app_version = ?, connection_status = ?, registered = ?,
		last_seen = ?, registered_at = ?, vendor = ?, previous_ips = ?,
		updated_at = CURRENT_TIMESTAMP
	WHERE id = ?
	`

	_, err := d.db.Exec(query,
		device.IPAddress, device.Hostname, device.DeviceName, device.DeviceModel,
		device.OSVersion, device.AppVersion, device.ConnectionStatus, device.Registered,
		device.LastSeen, device.RegisteredAt, device.Vendor, device.PreviousIPs,
		device.ID,
	)
	return err
}

// GetDeviceByID retrieves a device by its ID
func (d *Database) GetDeviceByID(id string) (*DeviceRow, error) {
	query := `
	SELECT id, mac_address, ip_address, hostname, device_name, device_model,
		   os_version, app_version, connection_status, registered, first_seen,
		   last_seen, registered_at, vendor, previous_ips, created_at, updated_at
	FROM devices WHERE id = ?
	`

	row := d.db.QueryRow(query, id)
	device := &DeviceRow{}

	err := row.Scan(
		&device.ID, &device.MacAddress, &device.IPAddress, &device.Hostname,
		&device.DeviceName, &device.DeviceModel, &device.OSVersion, &device.AppVersion,
		&device.ConnectionStatus, &device.Registered, &device.FirstSeen,
		&device.LastSeen, &device.RegisteredAt, &device.Vendor, &device.PreviousIPs,
		&device.CreatedAt, &device.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return device, err
}

// GetDeviceByMAC retrieves a device by its MAC address
func (d *Database) GetDeviceByMAC(macAddress string) (*DeviceRow, error) {
	query := `
	SELECT id, mac_address, ip_address, hostname, device_name, device_model,
		   os_version, app_version, connection_status, registered, first_seen,
		   last_seen, registered_at, vendor, previous_ips, created_at, updated_at
	FROM devices WHERE mac_address = ?
	`

	row := d.db.QueryRow(query, macAddress)
	device := &DeviceRow{}

	err := row.Scan(
		&device.ID, &device.MacAddress, &device.IPAddress, &device.Hostname,
		&device.DeviceName, &device.DeviceModel, &device.OSVersion, &device.AppVersion,
		&device.ConnectionStatus, &device.Registered, &device.FirstSeen,
		&device.LastSeen, &device.RegisteredAt, &device.Vendor, &device.PreviousIPs,
		&device.CreatedAt, &device.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return device, err
}

// ListDevices retrieves devices based on filter criteria
func (d *Database) ListDevices(filter DeviceFilter, limit, offset int) ([]*DeviceRow, error) {
	query := `
	SELECT id, mac_address, ip_address, hostname, device_name, device_model,
		   os_version, app_version, connection_status, registered, first_seen,
		   last_seen, registered_at, vendor, previous_ips, created_at, updated_at
	FROM devices
	`

	var args []interface{}
	var whereClause string

	switch filter {
	case DeviceFilterConnected:
		whereClause = "WHERE connection_status = 1"
	case DeviceFilterRegistered:
		whereClause = "WHERE registered = TRUE"
	case DeviceFilterConnectedRegistered:
		whereClause = "WHERE connection_status = 1 AND registered = TRUE"
	case DeviceFilterConnectedUnregistered:
		whereClause = "WHERE connection_status = 1 AND registered = FALSE"
	}

	if whereClause != "" {
		query += " " + whereClause
	}

	query += " ORDER BY last_seen DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
		if offset > 0 {
			query += " OFFSET ?"
			args = append(args, offset)
		}
	}

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*DeviceRow
	for rows.Next() {
		device := &DeviceRow{}
		err := rows.Scan(
			&device.ID, &device.MacAddress, &device.IPAddress, &device.Hostname,
			&device.DeviceName, &device.DeviceModel, &device.OSVersion, &device.AppVersion,
			&device.ConnectionStatus, &device.Registered, &device.FirstSeen,
			&device.LastSeen, &device.RegisteredAt, &device.Vendor, &device.PreviousIPs,
			&device.CreatedAt, &device.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}

	return devices, rows.Err()
}

// DeleteDevice removes a device from the database
func (d *Database) DeleteDevice(id string) error {
	query := "DELETE FROM devices WHERE id = ?"
	_, err := d.db.Exec(query, id)
	return err
}

// CountDevices returns the total count of devices matching the filter
func (d *Database) CountDevices(filter DeviceFilter) (int, error) {
	query := "SELECT COUNT(*) FROM devices"

	var whereClause string
	switch filter {
	case DeviceFilterConnected:
		whereClause = "WHERE connection_status = 1"
	case DeviceFilterRegistered:
		whereClause = "WHERE registered = TRUE"
	case DeviceFilterConnectedRegistered:
		whereClause = "WHERE connection_status = 1 AND registered = TRUE"
	case DeviceFilterConnectedUnregistered:
		whereClause = "WHERE connection_status = 1 AND registered = FALSE"
	}

	if whereClause != "" {
		query += " " + whereClause
	}

	var count int
	err := d.db.QueryRow(query).Scan(&count)
	return count, err
}

// DeviceFilter represents the filter types for listing devices
type DeviceFilter int

const (
	DeviceFilterAll DeviceFilter = iota
	DeviceFilterConnected
	DeviceFilterRegistered
	DeviceFilterConnectedRegistered
	DeviceFilterConnectedUnregistered
)
