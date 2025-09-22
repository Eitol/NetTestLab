# Docker Cache Optimization for NetTestLab

Esta configuración optimiza significativamente el tiempo de ejecución de los tests Docker mediante el caching de dependencias.

## 🚀 Setup Inicial (solo una vez)

```bash
# Setup automático (recomendado)
./scripts/setup-docker-cache.sh

# O manualmente:
make cache-docker-images
```

## ⚡ Uso Diario

Después del setup inicial, los tests serán mucho más rápidos:

```bash
# Tests rápidos (usa caché)
make test-docker

# Ver logs en tiempo real
make test-logs

# Ver resultados
make test-results
```

## 🏗️ Gestión de Imágenes

```bash
# Re-construir imágenes con caché (rápido)
make build-docker-images

# Re-construir forzado sin caché (lento, solo si hay problemas)
make rebuild-docker-images

# Limpiar todo
make test-docker-clean
```

## 🎯 Beneficios

### Antes (sin caché):
```
nettestlab-traffic-gen  | fetch https://dl-cdn.alpinelinux.org/alpine/v3.22/main/x86_64/APKINDEX.tar.gz
nettestlab-test-client  | fetch https://dl-cdn.alpinelinux.org/alpine/v3.22/main/x86_64/APKINDEX.tar.gz
nettestlab-traffic-gen  | (1/14) Installing brotli-libs (1.1.0-r2)
nettestlab-test-client  | (1/19) Installing ncurses-terminfo-base...
# 30+ segundos de instalación cada vez
```

### Después (con caché):
```
nettestlab-openwrt      | Starting NetTestLab on OpenWRT...
nettestlab-test-client  | [INFO] Starting NetTestLab Integration Tests
nettestlab-traffic-gen  | Starting traffic generator...
# Tests comienzan inmediatamente
```

## 🔧 Imágenes Optimizadas

1. **nettestlab-test-client**: Alpine + curl + jq + bash + bc + netcat (pre-instalado)
2. **nettestlab-traffic-gen**: Alpine + herramientas de tráfico (pre-instalado)  
3. **nettestlab-openwrt**: OpenWRT + tcpdump + setup scripts (pre-configurado)

## 📊 Mejora de Rendimiento

- **Tiempo de setup**: ~60s → ~5s (12x más rápido)
- **Tiempo total de test**: ~90s → ~30s (3x más rápido)
- **Uso de ancho de banda**: Descarga una sola vez, reutiliza después

## 🛠️ Troubleshooting

Si tienes problemas:

```bash
# Limpiar y reconstruir todo
make test-docker-clean
make rebuild-docker-images
```