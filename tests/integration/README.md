# NetTestLab Docker Integration Tests

Este directorio contiene los tests automatizados para NetTestLab que se ejecutan en un entorno Docker con OpenWRT.

## Estructura

```
tests/integration/
├── scripts/
│   ├── run-integration-tests.sh     # Tests principales de funcionalidad
│   └── run-performance-tests.sh     # Tests de rendimiento y carga
├── config/
│   └── nettestlab.yaml              # Configuración para el entorno de test
├── results/                         # Resultados de los tests (generado)
└── README.md                        # Esta documentación
```

## Configuración del Entorno

### Requisitos

- Docker y Docker Compose instalados
- Al menos 2GB de RAM disponible para los contenedores
- Acceso a Internet para descargar las imágenes de Docker

### Imagen Base

Los tests utilizan la imagen oficial de OpenWRT:
- **Imagen**: `albrechtloh/openwrt-docker:latest`
- **Propósito**: Simular un router OpenWRT real
- **Características**: Incluye kernel Linux, iptables, tcpdump y herramientas de red

## Componentes del Test

### 1. Contenedor OpenWRT (`openwrt`)

- **Imagen**: `albrechtloh/openwrt-docker:latest`
- **Propósito**: Ejecuta NetTestLab en un entorno OpenWRT simulado
- **Puertos**:
  - `8080`: Servidor NetTestLab gRPC
  - `22`: SSH para debugging (opcional)
- **Volúmenes**:
  - `./bin/nettestlab`: Binario compilado de NetTestLab
  - `./tests/integration/config`: Archivos de configuración
  - `./tests/integration/data`: Datos temporales y capturas

### 2. Cliente de Test (`test-client`)

- **Imagen**: `alpine:latest`
- **Propósito**: Ejecuta los scripts de test contra el servidor NetTestLab
- **Herramientas**: curl, jq, bash para interactuar con la API REST/gRPC
- **Scripts**:
  - `/tests/run-integration-tests.sh`: Tests de funcionalidad completa
  - `/tests/run-performance-tests.sh`: Tests de rendimiento

### 3. Generador de Tráfico (`traffic-generator`)

- **Imagen**: `alpine:latest`
- **Propósito**: Genera tráfico HTTP/HTTPS para probar la captura
- **Tráfico**: Requests periódicos a servicios externos y internos
- **Herramientas**: curl, wget, netcat para diferentes tipos de tráfico

## Tests Implementados

### Tests de Funcionalidad (`run-integration-tests.sh`)

1. **Test de Salud**
   - Verifica que el servidor NetTestLab esté corriendo
   - Endpoint: `/nettestlab.v1.MonitoringService/GetHealth`

2. **Test de Gestión de Dispositivos**
   - Lista dispositivos (inicialmente vacío)
   - Registra un dispositivo de prueba
   - Valida persistencia en base de datos

3. **Test de Gestión de URL Targets**
   - Crea targets con patrones regex
   - Lista targets creados
   - Valida configuración de filtros

4. **Test de Captura de Tráfico**
   - Inicia captura con dispositivos y targets específicos
   - Monitorea el estado de la captura
   - Detiene captura y valida resultados

5. **Test de Integración del Sistema**
   - Verifica disponibilidad de tcpdump
   - Valida interfaces de red
   - Confirma capacidades del kernel

### Tests de Rendimiento (`run-performance-tests.sh`)

1. **Generación de Datos Masivos**
   - Registra 50 dispositivos de prueba
   - Crea 20 URL targets diferentes
   - Valida escalabilidad

2. **Performance de Listado**
   - Mide tiempo de respuesta para listar dispositivos
   - Valida umbral de rendimiento (< 2 segundos)

3. **Capturas Concurrentes**
   - Ejecuta 5 capturas simultáneas
   - Valida que el sistema maneja concurrencia
   - Monitorea recursos

4. **Uso de Memoria**
   - Simula monitoreo de memoria durante operaciones
   - (En producción monitorizaría uso real de contenedor)

## Ejecución

### Tests Completos

```bash
# Ejecutar todos los tests desde cero
make test-docker

# Limpiar entorno después de tests
make test-docker-clean
```

### Tests en Entorno Existente

```bash
# Ejecutar solo tests de funcionalidad
make test-integration

# Ver logs en tiempo real
make test-logs

# Ver resultados
make test-results
```

### Comandos Docker Directos

```bash
# Levantar entorno
docker-compose -f docker-compose.test.yml up -d

# Ejecutar tests manualmente
docker-compose -f docker-compose.test.yml exec test-client /tests/run-integration-tests.sh

# Ver logs
docker-compose -f docker-compose.test.yml logs -f openwrt

# Limpiar
docker-compose -f docker-compose.test.yml down -v
```

## Resultados de Tests

### Archivos Generados

