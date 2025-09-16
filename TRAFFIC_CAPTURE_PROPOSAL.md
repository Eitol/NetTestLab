# Sistema de Captura de Tráfico para NetTestLab

## 📋 Resumen Ejecutivo

Este documento describe la implementación completa de un sistema de captura de tráfico para NetTestLab, ejecutándose en routers OpenWRT. El sistema permitirá monitorear, capturar y analizar el tráfico de red de dispositivos conectados con capacidades avanzadas de filtrado y descifrado SSL opcional.

## 🏗️ Arquitectura del Sistema

### Componentes Principales

1. **TrafficCaptureService** - Servicio gRPC principal
2. **Device Management** - Gestión de dispositivos conectados y registrados
3. **URL Target Management** - Gestión de objetivos con regex
4. **SSL Decryption** - Descifrado opcional de tráfico HTTPS
5. **Storage Management** - Gestión de archivos de captura

## 🔌 Definición del Servicio gRPC

```protobuf
service TrafficCaptureService {
  // === GESTIÓN DE DISPOSITIVOS ===
  // Listar todos los dispositivos (conectados + registrados)
  rpc ListDevices(ListDevicesRequest) returns (ListDevicesResponse);
  
  // Registrar un dispositivo manualmente o via app
  rpc RegisterDevice(RegisterDeviceRequest) returns (RegisterDeviceResponse);
  
  // Actualizar información de dispositivo (desde la app del dispositivo)
  rpc UpdateDevice(UpdateDeviceRequest) returns (UpdateDeviceResponse);
  
  // Eliminar dispositivo registrado
  rpc DeleteDevice(DeleteDeviceRequest) returns (DeleteDeviceResponse);
  
  // === GESTIÓN DE URL TARGETS ===
  rpc CreateUrlTarget(CreateUrlTargetRequest) returns (CreateUrlTargetResponse);
  rpc ListUrlTargets(ListUrlTargetsRequest) returns (ListUrlTargetsResponse);
  rpc UpdateUrlTarget(UpdateUrlTargetRequest) returns (UpdateUrlTargetResponse);
  rpc DeleteUrlTarget(DeleteUrlTargetRequest) returns (DeleteUrlTargetResponse);
  
  // === CAPTURA DE TRÁFICO ===
  // Iniciar captura (todos los dispositivos o específicos)
  rpc StartCapture(StartCaptureRequest) returns (StartCaptureResponse);
  
  // Detener captura
  rpc StopCapture(StopCaptureRequest) returns (StopCaptureResponse);
  
  // Listar capturas existentes
  rpc ListCaptures(ListCapturesRequest) returns (ListCapturesResponse);
  
  // Obtener datos de captura con filtros
  rpc GetCaptureData(GetCaptureDataRequest) returns (GetCaptureDataResponse);
  
  // Borrar capturas
  rpc DeleteCapture(DeleteCaptureRequest) returns (DeleteCaptureResponse);
  
  // === SSL/TLS DESCIFRADO ===
  // Gestionar certificados para descifrado SSL
  rpc UploadSSLCertificate(UploadSSLCertificateRequest) returns (UploadSSLCertificateResponse);
  rpc ListSSLCertificates(ListSSLCertificatesRequest) returns (ListSSLCertificatesResponse);
  rpc DeleteSSLCertificate(DeleteSSLCertificateRequest) returns (DeleteSSLCertificateResponse);
  
  // === STREAMING EN TIEMPO REAL ===
  rpc StreamCapture(StreamCaptureRequest) returns (stream CapturePacket);
}
```

## 📊 Estructuras de Datos

### Device (Dispositivos)

```protobuf
message Device {
  string id = 1;                    // UUID generado
  string mac_address = 2;           // MAC única
  string ip_address = 3;            // IP actual (puede cambiar)
  string hostname = 4;              // Nombre de host
  
  // === Información registrada (opcional) ===
  string device_name = 5;           // Nombre personalizado
  string device_model = 6;          // ej: "iPhone 14 Pro"
  string os_version = 7;            // ej: "iOS 17.0"
  string app_version = 8;           // Versión de tu app cliente
  
  // === Estado de conexión y registro ===
  DeviceConnectionStatus connection_status = 9;  // CONNECTED o DISCONNECTED
  bool registered = 10;             // true = registrado manualmente/via app
  
  // === Timestamps ===
  google.protobuf.Timestamp first_seen = 11;
  google.protobuf.Timestamp last_seen = 12;
  google.protobuf.Timestamp registered_at = 13;  // Cuando se registró
  
  // === Datos de red ===
  string vendor = 14;               // Fabricante (via MAC lookup)
  repeated string previous_ips = 15; // IPs anteriores
}

enum DeviceConnectionStatus {
  DEVICE_CONNECTION_STATUS_UNSPECIFIED = 0;
  DEVICE_CONNECTION_STATUS_CONNECTED = 1;
  DEVICE_CONNECTION_STATUS_DISCONNECTED = 2;
}
```

