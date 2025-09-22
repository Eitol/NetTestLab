# Dockerfile para contenedor de test con dependencias pre-instaladas
FROM alpine:latest

# Instalar todas las dependencias de test de una vez
RUN apk add --no-cache \
    curl \
    jq \
    bash \
    bc \
    netcat-openbsd \
    && rm -rf /var/cache/apk/*

# Copy and run initialization script
COPY docker/scripts/test-client-init.sh /usr/local/bin/test-client-init.sh
RUN chmod +x /usr/local/bin/test-client-init.sh && /usr/local/bin/test-client-init.sh

# Crear directorio de trabajo
WORKDIR /tests

# Esta imagen ya tiene todo listo para ejecutar tests
CMD ["/bin/bash"]