package capture

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/Eitol/NetTestLab/api/nettestlab/v1"
	"github.com/Eitol/NetTestLab/internal/device"
	"github.com/Eitol/NetTestLab/internal/target"
)

// CaptureManager administra el ciclo completo de capturas de tráfico
type CaptureManager struct {
	tcpdumpManager *TcpdumpManager
	deviceManager  *device.Manager
	targetManager  *target.Manager
	filterBuilder  *FilterBuilder

	// Configuración
	baseDir            string
	maxCaptures        int
	maxCaptureDuration time.Duration
	maxCaptureSize     int64 // en MB

	// Estado interno
	mutex sync.RWMutex
}

// NewCaptureManager crea un nuevo administrador de capturas
func NewCaptureManager(
	baseDir string,
	deviceManager *device.Manager,
	targetManager *target.Manager,
) *CaptureManager {
	// Configuración por defecto
	maxCaptures := 5
	maxCaptureDuration := 24 * time.Hour
	maxCaptureSize := int64(1024) // 1GB por defecto

	tcpdumpManager := NewTcpdumpManager(filepath.Join(baseDir, "captures"), maxCaptures)
	filterBuilder := NewFilterBuilder()

	return &CaptureManager{
		tcpdumpManager:     tcpdumpManager,
		deviceManager:      deviceManager,
		targetManager:      targetManager,
		filterBuilder:      filterBuilder,
		baseDir:            baseDir,
		maxCaptures:        maxCaptures,
		maxCaptureDuration: maxCaptureDuration,
		maxCaptureSize:     maxCaptureSize,
	}
}

// StartCapture inicia una nueva captura basada en la solicitud
func (cm *CaptureManager) StartCapture(ctx context.Context, req *pb.StartCaptureRequest) (*pb.StartCaptureResponse, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Validar entrada
	if req.CaptureName == "" {
		return &pb.StartCaptureResponse{
			Success: false,
			Message: "capture name is required",
		}, nil
	}

	// Construir filtro BPF
	bpfFilter, err := cm.buildBPFFilter(req)
	if err != nil {
		return &pb.StartCaptureResponse{
			Success: false,
			Message: fmt.Sprintf("failed to build BPF filter: %v", err),
		}, nil
	}

	// Determinar duración máxima
	var maxDuration time.Duration
	if req.Duration != nil {
		maxDuration = req.Duration.AsDuration()
	} else {
		maxDuration = cm.maxCaptureDuration
	}

	// Determinar tamaño máximo
	maxSizeMB := req.MaxSizeMb
	if maxSizeMB == 0 {
		maxSizeMB = cm.maxCaptureSize
	}

	// Configurar opciones de captura
	opts := StartCaptureOptions{
		Name:           req.CaptureName,
		Interface:      "", // Dejar vacío para autodetección
		BPFFilter:      bpfFilter,
		MaxDuration:    maxDuration,
		MaxSizeMB:      maxSizeMB,
		CapturePayload: req.CapturePayload,
	}

	// Iniciar captura con tcpdump
	process, err := cm.tcpdumpManager.StartCapture(opts)
	if err != nil {
		return &pb.StartCaptureResponse{
			Success: false,
			Message: fmt.Sprintf("failed to start capture: %v", err),
		}, nil
	}

	return &pb.StartCaptureResponse{
		CaptureId: process.ID,
		Success:   true,
		Message:   fmt.Sprintf("Capture '%s' started successfully", req.CaptureName),
		BpfFilter: bpfFilter,
	}, nil
}