### UrlTarget (Objetivos de Captura)

```protobuf
message UrlTarget {
  string id = 1;
  string name = 2;                  // ej: "Redes Sociales"
  string description = 3;
  string host_regex = 4;            // ej: ".*\\.(facebook|instagram|tiktok|twitter)\\.com"
  repeated int32 ports = 5;         // [80, 443, 8080]
  string protocol_filter = 6;       // "HTTP", "HTTPS", "ALL"
  bool enabled = 7;
  google.protobuf.Timestamp created_at = 8;
  google.protobuf.Timestamp updated_at = 9;
}
```

### TrafficCapture (Capturas de Tráfico)

```protobuf
message TrafficCapture {
  string id = 1;
  string name = 2;                  // Nombre descriptivo
  
  // === Configuración de captura ===
  CaptureScope scope = 3;           // ALL_DEVICES, SPECIFIC_DEVICES
  repeated string device_ids = 4;   // Si scope = SPECIFIC_DEVICES
  repeated string url_target_ids = 5; // Targets a capturar
  
  // === Filtros opcionales ===
  bool only_registered_devices = 6;    // Solo dispositivos registrados
  bool only_connected_devices = 7;     // Solo dispositivos conectados
  
  // === Estado ===
  CaptureStatus status = 8;
  google.protobuf.Timestamp started_at = 9;
  google.protobuf.Timestamp stopped_at = 10;
  
  // === Datos de resultado ===
  string pcap_file_path = 11;
  uint64 packet_count = 12;
  uint64 total_bytes = 13;
  uint64 ssl_decrypted_count = 14;  // Paquetes SSL descifrados
  
  // === Configuración SSL ===
  repeated string ssl_certificate_ids = 15; // Certificados usados
}

enum CaptureScope {
  CAPTURE_SCOPE_UNSPECIFIED = 0;
  CAPTURE_SCOPE_ALL_DEVICES = 1;
  CAPTURE_SCOPE_SPECIFIC_DEVICES = 2;
}

enum CaptureStatus {
  CAPTURE_STATUS_UNSPECIFIED = 0;
  CAPTURE_STATUS_RUNNING = 1;
  CAPTURE_STATUS_STOPPED = 2;
  CAPTURE_STATUS_FAILED = 3;
}
```

### SSLCertificate (Certificados SSL)

```protobuf
message SSLCertificate {
  string id = 1;
  string name = 2;                  // Nombre descriptivo
  string domain = 3;                // ej: "*.google.com"
  bytes certificate_data = 4;       // Certificado PEM
  bytes private_key_data = 5;       // Clave privada PEM
  google.protobuf.Timestamp uploaded_at = 6;
  google.protobuf.Timestamp expires_at = 7;
}
```

## 🔧 Implementación Técnica

### Estructura de Directorios

```
proto/nettestlab/v1/
├── traffic_capture.proto         # Servicio completo de captura

internal/
├── capture/
│   ├── manager.go               # Gestor principal de capturas
│   ├── tcpdump.go              # Wrapper para tcpdump
│   ├── filters.go              # Construcción de filtros BPF
│   └── storage.go              # Gestión de archivos PCAP
├── device/
│   ├── manager.go              # Gestión de dispositivos
│   ├── discovery.go            # Auto-descubrimiento de dispositivos
│   ├── registry.go             # Registro manual de dispositivos
│   └── vendor_lookup.go        # Lookup de fabricante por MAC
├── ssl/
│   ├── certificate_manager.go  # Gestión de certificados SSL
│   ├── decryptor.go            # Descifrado de tráfico SSL
│   └── keylog.go               # Generación de keylog files
├── target/
│   ├── manager.go              # Gestión de URL targets
│   └── regex_validator.go      # Validación de regex
└── server/
    └── traffic_capture_service.go

data/
├── captures/                   # Archivos PCAP
├── ssl_certs/                  # Certificados SSL
└── devices.db                  # SQLite con dispositivos registrados
```

