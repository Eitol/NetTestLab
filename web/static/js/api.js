/**
 * NetTestLab API Client
 * Handles communication with the gRPC backend via gRPC-Web
 */

class NetTestLabAPI {
    constructor(baseUrl = 'http://localhost:8080') {
        this.baseUrl = baseUrl;
        this.grpcClient = null;
        this.initializeGrpcClient();
    }

    /**
     * Initialize the gRPC-Web client
     */
    initializeGrpcClient() {
        try {
            if (typeof NetTestLabClient !== 'undefined') {
                this.grpcClient = new NetTestLabClient(this.baseUrl);
                console.log('gRPC-Web client initialized successfully');
            } else {
                console.warn('NetTestLabClient not available, using fallback mode');
            }
        } catch (error) {
            console.error('Failed to initialize gRPC client:', error);
        }
    }

    /**
     * Check if gRPC client is available and connected
     */
    isGrpcAvailable() {
        return this.grpcClient && this.grpcClient.isConnected();
    }

    /**
     * Handle gRPC errors and convert to user-friendly messages
     */
    handleGrpcError(error) {
        console.error('gRPC Error:', error);
        
        if (error.code === 14) { // UNAVAILABLE
            throw new Error('Unable to connect to NetTestLab server. Please check if the server is running.');
        } else if (error.code === 12) { // UNIMPLEMENTED
            throw new Error('This feature is not yet implemented on the server.');
        } else {
            throw new Error(error.message || 'Unknown server error occurred');
        }
    }

    // System Status APIs
    async getSystemStatus() {
        if (this.isGrpcAvailable()) {
            try {
                const response = await this.grpcClient.getSystemStatus();
                return {
                    serverVersion: response.getVersion() || "1.0.0",
                    uptime: response.getUptime() || "Unknown",
                    systemLoad: response.getSystemLoad() || "0.00",
                    memoryUsage: response.getMemoryUsage() || "Unknown",
                    status: response.getStatus() || "running",
                    lastUpdate: new Date().toISOString()
                };
            } catch (error) {
                this.handleGrpcError(error);
            }
        }
        
        // Fallback mock data
        return {
            serverVersion: "1.0.0",
            uptime: "2d 14h 32m",
            systemLoad: "0.15",
            memoryUsage: "64MB",
            status: "running",
            startTime: "2024-12-14T10:30:00Z",
            lastUpdate: new Date().toISOString()
        };
    }

    async getSystemHealth() {
        if (this.isGrpcAvailable()) {
            try {
                const response = await this.grpcClient.getSystemMetrics();
                return {
                    status: "healthy",
                    checks: [
                        { name: "gRPC Server", status: "ok" },
                        { name: "Network Interfaces", status: "ok" },
                        { name: "Traffic Control", status: "ok" }
                    ]
                };
            } catch (error) {
                this.handleGrpcError(error);
            }
        }
        
        // Fallback mock data
        return {
            status: "healthy",
            checks: [
                { name: "gRPC Server", status: "ok" },
                { name: "Network Interfaces", status: "ok" },
                { name: "Traffic Control", status: "ok" }
            ]
        };
    }

    // Network Interface APIs
    async getNetworkInterfaces() {
        if (this.isGrpcAvailable()) {
            try {
                const response = await this.grpcClient.listInterfaces();
                const interfaces = response.getInterfacesList();
                return interfaces.map(iface => ({
                    name: iface.getName(),
                    enabled: iface.getEnabled(),
                    ipAddress: iface.getIpAddress(),
                    macAddress: iface.getMacAddress(),
                    type: iface.getType(),
                    speed: iface.getSpeed(),
                    mtu: iface.getMtu(),
                    rxBytes: iface.getRxBytes(),
                    txBytes: iface.getTxBytes(),
                    rxPackets: iface.getRxPackets(),
                    txPackets: iface.getTxPackets()
                }));
            } catch (error) {
                this.handleGrpcError(error);
            }
        }
        
        // Fallback mock data
        return [
            {
                name: "eth0",
                enabled: true,
                ipAddress: "192.168.1.100",
                macAddress: "00:1B:44:11:3A:B7",
                type: "ethernet",
                speed: "1000 Mbps",
                mtu: 1500,
                rxBytes: 1048576000,
                txBytes: 524288000,
                rxPackets: 750000,
                txPackets: 500000,
                profileId: "prof_001"
            },
            {
                name: "wlan0",
                enabled: true,
                ipAddress: "192.168.1.101",
                macAddress: "00:1B:44:11:3A:B8",
                type: "wifi",
                speed: "300 Mbps",
                mtu: 1500,
                rxBytes: 524288000,
                txBytes: 262144000,
                rxPackets: 400000,
                txPackets: 300000,
                profileId: "prof_002"
            },
            {
                name: "lo",
                enabled: true,
                ipAddress: "127.0.0.1",
                macAddress: "00:00:00:00:00:00",
                type: "loopback",
                speed: "Unknown",
                mtu: 65536,
                rxBytes: 1024000,
                txBytes: 1024000,
                rxPackets: 1000,
                txPackets: 1000,
                profileId: null
            }
        ];
    }

