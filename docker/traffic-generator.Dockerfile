# Dockerfile para generador de tráfico con dependencias pre-instaladas
FROM alpine:latest

# Instalar herramientas de generación de tráfico
RUN apk add --no-cache \
    curl \
    wget \
    netcat-openbsd \
    && rm -rf /var/cache/apk/*

# Script para generar tráfico
COPY docker/scripts/generate-traffic.sh /usr/local/bin/generate-traffic.sh
RUN chmod +x /usr/local/bin/generate-traffic.sh

CMD ["/usr/local/bin/generate-traffic.sh"]