### Detección de Dispositivos

El sistema detectará automáticamente dispositivos conectados mediante:

1. **Tabla ARP** (`/proc/net/arp`) - Dispositivos activos
2. **DHCP Leases** (`/var/lib/dhcp/dhcpd.leases`) - Asignaciones IP
3. **MAC Vendor Lookup** - Identificación de fabricante
4. **Reverse DNS** - Resolución de hostname

### Construcción de Filtros BPF

```go
// Ejemplo de filtro generado para captura específica
func buildBPFFilter(devices []Device, targets []UrlTarget) string {
    // Resultado: 
    // "(ether host aa:bb:cc:dd:ee:ff or ether host 11:22:33:44:55:66) and 
    //  (dst port 80 or dst port 443) and 
    //  (host 172.217.0.0/16 or host 142.250.0.0/15)"
}
```

### Descifrado SSL

Para sitios con certificados disponibles:

1. Cargar certificados y claves privadas
2. Generar keylog file para Wireshark/tshark
3. Procesar PCAP con tshark usando keylog
4. Extraer datos HTTP/HTTPS descifrados

## 📝 Casos de Uso

### 1. Dispositivos por Estado

**Conectado No Registrado:**
```json
{
  "mac_address": "aa:bb:cc:dd:ee:ff",
  "connection_status": "CONNECTED",
  "registered": false,
  "hostname": "unknown-device"
}
```

**Registrado y Conectado:**
```json
{
  "mac_address": "11:22:33:44:55:66",
  "connection_status": "CONNECTED", 
  "registered": true,
  "device_name": "iPhone de Juan"
}
```

**Registrado pero Desconectado:**
```json
{
  "mac_address": "99:88:77:66:55:44",
  "connection_status": "DISCONNECTED",
  "registered": true,
  "device_name": "Laptop María"
}
```

### 2. Tipos de Captura

**Captura Global Filtrada:**
```json
{
  "name": "Monitoreo Redes Sociales",
  "scope": "ALL_DEVICES",
  "only_registered_devices": true,
  "url_target_ids": ["social_media_target"]
}
```

**Captura Específica:**
```json
{
  "name": "Análisis iPhone Juan",
  "scope": "SPECIFIC_DEVICES",
  "device_ids": ["dev_juan_iphone"],
  "url_target_ids": ["streaming_target", "gaming_target"]
}
```

## 🔍 APIs de Consulta

### Listar Dispositivos con Filtros

```
GET /api/devices?filter=all|connected|registered|connected_registered
```

### Buscar en Capturas

```
GET /api/captures/data?
  capture_id=cap_123&
  device_mac=aa:bb:cc:dd:ee:ff&
  protocol=HTTPS&
  domain_contains=facebook&
  time_range=2h&
  decrypted_only=true
```

## 🛠️ Consideraciones OpenWRT

### Dependencias
- `tcpdump` (disponible en OpenWRT)
- `libpcap` (requerido por tcpdump)
- `tshark` (opcional, para descifrado SSL)

### Optimizaciones
- Filtros BPF específicos para reducir carga CPU
- Rotación automática de archivos PCAP
- Compresión de capturas antiguas
- Límites de almacenamiento configurables

### Gestión de Recursos
- Monitoreo de uso de CPU/memoria durante captura
- Alertas por uso excesivo de recursos
- Parada automática si se exceden límites

## 🔒 Seguridad

### Certificados SSL
- Almacenamiento seguro de claves privadas
- Validación de certificados antes de uso
- Logs de acceso a certificados

### Acceso a Datos
- Autenticación requerida para APIs sensibles
- Logs de acceso a capturas
- Cifrado opcional de archivos PCAP en reposo

## 📊 Métricas y Monitoreo

### Métricas del Sistema
- Capturas activas
- Dispositivos monitoreados
- Uso de almacenamiento
- Rendimiento de captura (pps, bps)

### Integración con MonitoringService
- Agregar métricas de captura a endpoints existentes
- Alertas de estado del sistema de captura
- Dashboards de uso de red por dispositivo

---

**Documento Versión:** 1.0  
**Fecha:** 15 de septiembre de 2025  
**Estado:** Propuesta Final