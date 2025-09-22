#!/bin/sh

# Script para configurar la VM OpenWRT con tcpdump
set -e

echo "Setting up OpenWRT VM with tcpdump..."

# Esperar a que OpenWRT arranque
sleep 15

# Configuración SSH
OPENWRT_HOST="localhost"
OPENWRT_PORT="22"
OPENWRT_USER="root"
OPENWRT_PASS=""  # OpenWRT por defecto no tiene password

# Función para ejecutar comandos SSH sin password
ssh_exec() {
    ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
        -o ConnectTimeout=10 -o PasswordAuthentication=yes \
        ${OPENWRT_USER}@${OPENWRT_HOST} "$1" 2>/dev/null || return 1
}

# Función para ejecutar comandos SSH con sshpass (si hay password)
ssh_exec_pass() {
    if [ -n "$OPENWRT_PASS" ]; then
        sshpass -p "$OPENWRT_PASS" ssh -o StrictHostKeyChecking=no \
            -o UserKnownHostsFile=/dev/null -o ConnectTimeout=10 \
            ${OPENWRT_USER}@${OPENWRT_HOST} "$1" 2>/dev/null || return 1
    else
        ssh_exec "$1"
    fi
}

# Intentar conectar a OpenWRT VM vía SSH (múltiples intentos)
echo "Waiting for OpenWRT SSH to be ready..."
for i in $(seq 1 30); do
    if ssh_exec "echo 'SSH connection successful'" >/dev/null 2>&1; then
        echo "✓ SSH connection established"
        break
    fi
    
    if [ $i -eq 30 ]; then
        echo "⚠ SSH connection failed after 30 attempts"
        echo "Proceeding without tcpdump installation..."
        exit 0
    fi
    
    sleep 2
done

# Instalar tcpdump dentro de la VM OpenWRT
echo "Installing tcpdump in OpenWRT VM..."

# Actualizar repositorios
if ssh_exec "opkg update"; then
    echo "✓ opkg update successful"
else
    echo "⚠ opkg update failed"
fi

# Instalar tcpdump
if ssh_exec "opkg install tcpdump"; then
    echo "✓ tcpdump installed successfully"
else
    echo "⚠ tcpdump installation failed, trying alternative packages..."
    
    # Intentar con paquetes alternativos
    ssh_exec "opkg install tcpdump-mini" || \
    ssh_exec "opkg install busybox-tcpdump" || \
    echo "⚠ All tcpdump packages failed to install"
fi

# Verificar instalación
if ssh_exec "tcpdump --version"; then
    echo "✓ tcpdump is working in OpenWRT VM"
else
    echo "⚠ tcpdump verification failed"
fi

# Copiar binario de NetTestLab a la VM
echo "Setting up NetTestLab in OpenWRT VM..."
if [ -f "/tmp/nettestlab-bin" ]; then
    # Copiar binario vía SCP
    scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
        /tmp/nettestlab-bin ${OPENWRT_USER}@${OPENWRT_HOST}:/tmp/nettestlab 2>/dev/null || \
        echo "⚠ Failed to copy NetTestLab binary to VM"
    
    # Hacer ejecutable
    ssh_exec "chmod +x /tmp/nettestlab" || true
fi

echo "OpenWRT VM setup completed"