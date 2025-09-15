// Connect-compatible client for NetTestLab
// Using standard fetch API with Connect protocol format

export class NetTestLabConnectClient {
    constructor() {
        // Use current host and port instead of hardcoded localhost
        const currentHost = window.location.hostname;
        const currentPort = window.location.port || '8080';
        this.baseUrl = `http://${currentHost}:${currentPort}`;
    }

    // Network Control Service methods
    async getSystemStatus() {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.NetworkControlService/GetSystemStatus`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({})
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to get system status:', error);
            throw error;
        }
    }

    async applyNetworkConditions(interface_, conditions, direction = 0) {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.NetworkControlService/ApplyNetworkConditions`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    interface: interface_, 
                    conditions,
                    direction
                })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to apply network conditions:', error);
            throw error;
        }
    }

    async resetNetworkConditions(interface_, direction = 0) {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.NetworkControlService/ResetNetworkConditions`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    interface: interface_,
                    direction
                })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to reset network conditions:', error);
            throw error;
        }
    }

    async getNetworkConditions(interface_) {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.NetworkControlService/GetNetworkConditions`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    interface: interface_
                })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to get network conditions:', error);
            throw error;
        }
    }

    // Profile Service methods
    async listProfiles(type = 0, builtInOnly = false) {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.ProfileService/ListProfiles`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    type,
                    builtInOnly
                })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to list profiles:', error);
            throw error;
        }
    }

    async getProfile(name) {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.ProfileService/GetProfile`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ name })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to get profile:', error);
            throw error;
        }
    }

    async createProfile(profile) {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.ProfileService/CreateProfile`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ profile })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to create profile:', error);
            throw error;
        }
    }

    async updateProfile(name, profile) {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.ProfileService/UpdateProfile`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ name, profile })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to update profile:', error);
            throw error;
        }
    }

    async deleteProfile(name) {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.ProfileService/DeleteProfile`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ name })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to delete profile:', error);
            throw error;
        }
    }

    async applyProfile(profileName, interface_, direction = 0) {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.ProfileService/ApplyProfile`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    profileName,
                    interface: interface_,
                    direction
                })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to apply profile:', error);
            throw error;
        }
    }

    // Monitoring Service methods
    async getHealth() {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.MonitoringService/GetHealth`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({})
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to get health:', error);
            throw error;
        }
    }

    async getMetrics() {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.MonitoringService/GetMetrics`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({})
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to get metrics:', error);
            throw error;
        }
    }
}