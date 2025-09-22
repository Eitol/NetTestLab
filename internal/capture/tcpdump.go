package capture

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
)

// TcpdumpProcess representa un proceso de tcpdump en ejecución
type TcpdumpProcess struct {
	ID         string     // UUID del proceso
	Name       string     // Nombre descriptivo de la captura
	PID        int        // Process ID
	Interface  string     // Interfaz de red (ej: "br-lan")
	BPFFilter  string     // Filtro BPF aplicado
	OutputFile string     // Archivo de salida .pcap
	StartTime  time.Time  // Momento de inicio
	EndTime    *time.Time // Momento de finalización (nil si está activo)

	// Control del proceso
	cmd    *exec.Cmd
	ctx    context.Context
	cancel context.CancelFunc
	mutex  sync.RWMutex

	// Estado del proceso
	Status   ProcessStatus
	ErrorMsg string

	// Estadísticas en tiempo real
	Stats       *CaptureStatistics
	statsReader io.ReadCloser
}

// ProcessStatus representa el estado de un proceso de captura
type ProcessStatus int

const (
	StatusStarting ProcessStatus = iota
	StatusActive
	StatusStopping
	StatusCompleted
	StatusFailed
	StatusCancelled
)

func (s ProcessStatus) String() string {
	switch s {
	case StatusStarting:
		return "starting"
	case StatusActive:
		return "active"
	case StatusStopping:
		return "stopping"
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	case StatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// CaptureStatistics contiene estadísticas de la captura en tiempo real
type CaptureStatistics struct {
	PacketsCaptured int64     // Paquetes capturados
	PacketsDropped  int64     // Paquetes perdidos
	BytesCaptured   int64     // Bytes capturados (estimado)
	LastUpdate      time.Time // Última actualización de stats
	FileSizeBytes   int64     // Tamaño actual del archivo
}

// TcpdumpManager administra múltiples procesos de tcpdump
type TcpdumpManager struct {
	processes   map[string]*TcpdumpProcess
	mutex       sync.RWMutex
	baseDir     string // Directorio base para archivos .pcap
	maxCaptures int    // Máximo de capturas simultáneas
}

// NewTcpdumpManager crea un nuevo administrador de tcpdump
func NewTcpdumpManager(baseDir string, maxCaptures int) *TcpdumpManager {
	// Crear directorio base si no existe
	os.MkdirAll(baseDir, 0755)

	return &TcpdumpManager{
		processes:   make(map[string]*TcpdumpProcess),
		baseDir:     baseDir,
		maxCaptures: maxCaptures,
	}
}

// StartCaptureOptions contiene las opciones para iniciar una captura
type StartCaptureOptions struct {
	Name           string        // Nombre descriptivo
	Interface      string        // Interfaz de red (default: "br-lan")
	BPFFilter      string        // Filtro BPF
	MaxDuration    time.Duration // Duración máxima (0 = sin límite)
	MaxSizeMB      int64         // Tamaño máximo en MB (0 = sin límite)
	CapturePayload bool          // Si capturar payload completo
}

// StartCapture inicia una nueva captura de tcpdump
func (m *TcpdumpManager) StartCapture(opts StartCaptureOptions) (*TcpdumpProcess, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Verificar límite de capturas simultáneas
	activeCount := 0
	for _, proc := range m.processes {
		if proc.Status == StatusActive || proc.Status == StatusStarting {
			activeCount++
		}
	}

	if activeCount >= m.maxCaptures {
		return nil, fmt.Errorf("maximum number of simultaneous captures reached (%d)", m.maxCaptures)
	}

	// Generar ID único
	captureID := uuid.New().String()

	// Configurar valores por defecto
	if opts.Interface == "" {
		// Detectar interfaz por defecto según el sistema
		if _, err := os.Stat("/sys/class/net/br-lan"); err == nil {
			opts.Interface = "br-lan" // OpenWRT bridge interface
		} else if _, err := os.Stat("/sys/class/net/eth0"); err == nil {
			opts.Interface = "eth0" // Standard ethernet interface
		} else {
			opts.Interface = "any" // Capture on all interfaces
		}
	}

	// Crear archivo de salida
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.pcap", timestamp, strings.ReplaceAll(opts.Name, " ", "_"))
	outputFile := filepath.Join(m.baseDir, filename)

	// Crear contexto con timeout si se especifica duración máxima
	ctx := context.Background()
	if opts.MaxDuration > 0 {
		ctx, _ = context.WithTimeout(ctx, opts.MaxDuration)
	}
	ctx, cancel := context.WithCancel(ctx)

	// Crear proceso
	process := &TcpdumpProcess{
		ID:         captureID,
		Name:       opts.Name,
		Interface:  opts.Interface,
		BPFFilter:  opts.BPFFilter,
		OutputFile: outputFile,
		StartTime:  time.Now(),
		ctx:        ctx,
		cancel:     cancel,
		Status:     StatusStarting,
		Stats: &CaptureStatistics{
			LastUpdate: time.Now(),
		},
	}

	// Verificar que tcpdump esté disponible
	if err := m.checkTcpdumpAvailable(); err != nil {
		cancel()
		return nil, fmt.Errorf("tcpdump not available: %w", err)
	}

	// Construir comando tcpdump
	args := m.buildTcpdumpArgs(opts, outputFile)

	// Log del comando completo para debugging
	fmt.Printf("Starting tcpdump with command: tcpdump %s\n", strings.Join(args, " "))
	fmt.Printf("Output file: %s\n", outputFile)
	fmt.Printf("Interface: %s\n", opts.Interface)

	process.cmd = exec.CommandContext(ctx, "tcpdump", args...)

	// Configurar stderr para capturar estadísticas
	stderr, err := process.cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	process.statsReader = stderr

	// Agregar proceso al mapa
	m.processes[captureID] = process

	// Iniciar proceso en goroutine
	go m.runCapture(process, opts.MaxSizeMB)

	return process, nil
}

// buildTcpdumpArgs construye los argumentos para el comando tcpdump
func (m *TcpdumpManager) buildTcpdumpArgs(opts StartCaptureOptions, outputFile string) []string {
	args := []string{
		"-i", opts.Interface, // Interfaz
		"-w", outputFile, // Archivo de salida
		"-U", // Unbuffered output
		"-v", // Verbose (para estadísticas en stderr)
	}

	// Tamaño de snapshot (payload)
	if opts.CapturePayload {
		args = append(args, "-s", "0") // Capturar paquete completo
	} else {
		args = append(args, "-s", "96") // Solo headers
	}

	// Agregar filtro BPF si se especifica
	if opts.BPFFilter != "" {
		args = append(args, opts.BPFFilter)
	}

	return args
}

// runCapture ejecuta el proceso de tcpdump y monitorea su estado
func (m *TcpdumpManager) runCapture(process *TcpdumpProcess, maxSizeMB int64) {
	defer func() {
		process.mutex.Lock()
		if process.Status == StatusActive || process.Status == StatusStarting {
			process.Status = StatusCompleted
		}
		now := time.Now()
		process.EndTime = &now
		process.mutex.Unlock()

		if process.statsReader != nil {
			process.statsReader.Close()
		}
	}()

	// Iniciar monitoreo de estadísticas
	go m.monitorStats(process)

	// Iniciar monitoreo de tamaño de archivo si hay límite
	if maxSizeMB > 0 {
		go m.monitorFileSize(process, maxSizeMB)
	}

	// Ejecutar tcpdump
	process.mutex.Lock()
	process.Status = StatusActive
	process.mutex.Unlock()

	err := process.cmd.Run()

	process.mutex.Lock()
	defer process.mutex.Unlock()

	if err != nil {
		// Log del error detallado para debugging
		fmt.Printf("tcpdump process failed: %v\n", err)
		if exitError, ok := err.(*exec.ExitError); ok {
			fmt.Printf("tcpdump stderr: %s\n", string(exitError.Stderr))
		}

		// Verificar si fue cancelado por contexto
		if process.ctx.Err() == context.Canceled {
			process.Status = StatusCancelled
		} else if process.ctx.Err() == context.DeadlineExceeded {
			process.Status = StatusCompleted // Terminó por timeout
		} else {
			process.Status = StatusFailed
			process.ErrorMsg = err.Error()
		}
	} else {
		process.Status = StatusCompleted
		fmt.Printf("tcpdump process completed successfully for capture: %s\n", process.Name)
	}

	// Obtener PID si está disponible
	if process.cmd.Process != nil {
		process.PID = process.cmd.Process.Pid
	}
}

// monitorStats monitorea las estadísticas del proceso tcpdump
func (m *TcpdumpManager) monitorStats(process *TcpdumpProcess) {
	scanner := bufio.NewScanner(process.statsReader)

	for scanner.Scan() {
		line := scanner.Text()

		// tcpdump envía estadísticas periódicamente por stderr
		// Ejemplo: "1234 packets captured"
		if strings.Contains(line, "packets captured") {
			if packets := parsePacketCount(line); packets > 0 {
				process.mutex.Lock()
				process.Stats.PacketsCaptured = packets
				process.Stats.LastUpdate = time.Now()
				process.mutex.Unlock()
			}
		}

		if strings.Contains(line, "packets dropped") {
			if dropped := parsePacketCount(line); dropped > 0 {
				process.mutex.Lock()
				process.Stats.PacketsDropped = dropped
				process.Stats.LastUpdate = time.Now()
				process.mutex.Unlock()
			}
		}
	}
}

// monitorFileSize monitorea el tamaño del archivo de captura
func (m *TcpdumpManager) monitorFileSize(process *TcpdumpProcess, maxSizeMB int64) {
	maxBytes := maxSizeMB * 1024 * 1024
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-process.ctx.Done():
			return
		case <-ticker.C:
			if stat, err := os.Stat(process.OutputFile); err == nil {
				process.mutex.Lock()
				process.Stats.FileSizeBytes = stat.Size()
				process.mutex.Unlock()

				if stat.Size() >= maxBytes {
					// Parar captura por límite de tamaño
					process.cancel()
					return
				}
			}
		}
	}
}