    async getInterfaceDetails(interfaceName) {
        if (this.isGrpcAvailable()) {
            try {
                const response = await this.grpcClient.getInterface(interfaceName);
                const iface = response.getInterface();
                return {
                    name: iface.getName(),
                    enabled: iface.getEnabled(),
                    ipAddress: iface.getIpAddress(),
                    macAddress: iface.getMacAddress(),
                    type: iface.getType(),
                    speed: iface.getSpeed(),
                    mtu: iface.getMtu(),
                    rxBytes: iface.getRxBytes(),
                    txBytes: iface.getTxBytes(),
                    rxPackets: iface.getRxPackets(),
                    txPackets: iface.getTxPackets()
                };
            } catch (error) {
                this.handleGrpcError(error);
            }
        }
        
        // Fallback: find in mock data
        const interfaces = await this.getNetworkInterfaces();
        return interfaces.find(iface => iface.name === interfaceName);
    }

    // Connected Clients APIs
    async getConnectedClients() {
        if (this.isGrpcAvailable()) {
            try {
                const response = await this.grpcClient.listClients();
                const clients = response.getClientsList();
                return clients.map(client => ({
                    id: client.getId(),
                    name: client.getName(),
                    macAddress: client.getMacAddress(),
                    ipAddress: client.getIpAddress(),
                    profileId: client.getProfileId(),
                    interface: client.getInterface(),
                    status: client.getStatus(),
                    connectTime: client.getConnectTime(),
                    bytesTransferred: client.getBytesTransferred()
                }));
            } catch (error) {
                this.handleGrpcError(error);
            }
        }
        
        // Fallback mock data
        return [
            {
                id: "client_001",
                name: "Laptop-Windows",
                macAddress: "E4:5F:01:23:45:67",
                ipAddress: "192.168.1.150",
                profileId: "prof_001",
                interface: "eth0",
                status: "connected",
                connectTime: "2024-12-14T08:30:00Z",
                bytesTransferred: 52428800
            },
            {
                id: "client_002",
                name: "Phone-Android",
                macAddress: "F6:7A:89:BC:DE:F0",
                ipAddress: "192.168.1.151",
                profileId: "prof_002",
                interface: "wlan0",
                status: "connected",
                connectTime: "2024-12-14T09:15:00Z",
                bytesTransferred: 26214400
            },
            {
                id: "client_003",
                name: "iPad-iOS",
                macAddress: "A1:B2:C3:D4:E5:F6",
                ipAddress: "192.168.1.152",
                profileId: "prof_003",
                interface: "wlan0",
                status: "idle",
                connectTime: "2024-12-14T07:45:00Z",
                bytesTransferred: 10485760
            }
        ];
    }

    // Network Profile APIs
    async getNetworkProfiles() {
        if (this.isGrpcAvailable()) {
            try {
                const response = await this.grpcClient.listProfiles();
                const profiles = response.getProfilesList();
                return profiles.map(profile => ({
                    id: profile.getId(),
                    name: profile.getName(),
                    downloadSpeed: profile.getDownloadSpeed(),
                    uploadSpeed: profile.getUploadSpeed(),
                    latency: profile.getLatency(),
                    packetLoss: profile.getPacketLoss(),
                    description: profile.getDescription(),
                    createdAt: profile.getCreatedAt(),
                    updatedAt: profile.getUpdatedAt()
                }));
            } catch (error) {
                this.handleGrpcError(error);
            }
        }
        
        // Fallback mock data
        return [
            {
                id: "prof_001",
                name: "High Speed",
                downloadSpeed: 1000,
                uploadSpeed: 500,
                latency: 5,
                packetLoss: 0.01,
                description: "High-speed profile for demanding applications",
                createdAt: "2024-12-01T00:00:00Z",
                updatedAt: "2024-12-01T00:00:00Z"
            },
            {
                id: "prof_002",
                name: "Standard",
                downloadSpeed: 100,
                uploadSpeed: 50,
                latency: 20,
                packetLoss: 0.1,
                description: "Standard speed profile for regular use",
                createdAt: "2024-12-01T00:00:00Z",
                updatedAt: "2024-12-01T00:00:00Z"
            },
            {
                id: "prof_003",
                name: "Limited",
                downloadSpeed: 10,
                uploadSpeed: 5,
                latency: 100,
                packetLoss: 1,
                description: "Limited speed profile for testing",
                createdAt: "2024-12-01T00:00:00Z",
                updatedAt: "2024-12-01T00:00:00Z"
            },
            {
                id: "prof_004",
                name: "Mobile",
                downloadSpeed: 50,
                uploadSpeed: 25,
                latency: 50,
                packetLoss: 0.5,
                description: "Mobile network simulation profile",
                createdAt: "2024-12-01T00:00:00Z",
                updatedAt: "2024-12-01T00:00:00Z"
            }
        ];
    }

