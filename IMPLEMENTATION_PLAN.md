# Plan de Implementación - Sistema de Captura de Tráfico

## 🎯 Estrategia de Implementación

El desarrollo se realizará en **5 fases incrementales**, donde cada fase produce funcionalidad utilizable independientemente. Esto permite validar el sistema paso a paso y obtener feedback temprano.

---

## 📋 Fase 1: Fundamentos y Detección de Dispositivos

**⏱️ Duración Estimada:** 2-3 semanas  
**🎯 Objetivo:** Establecer la base del sistema y detección automática de dispositivos

### 1.1 Configuración Inicial del Proyecto

**Tareas:**
- [ ] Crear `proto/nettestlab/v1/traffic_capture.proto` con mensajes básicos
- [ ] Configurar generación de código gRPC (buf generate)
- [ ] Crear estructura de directorios en `internal/`
- [ ] Configurar SQLite para almacenamiento de dispositivos

**Entregables:**
- Protobuf definitions para Device y APIs básicas
- Estructura de proyecto completa
- Base de datos SQLite configurada

### 1.2 Sistema de Detección de Dispositivos

**Tareas:**
- [ ] Implementar `internal/device/discovery.go`
  - Lectura de tabla ARP (`/proc/net/arp`)
  - Parsing de DHCP leases (`/var/lib/dhcp/dhcpd.leases`)
  - Detección periódica (cada 30s)
- [ ] Implementar `internal/device/vendor_lookup.go`
  - Base de datos MAC → Vendor
  - API de lookup local
- [ ] Implementar `internal/device/manager.go`
  - CRUD de dispositivos
  - Merge de dispositivos detectados vs registrados

**Entregables:**
- Detección automática funcionando
- API de gestión de dispositivos
- Sistema de identificación de vendors

### 1.3 APIs Básicas de Dispositivos

**Tareas:**
- [ ] Implementar `internal/server/traffic_capture_service.go` (parcial)
  - `ListDevices` - con filtros básicos
  - `RegisterDevice` - registro manual
  - `UpdateDevice` - actualización de info
  - `DeleteDevice` - eliminación de registros
- [ ] Tests unitarios para device management
- [ ] Documentación de APIs

**Entregables:**
- APIs de dispositivos funcionales
- Sistema de detección automática activo
- Tests básicos

**✅ Criterios de Aceptación Fase 1:**
- Router detecta automáticamente dispositivos conectados
- API permite listar dispositivos (conectados + registrados)
- Se puede registrar dispositivos manualmente
- Vendor lookup funciona correctamente

---

## 📋 Fase 2: Gestión de URL Targets y Filtros

**⏱️ Duración Estimada:** 1-2 semanas  
**🎯 Objetivo:** Sistema de objetivos con regex para filtrar tráfico

### 2.1 Sistema de URL Targets

**Tareas:**
- [ ] Implementar `internal/target/manager.go`
  - CRUD completo de targets
  - Validación de regex
  - Almacenamiento en SQLite
- [ ] Implementar `internal/target/regex_validator.go`
  - Validación de expresiones regulares
  - Testing de regex contra dominios conocidos
  - Optimización de patrones

**Entregables:**
- Gestión completa de URL targets
- Validador de regex robusto

### 2.2 APIs de URL Targets

**Tareas:**
- [ ] Extender `traffic_capture_service.go` con:
  - `CreateUrlTarget`
  - `ListUrlTargets`
  - `UpdateUrlTarget`
  - `DeleteUrlTarget`
- [ ] Implementar filtros y búsqueda de targets
- [ ] Tests para validación de regex

**Entregables:**
- APIs de targets completas
- Sistema de validación funcionando

### 2.3 Constructor de Filtros BPF

**Tareas:**
- [ ] Implementar `internal/capture/filters.go`
  - Construcción de filtros BPF desde devices + targets
  - Optimización de filtros complejos
  - Testing con tcpdump

**Entregables:**
- Sistema de filtros BPF funcional
- Integración con dispositivos y targets