// parsePacketCount extrae el número de paquetes de una línea de estadísticas
func parsePacketCount(line string) int64 {
	// Buscar número seguido de "packets"
	fields := strings.Fields(line)
	for i, field := range fields {
		if field == "packets" && i > 0 {
			if count, err := strconv.ParseInt(fields[i-1], 10, 64); err == nil {
				return count
			}
		}
	}
	return 0
}

// StopCapture detiene una captura específica
func (m *TcpdumpManager) StopCapture(captureID string) error {
	m.mutex.RLock()
	process, exists := m.processes[captureID]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("capture %s not found", captureID)
	}

	process.mutex.Lock()
	defer process.mutex.Unlock()

	if process.Status != StatusActive && process.Status != StatusStarting {
		return fmt.Errorf("capture %s is not active (status: %s)", captureID, process.Status)
	}

	process.Status = StatusStopping

	// Intentar terminación limpia primero
	if process.cmd.Process != nil {
		if err := process.cmd.Process.Signal(syscall.SIGTERM); err == nil {
			// Esperar un poco para terminación limpia
			go func() {
				time.Sleep(5 * time.Second)
				// Si aún está ejecutándose, forzar terminación
				if process.cmd.Process != nil {
					process.cmd.Process.Signal(syscall.SIGKILL)
				}
			}()
		}
	}

	// Cancelar contexto
	process.cancel()

	return nil
}