// buildBPFFilter construye el filtro BPF basado en la solicitud
func (cm *CaptureManager) buildBPFFilter(req *pb.StartCaptureRequest) (string, error) {
	var devices []*pb.Device
	var targets []*pb.UrlTarget

	// Obtener dispositivos si se especificaron IDs
	if len(req.DeviceIds) > 0 {
		// Por ahora, necesitamos listar todos los dispositivos y filtrar
		// TODO: Implementar GetDevice en device.Manager
		allDevices, _, _, err := cm.deviceManager.ListDevices(pb.DeviceFilter_DEVICE_FILTER_ALL, 0, "")
		if err != nil {
			return "", fmt.Errorf("failed to list devices: %w", err)
		}

		// Filtrar por IDs solicitados
		deviceMap := make(map[string]*pb.Device)
		for _, device := range allDevices {
			deviceMap[device.Id] = device
		}

		for _, deviceID := range req.DeviceIds {
			if device, exists := deviceMap[deviceID]; exists {
				devices = append(devices, device)
			} else {
				return "", fmt.Errorf("device %s not found", deviceID)
			}
		}
	}

	// Obtener targets si se especificaron IDs
	if len(req.UrlTargetIds) > 0 {
		for _, targetID := range req.UrlTargetIds {
			target, err := cm.targetManager.GetTargetByID(targetID)
			if err != nil {
				return "", fmt.Errorf("URL target %s not found: %w", targetID, err)
			}
			targets = append(targets, target)
		}
	}

	// Si hay filtro BPF personalizado, usarlo
	if req.CustomBpf != "" {
		return req.CustomBpf, nil
	}

	// Si hay dispositivos o targets, usar FilterBuilder
	if len(devices) > 0 || len(targets) > 0 {
		filter, err := cm.filterBuilder.BuildFilter(devices, targets, false)
		if err != nil {
			return "", fmt.Errorf("failed to build filter: %w", err)
		}
		return filter, nil
	}

	// Si no hay dispositivos ni targets, construir filtro básico de protocolos/puertos
	return cm.buildBasicFilter(req.Protocols, req.Ports), nil
}

// buildBasicFilter construye un filtro BPF básico para protocolos y puertos
func (cm *CaptureManager) buildBasicFilter(protocols []string, ports []int32) string {
	var filterParts []string

	// Agregar filtro de protocolos si se especifica
	if len(protocols) > 0 {
		var protocolFilters []string
		for _, protocol := range protocols {
			switch strings.ToLower(protocol) {
			case "tcp":
				protocolFilters = append(protocolFilters, "tcp")
			case "udp":
				protocolFilters = append(protocolFilters, "udp")
			case "icmp":
				protocolFilters = append(protocolFilters, "icmp")
			}
		}
		if len(protocolFilters) > 0 {
			if len(protocolFilters) == 1 {
				filterParts = append(filterParts, protocolFilters[0])
			} else {
				filterParts = append(filterParts, fmt.Sprintf("(%s)", strings.Join(protocolFilters, " or ")))
			}
		}
	}

	// Agregar filtro de puertos si se especifica
	if len(ports) > 0 {
		var portFilters []string
		for _, port := range ports {
			portFilters = append(portFilters, fmt.Sprintf("port %d", port))
		}
		if len(portFilters) == 1 {
			filterParts = append(filterParts, portFilters[0])
		} else {
			filterParts = append(filterParts, fmt.Sprintf("(%s)", strings.Join(portFilters, " or ")))
		}
	}

	// Si no hay criterios, capturar todo el tráfico
	if len(filterParts) == 0 {
		return "" // Filtro vacío captura todo
	}

	return strings.Join(filterParts, " and ")
}

// StopCapture detiene una captura específica
func (cm *CaptureManager) StopCapture(ctx context.Context, req *pb.StopCaptureRequest) (*pb.StopCaptureResponse, error) {
	if req.CaptureId == "" {
		return &pb.StopCaptureResponse{
			Success: false,
			Message: "capture ID is required",
		}, nil
	}

	err := cm.tcpdumpManager.StopCapture(req.CaptureId)
	if err != nil {
		return &pb.StopCaptureResponse{
			Success: false,
			Message: fmt.Sprintf("failed to stop capture: %v", err),
		}, nil
	}

	// Obtener estadísticas finales
	process, err := cm.tcpdumpManager.GetCapture(req.CaptureId)
	if err != nil {
		return &pb.StopCaptureResponse{
			Success: true,
			Message: "Capture stopped successfully",
		}, nil
	}

	stats := cm.convertStats(process.Stats, process.StartTime, process.EndTime)

	return &pb.StopCaptureResponse{
		Success: true,
		Message: fmt.Sprintf("Capture '%s' stopped successfully", process.Name),
		Stats:   stats,
	}, nil
}