**✅ Criterios de Aceptación Fase 2:**
- Se pueden crear targets con regex complejas
- Filtros BPF se generan correctamente
- Sistema valida regex antes de guardar
- APIs de targets completamente funcionales

---

## 📋 Fase 3: Sistema de Captura Básico

**⏱️ Duración Estimada:** 2-3 semanas  
**🎯 Objetivo:** Captura de tráfico funcional sin descifrado SSL

### 3.1 Motor de Captura con tcpdump

**Tareas:**
- [ ] Implementar `internal/capture/tcpdump.go`
  - Wrapper para ejecutar tcpdump
  - Gestión de procesos background
  - Manejo de errores y timeouts
- [ ] Implementar `internal/capture/manager.go`
  - Gestión de capturas activas
  - Estado de capturas (running, stopped, failed)
  - Limitación de capturas concurrentes

**Entregables:**
- Motor de captura funcional
- Gestión de procesos tcpdump

### 3.2 Almacenamiento de Capturas

**Tareas:**
- [ ] Implementar `internal/capture/storage.go`
  - Gestión de archivos PCAP
  - Rotación automática por tamaño/tiempo
  - Cleanup de archivos antiguos
  - Metadata de capturas en SQLite
- [ ] Sistema de nombres únicos para archivos
- [ ] Compresión opcional de capturas antiguas

**Entregables:**
- Sistema de almacenamiento robusto
- Gestión automática de espacio

### 3.3 APIs de Captura

**Tareas:**
- [ ] Implementar APIs de captura en `traffic_capture_service.go`:
  - `StartCapture` - iniciar captura con filtros
  - `StopCapture` - detener captura específica
  - `ListCaptures` - listar capturas existentes
  - `GetCaptureData` - obtener datos con filtros básicos
  - `DeleteCapture` - borrar capturas
- [ ] Validación de parámetros de captura
- [ ] Tests de integración

**Entregables:**
- APIs de captura completas
- Sistema de captura funcional end-to-end

**✅ Criterios de Aceptación Fase 3:**
- Se puede iniciar/detener capturas por dispositivos específicos
- Filtros por URL targets funcionan correctamente
- Archivos PCAP se generan y almacenan correctamente
- APIs permiten gestión completa de capturas
- Sistema de limpieza automática funciona

---

## 📋 Fase 4: Descifrado SSL y Análisis Avanzado

**⏱️ Duración Estimada:** 2-3 semanas  
**🎯 Objetivo:** Capacidades de descifrado SSL y análisis de contenido

### 4.1 Gestión de Certificados SSL

**Tareas:**
- [ ] Implementar `internal/ssl/certificate_manager.go`
  - Upload y validación de certificados PEM
  - Almacenamiento seguro de claves privadas
  - Gestión de expiración de certificados
- [ ] APIs para certificados SSL:
  - `UploadSSLCertificate`
  - `ListSSLCertificates`
  - `DeleteSSLCertificate`

**Entregables:**
- Gestión segura de certificados SSL
- APIs de certificados funcionales

### 4.2 Sistema de Descifrado

**Tareas:**
- [ ] Implementar `internal/ssl/decryptor.go`
  - Integración con tshark para descifrado
  - Generación de keylog files
  - Processing de PCAP descifrados
- [ ] Implementar `internal/ssl/keylog.go`
  - Generación de SSLKEYLOGFILE format
  - Gestión de claves de sesión

**Entregables:**
- Sistema de descifrado SSL funcional
- Integración con tshark

### 4.3 Análisis de Contenido Descifrado

**Tareas:**
- [ ] Parsing de HTTP/HTTPS descifrado
- [ ] Extracción de URLs, headers, payload
- [ ] Filtros avanzados para contenido descifrado
- [ ] APIs de búsqueda en contenido

**Entregables:**
- Análisis de contenido HTTP/HTTPS
- Búsqueda en datos descifrados

**✅ Criterios de Aceptación Fase 4:**
- Certificados SSL se pueden cargar y gestionar
- Tráfico HTTPS se descifra correctamente
- Se puede buscar en contenido HTTP descifrado
- APIs de SSL completamente funcionales

