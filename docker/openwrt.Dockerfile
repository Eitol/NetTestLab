# Dockerfile para OpenWRT con tcpdump funcional 
FROM albrechtloh/openwrt-docker:latest

# La imagen OpenWRT ya incluye tcpdump en la VM, solo necesitamos configurarlo
# correctamente para que NetTestLab lo encuentre

# Crear directorios necesarios
RUN mkdir -p /tmp/nettestlab/captures && \
    mkdir -p /tmp/nettestlab/profiles

# Script de inicio optimizado
COPY docker/scripts/openwrt-start.sh /usr/local/bin/openwrt-start.sh
RUN chmod +x /usr/local/bin/openwrt-start.sh

CMD ["/usr/local/bin/openwrt-start.sh"]