- `test_report.json`: Reporte completo de tests de funcionalidad
- `performance_report.json`: Métricas de rendimiento
- `test_device_id.txt`: ID del dispositivo de prueba creado
- `test_target_id.txt`: ID del target de prueba creado
- `test_capture_id.txt`: ID de la captura de prueba

### Formato del Reporte

```json
{
  "test_run": {
    "timestamp": "2024-01-15T10:30:00.000Z",
    "environment": "docker-openwrt",
    "total_tests": 15,
    "passed": 15,
    "failed": 0,
    "success_rate": 100.0
  },
  "test_categories": {
    "health": "completed",
    "device_management": "completed", 
    "url_targets": "completed",
    "traffic_capture": "completed",
    "system_integration": "completed"
  }
}
```

## Debugging

### Acceso al Contenedor OpenWRT

```bash
# Ejecutar shell en el contenedor
docker-compose -f docker-compose.test.yml exec openwrt sh

# Ver logs del servidor NetTestLab
docker-compose -f docker-compose.test.yml exec openwrt tail -f /tmp/nettestlab/nettestlab.log

# Verificar procesos
docker-compose -f docker-compose.test.yml exec openwrt ps aux | grep nettestlab
```

### Debugging de Tests

```bash
# Ejecutar tests paso a paso
docker-compose -f docker-compose.test.yml exec test-client sh
cd /tests
bash -x ./run-integration-tests.sh
```

### Verificación Manual de APIs

```bash
# Test de salud
curl -X POST http://localhost:8080/nettestlab.v1.MonitoringService/GetHealth

# Listar dispositivos
curl -X POST http://localhost:8080/nettestlab.v1.TrafficCaptureService/ListDevices \
  -H "Content-Type: application/json" -d '{}'
```

## Configuración Avanzada

### Variables de Entorno

- `OPENWRT_HOST`: Host del servidor NetTestLab (default: localhost)
- `OPENWRT_PORT`: Puerto del servidor (default: 8080)
- `TEST_RESULTS_DIR`: Directorio para resultados (default: /results)

### Personalización de Tests

Para añadir nuevos tests:

1. Editar `run-integration-tests.sh`
2. Añadir nuevas funciones de test
3. Llamar las funciones desde `main()`
4. Actualizar contadores y reportes

### Red Docker

Los contenedores se comunican a través de la red `nettestlab-network`:
- Subnet: `192.168.100.0/24`
- Driver: bridge
- DNS automático entre contenedores

## Solución de Problemas

### Problemas Comunes

1. **"Server is not available"**
   - Verificar que el binario de NetTestLab está en `./bin/nettestlab`
   - Comprobar que el contenedor OpenWRT tiene suficiente tiempo para arrancar
   - Revisar logs: `make test-logs`

2. **"tcpdump not found"**
   - Verificar que la imagen OpenWRT incluye tcpdump
   - Comprobar instalación en el comando de inicio

3. **Tests fallan aleatoriamente**
   - Aumentar timeouts en los scripts
   - Verificar recursos disponibles del sistema
   - Comprobar conectividad de red entre contenedores

### Logs Importantes

- **NetTestLab Server**: `/tmp/nettestlab/nettestlab.log`
- **Docker Compose**: `docker-compose -f docker-compose.test.yml logs`
- **Test Results**: `./tests/integration/results/`

## Contribución

Para contribuir nuevos tests:

1. Fork del repositorio
2. Crear branch para el nuevo test
3. Implementar test siguiendo el patrón existente
4. Actualizar documentación
5. Crear Pull Request

Los tests deben ser:
- **Idempotentes**: Pueden ejecutarse múltiples veces
- **Aislados**: No dependen del estado de otros tests
- **Deterministas**: Producen resultados consistentes
- **Documentados**: Con comentarios claros sobre qué prueban

---

## Tests Legacy de OpenWRT Físico

Esta sección contiene documentación para tests en routers OpenWRT físicos (legacy):

### Configuración Legacy

Set environment variables before running tests:

```bash
# Required
export OPENWRT_SDK_PATH="/path/to/openwrt-sdk"
export NETTESTLAB_ROUTER_IP="192.168.1.1"

# Optional
export NETTESTLAB_ROUTER_USER="root"               # default: root
export NETTESTLAB_ROUTER_PASSWORD="your-password"  # if using password auth
export SSH_KEY_PATH="/path/to/ssh/key"             # if using key auth
```

### Legacy Test Scenarios

1. Package Build Test - Compiles package using OpenWrt build system
2. Package Deployment Test - Installs package via SCP and opkg
3. Service Status Test - Verifies service health
4. gRPC Connectivity Test - Tests API responsiveness
5. Profile Management Test - CRUD operations on profiles
6. Network Control Test - Applies and verifies network conditions
7. System Monitoring Test - Retrieves system metrics

Para más detalles sobre tests legacy, consultar la documentación histórica del proyecto.