---

## 📋 Fase 5: Streaming, Optimización y UI

**⏱️ Duración Estimada:** 2-3 semanas  
**🎯 Objetivo:** Streaming en tiempo real, optimizaciones y interfaz web

### 5.1 Streaming en Tiempo Real

**Tareas:**
- [ ] Implementar `StreamCapture` API
  - Streaming de paquetes en tiempo real
  - Filtros aplicados al stream
  - Gestión de backpressure
- [ ] Optimización de performance para streaming
- [ ] Rate limiting y gestión de conexiones concurrentes

**Entregables:**
- Streaming de captura en tiempo real
- APIs de streaming optimizadas

### 5.2 Optimizaciones de Rendimiento

**Tareas:**
- [ ] Optimización de filtros BPF
- [ ] Caching de dispositivos detectados
- [ ] Reducción de overhead de tcpdump
- [ ] Monitoreo de recursos del sistema
- [ ] Alertas de uso excesivo de CPU/memoria

**Entregables:**
- Sistema optimizado para OpenWRT
- Monitoreo de recursos integrado

### 5.3 Interfaz Web Extendida

**Tareas:**
- [ ] Extender UI web existente con:
  - Gestión visual de dispositivos
  - Configuración de URL targets
  - Visualización de capturas en tiempo real
  - Dashboard de análisis de tráfico
- [ ] Gráficos de uso de red por dispositivo
- [ ] Exportación de reportes

**Entregables:**
- Interfaz web completa
- Dashboards de análisis

### 5.4 Integración con Sistema Existente

**Tareas:**
- [ ] Integrar métricas de captura en `MonitoringService`
- [ ] Extender clientes (Go, Python, JavaScript, Java)
- [ ] Documentación completa de APIs
- [ ] Ejemplos de uso y tutoriales

**Entregables:**
- Integración completa con NetTestLab existente
- Documentación y ejemplos

**✅ Criterios de Aceptación Fase 5:**
- Streaming en tiempo real funciona correctamente
- UI web permite gestión completa del sistema
- Performance optimizada para OpenWRT
- Integración completa con sistema existente
- Documentación y ejemplos disponibles

---

## 📊 Cronograma General

| Fase | Duración | Semanas Acumuladas | Funcionalidad Principal |
|------|----------|-------------------|------------------------|
| **Fase 1** | 2-3 semanas | 0-3 | Detección de dispositivos |
| **Fase 2** | 1-2 semanas | 3-5 | URL Targets y filtros |
| **Fase 3** | 2-3 semanas | 5-8 | Captura básica |
| **Fase 4** | 2-3 semanas | 8-11 | Descifrado SSL |
| **Fase 5** | 2-3 semanas | 11-14 | Streaming y UI |

**📅 Duración Total Estimada:** 11-14 semanas (2.5-3.5 meses)

## 🔄 Metodología de Desarrollo

### Principios
- **Desarrollo incremental**: Cada fase produce valor utilizable
- **Testing continuo**: Tests unitarios y de integración en cada fase
- **Feedback temprano**: Validación al final de cada fase
- **Documentación continua**: APIs documentadas durante desarrollo

### Validación Entre Fases
- **Demo funcional** al final de cada fase
- **Tests de rendimiento** en fases 3, 4 y 5
- **Feedback y ajustes** antes de continuar a siguiente fase

### Criterios de Calidad
- **Cobertura de tests > 80%** para componentes críticos
- **Performance**: Capturas sin impact significativo en router
- **Memoria**: Límites configurables y respetados
- **Logs**: Logging comprehensivo para debugging

---

## 🚀 Recomendación de Inicio

**Sugerencia:** Comenzar con **Fase 1** implementando primero la detección de dispositivos, ya que es la base fundamental del sistema y proporciona valor inmediato para entender qué dispositivos están conectados al router.

**¿Estás listo para comenzar con la Fase 1, o prefieres ajustar algún aspecto del plan?**