// GetCapture obtiene información de una captura específica
func (m *TcpdumpManager) GetCapture(captureID string) (*TcpdumpProcess, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	process, exists := m.processes[captureID]
	if !exists {
		return nil, fmt.Errorf("capture %s not found", captureID)
	}

	return process, nil
}

// ListCaptures lista todas las capturas (activas e históricas)
func (m *TcpdumpManager) ListCaptures() []*TcpdumpProcess {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var captures []*TcpdumpProcess
	for _, process := range m.processes {
		captures = append(captures, process)
	}

	return captures
}

// DeleteCapture elimina una captura y opcionalmente sus archivos
func (m *TcpdumpManager) DeleteCapture(captureID string, deleteFiles bool) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	process, exists := m.processes[captureID]
	if !exists {
		return fmt.Errorf("capture %s not found", captureID)
	}

	// Si está activa, detenerla primero
	if process.Status == StatusActive || process.Status == StatusStarting {
		process.cancel()
	}

	// Eliminar archivos si se solicita
	if deleteFiles {
		if err := os.Remove(process.OutputFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete capture file: %w", err)
		}
	}

	// Eliminar del mapa
	delete(m.processes, captureID)

	return nil
}

// Cleanup limpia procesos terminados y archivos antiguos
func (m *TcpdumpManager) Cleanup(maxAge time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)

	for captureID, process := range m.processes {
		// Eliminar capturas terminadas antiguas
		if process.Status != StatusActive && process.Status != StatusStarting {
			if process.EndTime != nil && process.EndTime.Before(cutoff) {
				delete(m.processes, captureID)
			}
		}
	}

	return nil
}

// checkTcpdumpAvailable verifica si tcpdump está disponible en el sistema
func (m *TcpdumpManager) checkTcpdumpAvailable() error {
	// Verificar que tcpdump esté en el PATH
	tcpdumpPath, err := exec.LookPath("tcpdump")
	if err != nil {
		return fmt.Errorf("tcpdump command not found in PATH")
	}

	// Verificar que tcpdump funcione ejecutando --version
	cmd := exec.Command("tcpdump", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tcpdump is not working properly: %w", err)
	}

	// Detectar si es tcpdump mock (para testing)
	if strings.Contains(string(output), "mock") {
		fmt.Printf("Using mock tcpdump for testing: %s\n", tcpdumpPath)
	}

	return nil
}
