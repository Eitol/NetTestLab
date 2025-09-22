package capture

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // SQLite driver
)

// Database maneja el almacenamiento persistente de capturas
type Database struct {
	db   *sql.DB
	path string
}

// CaptureRecord representa un registro de captura en la base de datos
type CaptureRecord struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`

	// Configuración
	Config    string `json:"config"` // JSON de StartCaptureRequest
	BPFFilter string `json:"bpf_filter"`
	Interface string `json:"interface"`

	// Archivos y estadísticas
	FilePaths       string `json:"file_paths"` // JSON array de rutas
	TotalSizeBytes  int64  `json:"total_size_bytes"`
	PacketsCaptured int32  `json:"packets_captured"`
	PacketsDropped  int32  `json:"packets_dropped"`
	BytesCaptured   int64  `json:"bytes_captured"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewCaptureDatabase crea una nueva base de datos para capturas
func NewCaptureDatabase(dataDir string) (*Database, error) {
	// Asegurarse que el directorio existe
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "captures.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open captures database: %w", err)
	}

	database := &Database{
		db:   db,
		path: dbPath,
	}

	if err := database.createTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return database, nil
}

// Close cierra la base de datos
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// createTables crea las tablas necesarias
func (d *Database) createTables() error {
	createCapturesTable := `
	CREATE TABLE IF NOT EXISTS captures (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'starting',
		started_at DATETIME NOT NULL,
		ended_at DATETIME,
		
		-- Configuración
		config TEXT NOT NULL,      -- JSON de StartCaptureRequest
		bpf_filter TEXT,
		interface TEXT DEFAULT 'br-lan',
		
		-- Archivos y estadísticas
		file_paths TEXT DEFAULT '[]',  -- JSON array de rutas
		total_size_bytes INTEGER DEFAULT 0,
		packets_captured INTEGER DEFAULT 0,
		packets_dropped INTEGER DEFAULT 0,
		bytes_captured INTEGER DEFAULT 0,
		
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	createIndexes := `
	CREATE INDEX IF NOT EXISTS idx_captures_status ON captures(status);
	CREATE INDEX IF NOT EXISTS idx_captures_started_at ON captures(started_at);
	CREATE INDEX IF NOT EXISTS idx_captures_name ON captures(name);
	`

	if _, err := d.db.Exec(createCapturesTable); err != nil {
		return fmt.Errorf("failed to create captures table: %w", err)
	}

	if _, err := d.db.Exec(createIndexes); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// SaveCapture guarda o actualiza un registro de captura
func (d *Database) SaveCapture(record *CaptureRecord) error {
	// Serializar config y file_paths a JSON
	configJSON, err := json.Marshal(record.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	filePathsJSON, err := json.Marshal(record.FilePaths)
	if err != nil {
		return fmt.Errorf("failed to marshal file paths: %w", err)
	}

	now := time.Now()
	record.UpdatedAt = now

	query := `
	INSERT OR REPLACE INTO captures (
		id, name, status, started_at, ended_at,
		config, bpf_filter, interface,
		file_paths, total_size_bytes, packets_captured, packets_dropped, bytes_captured,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 
		COALESCE((SELECT created_at FROM captures WHERE id = ?), ?), ?)
	`

	_, err = d.db.Exec(query,
		record.ID, record.Name, record.Status, record.StartedAt, record.EndedAt,
		string(configJSON), record.BPFFilter, record.Interface,
		string(filePathsJSON), record.TotalSizeBytes, record.PacketsCaptured, record.PacketsDropped, record.BytesCaptured,
		record.ID, now, now,
	)

	if err != nil {
		return fmt.Errorf("failed to save capture: %w", err)
	}

	return nil
}

// GetCapture obtiene un registro de captura por ID
func (d *Database) GetCapture(id string) (*CaptureRecord, error) {
	query := `
	SELECT id, name, status, started_at, ended_at,
		config, bpf_filter, interface,
		file_paths, total_size_bytes, packets_captured, packets_dropped, bytes_captured,
		created_at, updated_at
	FROM captures WHERE id = ?
	`

	var record CaptureRecord
	var endedAt *string
	var configJSON, filePathsJSON string

	err := d.db.QueryRow(query, id).Scan(
		&record.ID, &record.Name, &record.Status, &record.StartedAt, &endedAt,
		&configJSON, &record.BPFFilter, &record.Interface,
		&filePathsJSON, &record.TotalSizeBytes, &record.PacketsCaptured, &record.PacketsDropped, &record.BytesCaptured,
		&record.CreatedAt, &record.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get capture: %w", err)
	}

	// Parsear ended_at
	if endedAt != nil {
		if t, err := time.Parse(time.RFC3339, *endedAt); err == nil {
			record.EndedAt = &t
		}
	}

	// Deserializar JSON
	record.Config = configJSON
	record.FilePaths = filePathsJSON

	return &record, nil
}

// ListCaptures lista capturas con filtros opcionales
func (d *Database) ListCaptures(status string, limit int, offset int) ([]*CaptureRecord, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `
		SELECT id, name, status, started_at, ended_at,
			config, bpf_filter, interface,
			file_paths, total_size_bytes, packets_captured, packets_dropped, bytes_captured,
			created_at, updated_at
		FROM captures WHERE status = ?
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
		`
		args = []interface{}{status, limit, offset}
	} else {
		query = `
		SELECT id, name, status, started_at, ended_at,
			config, bpf_filter, interface,
			file_paths, total_size_bytes, packets_captured, packets_dropped, bytes_captured,
			created_at, updated_at
		FROM captures
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
		`
		args = []interface{}{limit, offset}
	}

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list captures: %w", err)
	}
	defer rows.Close()

	var records []*CaptureRecord
	for rows.Next() {
		var record CaptureRecord
		var endedAt *string
		var configJSON, filePathsJSON string

		err := rows.Scan(
			&record.ID, &record.Name, &record.Status, &record.StartedAt, &endedAt,
			&configJSON, &record.BPFFilter, &record.Interface,
			&filePathsJSON, &record.TotalSizeBytes, &record.PacketsCaptured, &record.PacketsDropped, &record.BytesCaptured,
			&record.CreatedAt, &record.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan capture row: %w", err)
		}

		// Parsear ended_at
		if endedAt != nil {
			if t, err := time.Parse(time.RFC3339, *endedAt); err == nil {
				record.EndedAt = &t
			}
		}

		// Deserializar JSON
		record.Config = configJSON
		record.FilePaths = filePathsJSON

		records = append(records, &record)
	}

	return records, nil
}

// DeleteCapture elimina un registro de captura
func (d *Database) DeleteCapture(id string) error {
	query := `DELETE FROM captures WHERE id = ?`
	_, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete capture: %w", err)
	}
	return nil
}

// CountCaptures cuenta capturas por estado
func (d *Database) CountCaptures(status string) (int, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `SELECT COUNT(*) FROM captures WHERE status = ?`
		args = []interface{}{status}
	} else {
		query = `SELECT COUNT(*) FROM captures`
		args = []interface{}{}
	}

	var count int
	err := d.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count captures: %w", err)
	}

	return count, nil
}

// UpdateCaptureStats actualiza las estadísticas de una captura
func (d *Database) UpdateCaptureStats(id string, stats *CaptureStatistics) error {
	query := `
	UPDATE captures SET
		packets_captured = ?,
		packets_dropped = ?,
		bytes_captured = ?,
		total_size_bytes = ?,
		updated_at = CURRENT_TIMESTAMP
	WHERE id = ?
	`

	_, err := d.db.Exec(query,
		stats.PacketsCaptured,
		stats.PacketsDropped,
		stats.BytesCaptured,
		stats.FileSizeBytes,
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to update capture stats: %w", err)
	}

	return nil
}

// UpdateCaptureStatus actualiza el estado de una captura
func (d *Database) UpdateCaptureStatus(id string, status string, endedAt *time.Time) error {
	query := `
	UPDATE captures SET
		status = ?,
		ended_at = ?,
		updated_at = CURRENT_TIMESTAMP
	WHERE id = ?
	`

	_, err := d.db.Exec(query, status, endedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update capture status: %w", err)
	}

	return nil
}

// Cleanup elimina capturas antiguas basado en criterios
func (d *Database) Cleanup(maxAge time.Duration, keepCompleted int) error {
	cutoff := time.Now().Add(-maxAge)

	// Eliminar capturas fallidas o canceladas antiguas
	deleteOldFailed := `
	DELETE FROM captures 
	WHERE (status = 'failed' OR status = 'cancelled') 
	AND updated_at < ?
	`

	_, err := d.db.Exec(deleteOldFailed, cutoff)
	if err != nil {
		return fmt.Errorf("failed to delete old failed captures: %w", err)
	}

	// Mantener solo las N capturas completadas más recientes
	if keepCompleted > 0 {
		deleteOldCompleted := `
		DELETE FROM captures 
		WHERE status = 'completed' 
		AND id NOT IN (
			SELECT id FROM captures 
			WHERE status = 'completed' 
			ORDER BY ended_at DESC 
			LIMIT ?
		)
		`

		_, err = d.db.Exec(deleteOldCompleted, keepCompleted)
		if err != nil {
			return fmt.Errorf("failed to cleanup old completed captures: %w", err)
		}
	}

	return nil
}
