#!/bin/sh

# Script para generar tráfico de prueba
set -e

echo "Starting traffic generator..."

# Esperar a que el servidor esté listo
sleep 15

while true; do
    # HTTP traffic
    curl -s http://httpbin.org/get > /dev/null 2>&1 || true
    
    # HTTPS traffic  
    curl -s https://httpbin.org/get > /dev/null 2>&1 || true
    
    # Traffic to NetTestLab server
    nc -z openwrt 8080 || true
    
    echo "Generated traffic batch at $(date)"
    sleep 30
done