// GetCaptureStatus obtiene el estado de una captura específica
func (cm *CaptureManager) GetCaptureStatus(ctx context.Context, req *pb.GetCaptureStatusRequest) (*pb.GetCaptureStatusResponse, error) {
	if req.CaptureId == "" {
		return nil, fmt.Errorf("capture ID is required")
	}

	// Obtener información de la captura
	process, err := cm.tcpdumpManager.GetCapture(req.CaptureId)
	if err != nil {
		return nil, fmt.Errorf("capture not found: %w", err)
	}

	// Convertir a formato protobuf
	captureInfo := cm.convertToCaptureInfo(process)

	// Obtener estadísticas actuales
	stats := cm.convertStats(process.Stats, process.StartTime, process.EndTime)

	return &pb.GetCaptureStatusResponse{
		Capture: captureInfo,
		Stats:   stats,
	}, nil
}

// ListCaptures lista todas las capturas
func (cm *CaptureManager) ListCaptures(ctx context.Context, req *pb.ListCapturesRequest) (*pb.ListCapturesResponse, error) {
	processes := cm.tcpdumpManager.ListCaptures()

	var captureInfos []*pb.CaptureInfo
	for _, process := range processes {
		// Filtrar por estado si se especifica
		if req.StatusFilter != pb.CaptureStatus_CAPTURE_STATUS_UNSPECIFIED {
			if cm.convertStatus(process.Status) != req.StatusFilter {
				continue
			}
		}

		info := cm.convertToCaptureInfo(process)
		captureInfos = append(captureInfos, info)
	}

	// TODO: Implementar paginación si es necesario
	return &pb.ListCapturesResponse{
		Captures:   captureInfos,
		TotalCount: int32(len(captureInfos)),
	}, nil
}

// GetCaptureData obtiene los datos de una captura
func (cm *CaptureManager) GetCaptureData(ctx context.Context, req *pb.GetCaptureDataRequest) (*pb.GetCaptureDataResponse, error) {
	if req.CaptureId == "" {
		return nil, fmt.Errorf("capture ID is required")
	}

	process, err := cm.tcpdumpManager.GetCapture(req.CaptureId)
	if err != nil {
		return nil, fmt.Errorf("capture not found: %w", err)
	}

	// Verificar que la captura esté completa
	if process.Status == StatusActive || process.Status == StatusStarting {
		return nil, fmt.Errorf("capture is still active, cannot retrieve data")
	}

	// Leer archivo de captura
	data, err := cm.readCaptureFile(process.OutputFile, req.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to read capture data: %w", err)
	}

	// Determinar content type
	contentType := cm.getContentType(req.Format)

	return &pb.GetCaptureDataResponse{
		Data:        data,
		ContentType: contentType,
		PacketCount: int32(process.Stats.PacketsCaptured),
		SizeBytes:   int64(len(data)),
	}, nil
}

// DeleteCapture elimina una captura
func (cm *CaptureManager) DeleteCapture(ctx context.Context, req *pb.DeleteCaptureRequest) (*pb.DeleteCaptureResponse, error) {
	if req.CaptureId == "" {
		return &pb.DeleteCaptureResponse{
			Success: false,
			Message: "capture ID is required",
		}, nil
	}

	err := cm.tcpdumpManager.DeleteCapture(req.CaptureId, req.DeleteData)
	if err != nil {
		return &pb.DeleteCaptureResponse{
			Success: false,
			Message: fmt.Sprintf("failed to delete capture: %v", err),
		}, nil
	}

	message := "Capture deleted successfully"
	if req.DeleteData {
		message = "Capture and data files deleted successfully"
	}

	return &pb.DeleteCaptureResponse{
		Success: true,
		Message: message,
	}, nil
}

// Métodos de conversión entre tipos internos y protobuf