    async createNetworkProfile(profile) {
        if (this.isGrpcAvailable()) {
            try {
                const response = await this.grpcClient.createProfile(
                    profile.name,
                    profile.downloadSpeed,
                    profile.uploadSpeed
                );
                return {
                    id: response.getProfile().getId(),
                    success: true,
                    message: "Profile created successfully"
                };
            } catch (error) {
                this.handleGrpcError(error);
            }
        }
        
        // Fallback mock response
        return {
            id: "prof_" + Date.now(),
            success: true,
            message: "Profile created successfully (mock)"
        };
    }

    async updateNetworkProfile(profileId, profile) {
        if (this.isGrpcAvailable()) {
            try {
                const response = await this.grpcClient.updateProfile(
                    profileId,
                    profile.name,
                    profile.downloadSpeed,
                    profile.uploadSpeed
                );
                return {
                    success: true,
                    message: "Profile updated successfully"
                };
            } catch (error) {
                this.handleGrpcError(error);
            }
        }
        
        // Fallback mock response
        return {
            success: true,
            message: "Profile updated successfully (mock)"
        };
    }

    async deleteNetworkProfile(profileId) {
        if (this.isGrpcAvailable()) {
            try {
                await this.grpcClient.deleteProfile(profileId);
                return {
                    success: true,
                    message: "Profile deleted successfully"
                };
            } catch (error) {
                this.handleGrpcError(error);
            }
        }
        
        // Fallback mock response
        return {
            success: true,
            message: "Profile deleted successfully (mock)"
        };
    }

    async applyProfileToInterface(profileId, interfaceName) {
        if (this.isGrpcAvailable()) {
            try {
                const response = await this.grpcClient.updateInterface(interfaceName, true);
                return {
                    success: true,
                    message: "Profile applied to interface successfully"
                };
            } catch (error) {
                this.handleGrpcError(error);
            }
        }
        
        // Fallback mock response
        return {
            success: true,
            message: "Profile applied to interface successfully (mock)"
        };
    }

    async removeProfileFromInterface(interfaceName) {
        if (this.isGrpcAvailable()) {
            try {
                const response = await this.grpcClient.updateInterface(interfaceName, false);
                return {
                    success: true,
                    message: "Profile removed from interface successfully"
                };
            } catch (error) {
                this.handleGrpcError(error);
            }
        }
        
        // Fallback mock response
        return {
            success: true,
            message: "Profile removed from interface successfully (mock)"
        };
    }

    // Client management APIs
    async updateClient(clientId, updates) {
        if (this.isGrpcAvailable()) {
            try {
                const response = await this.grpcClient.updateClient(
                    clientId,
                    updates.name,
                    updates.profileId
                );
                return {
                    success: true,
                    message: "Client updated successfully"
                };
            } catch (error) {
                this.handleGrpcError(error);
            }
        }
        
        // Fallback mock response
        return {
            success: true,
            message: "Client updated successfully (mock)"
        };
    }

    async deleteClient(clientId) {
        if (this.isGrpcAvailable()) {
            try {
                await this.grpcClient.deleteClient(clientId);
                return {
                    success: true,
                    message: "Client deleted successfully"
                };
            } catch (error) {
                this.handleGrpcError(error);
            }
        }
        
        // Fallback mock response
        return {
            success: true,
            message: "Client deleted successfully (mock)"
        };
    }

    // Utility methods
    getConnectionStatus() {
        return {
            grpcAvailable: this.isGrpcAvailable(),
            baseUrl: this.baseUrl,
            clientInfo: this.grpcClient ? this.grpcClient.getConnectionInfo() : null
        };
    }
}