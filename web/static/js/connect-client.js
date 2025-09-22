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

    // Traffic Capture Service methods - REAL API calls
    async listDevices() {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.TrafficCaptureService/ListDevices`, {
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
            console.error('Failed to list devices:', error);
            throw error;
        }
    }

    async listUrlTargets() {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.TrafficCaptureService/ListUrlTargets`, {
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
            console.error('Failed to list URL targets:', error);
            throw error;
        }
    }

    async startCapture(captureRequest) {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.TrafficCaptureService/StartCapture`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(captureRequest)
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to start capture:', error);
            throw error;
        }
    }

    async stopCapture(request) {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.TrafficCaptureService/StopCapture`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(request)
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to stop capture:', error);
            throw error;
        }
    }

    async getCaptureStatus(request) {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.TrafficCaptureService/GetCaptureStatus`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(request)
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to get capture status:', error);
            throw error;
        }
    }

    async createUrlTarget(target) {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.TrafficCaptureService/CreateUrlTarget`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(target)
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to create URL target:', error);
            throw error;
        }
    }

    async createDevice(device) {
        try {
            // Use RegisterDevice endpoint since there's no CreateDevice
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.TrafficCaptureService/RegisterDevice`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    mac_address: device.mac_address,
                    device_name: device.device_name,
                    device_model: device.device_model || device.device_type
                })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return { created: data.created, success: true };
        } catch (error) {
            console.error('Failed to create device:', error);
            throw error;
        }
    }

    async getDevice(deviceId) {
        // Since there's no GetDevice endpoint, we'll get the device from the list
        try {
            const devicesResponse = await this.listDevices();
            const device = devicesResponse.devices?.find(d => d.id === deviceId);
            if (device) {
                return { device };
            } else {
                throw new Error('Device not found');
            }
        } catch (error) {
            console.error('Failed to get device:', error);
            throw error;
        }
    }    async updateDevice(deviceId, device) {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.TrafficCaptureService/UpdateDevice`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ device_id: deviceId, ...device })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to update device:', error);
            throw error;
        }
    }

    async deleteDevice(deviceId) {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.TrafficCaptureService/DeleteDevice`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ device_id: deviceId })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to delete device:', error);
            throw error;
        }
    }

    async getUrlTarget(targetId) {
        // Since there's no GetUrlTarget endpoint, we'll get the target from the list
        try {
            const targetsResponse = await this.listUrlTargets();
            const target = targetsResponse.targets?.find(t => t.id === targetId);
            if (target) {
                return { target };
            } else {
                throw new Error('URL Target not found');
            }
        } catch (error) {
            console.error('Failed to get URL target:', error);
            throw error;
        }
    }    async updateUrlTarget(targetId, target) {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.TrafficCaptureService/UpdateUrlTarget`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ target_id: targetId, ...target })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to update URL target:', error);
            throw error;
        }
    }

    async deleteUrlTarget(targetId) {
        try {
            const response = await fetch(`${this.baseUrl}/nettestlab.v1.TrafficCaptureService/DeleteUrlTarget`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ target_id: targetId })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to delete URL target:', error);
            throw error;
        }
    }
}