func (cm *CaptureManager) convertToCaptureInfo(process *TcpdumpProcess) *pb.CaptureInfo {
	info := &pb.CaptureInfo{
		Id:        process.ID,
		Name:      process.Name,
		Status:    cm.convertStatus(process.Status),
		StartedAt: timestamppb.New(process.StartTime),
		Stats:     cm.convertStats(process.Stats, process.StartTime, process.EndTime),
		FilePaths: []string{process.OutputFile},
	}

	if process.EndTime != nil {
		info.EndedAt = timestamppb.New(*process.EndTime)
	}

	// Agregar configuración original (simulada por ahora)
	info.Config = &pb.StartCaptureRequest{
		CaptureName: process.Name,
		// TODO: Almacenar configuración original en el proceso
	}

	return info
}

func (cm *CaptureManager) convertStatus(status ProcessStatus) pb.CaptureStatus {
	switch status {
	case StatusStarting:
		return pb.CaptureStatus_CAPTURE_STATUS_STARTING
	case StatusActive:
		return pb.CaptureStatus_CAPTURE_STATUS_ACTIVE
	case StatusStopping:
		return pb.CaptureStatus_CAPTURE_STATUS_STOPPING
	case StatusCompleted:
		return pb.CaptureStatus_CAPTURE_STATUS_COMPLETED
	case StatusFailed:
		return pb.CaptureStatus_CAPTURE_STATUS_FAILED
	case StatusCancelled:
		return pb.CaptureStatus_CAPTURE_STATUS_CANCELLED
	default:
		return pb.CaptureStatus_CAPTURE_STATUS_UNSPECIFIED
	}
}

func (cm *CaptureManager) convertStats(stats *CaptureStatistics, startTime time.Time, endTime *time.Time) *pb.CaptureStats {
	if stats == nil {
		return &pb.CaptureStats{}
	}

	var duration time.Duration
	if endTime != nil {
		duration = endTime.Sub(startTime)
	} else {
		duration = time.Since(startTime)
	}

	return &pb.CaptureStats{
		PacketsCaptured: int32(stats.PacketsCaptured),
		PacketsDropped:  int32(stats.PacketsDropped),
		BytesCaptured:   stats.BytesCaptured,
		Duration:        durationpb.New(duration),
		// TODO: Agregar estadísticas por protocolo y dispositivo
		ProtocolCounts: make(map[string]int32),
		DeviceCounts:   make(map[string]int32),
	}
}

// readCaptureFile lee y convierte el archivo de captura al formato solicitado
func (cm *CaptureManager) readCaptureFile(filePath, format string) ([]byte, error) {
	switch strings.ToLower(format) {
	case "pcap", "":
		// Leer archivo pcap directamente
		return cm.readBinaryFile(filePath)
	case "json":
		// TODO: Convertir pcap a JSON usando tshark
		return nil, fmt.Errorf("JSON format not implemented yet")
	case "csv":
		// TODO: Convertir pcap a CSV usando tshark
		return nil, fmt.Errorf("CSV format not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func (cm *CaptureManager) readBinaryFile(filePath string) ([]byte, error) {
	// Por ahora, leer archivo completo
	// TODO: Implementar lectura parcial basada en filtros de tiempo/paquetes
	return os.ReadFile(filePath)
}

func (cm *CaptureManager) getContentType(format string) string {
	switch strings.ToLower(format) {
	case "pcap", "":
		return "application/vnd.tcpdump.pcap"
	case "json":
		return "application/json"
	case "csv":
		return "text/csv"
	default:
		return "application/octet-stream"
	}
}

// PerformCleanup limpia capturas antiguas y archivos temporales
func (cm *CaptureManager) PerformCleanup(maxAge time.Duration) error {
	return cm.tcpdumpManager.Cleanup(maxAge)
}

// GetStatistics obtiene estadísticas generales del sistema de capturas
func (cm *CaptureManager) GetStatistics() map[string]interface{} {
	processes := cm.tcpdumpManager.ListCaptures()

	stats := map[string]interface{}{
		"total_captures":     len(processes),
		"active_captures":    0,
		"completed_captures": 0,
		"failed_captures":    0,
	}

	for _, process := range processes {
		switch process.Status {
		case StatusActive, StatusStarting:
			stats["active_captures"] = stats["active_captures"].(int) + 1
		case StatusCompleted:
			stats["completed_captures"] = stats["completed_captures"].(int) + 1
		case StatusFailed:
			stats["failed_captures"] = stats["failed_captures"].(int) + 1
		}
	}

	return stats
}
