/**
 * NetTestLab gRPC-Web Client
 * 
 * A JavaScript client for the NetTestLab gRPC services that can be used
 * directly in web browsers. This client uses the generated gRPC-Web files
 * and provides a simple promise-based API.
 * 
 * Usage:
 *   const client = new NetTestLabClient('http://localhost:8080');
 *   client.getSystemStatus().then(status => console.log(status));
 */
class NetTestLabClient {
  constructor(baseUrl = 'http://localhost:8080') {
    this.baseUrl = baseUrl;
    
    // Initialize gRPC-Web clients when the generated code is available
    if (typeof proto !== 'undefined' && proto.nettestlab && proto.nettestlab.v1) {
      this.monitoringClient = new proto.nettestlab.v1.MonitoringServiceClient(baseUrl, null, null);
      this.networkClient = new proto.nettestlab.v1.NetworkServiceClient(baseUrl, null, null);
      this.profileClient = new proto.nettestlab.v1.ProfileServiceClient(baseUrl, null, null);
    } else {
      console.warn('NetTestLab gRPC-Web generated code not found. Make sure to include the generated files.');
    }
  }

  // Helper method to promisify gRPC calls
  _promisifyCall(client, method, request) {
    return new Promise((resolve, reject) => {
      client[method](request, {}, (err, response) => {
        if (err) {
          reject(err);
        } else {
          resolve(response);
        }
      });
    });
  }

  // Monitoring Service Methods
  async getSystemStatus() {
    if (!this.monitoringClient) {
      throw new Error('Monitoring client not initialized');
    }
    const request = new proto.nettestlab.v1.GetSystemStatusRequest();
    return this._promisifyCall(this.monitoringClient, 'getSystemStatus', request);
  }

  async getSystemMetrics() {
    if (!this.monitoringClient) {
      throw new Error('Monitoring client not initialized');
    }
    const request = new proto.nettestlab.v1.GetSystemMetricsRequest();
    return this._promisifyCall(this.monitoringClient, 'getSystemMetrics', request);
  }

  // Network Service Methods
  async listInterfaces() {
    if (!this.networkClient) {
      throw new Error('Network client not initialized');
    }
    const request = new proto.nettestlab.v1.ListInterfacesRequest();
    return this._promisifyCall(this.networkClient, 'listInterfaces', request);
  }

  async getInterface(interfaceId) {
    if (!this.networkClient) {
      throw new Error('Network client not initialized');
    }
    const request = new proto.nettestlab.v1.GetInterfaceRequest();
    request.setInterfaceId(interfaceId);
    return this._promisifyCall(this.networkClient, 'getInterface', request);
  }

  async updateInterface(interfaceId, enabled) {
    if (!this.networkClient) {
      throw new Error('Network client not initialized');
    }
    const request = new proto.nettestlab.v1.UpdateInterfaceRequest();
    request.setInterfaceId(interfaceId);
    request.setEnabled(enabled);
    return this._promisifyCall(this.networkClient, 'updateInterface', request);
  }

  async listClients() {
    if (!this.networkClient) {
      throw new Error('Network client not initialized');
    }
    const request = new proto.nettestlab.v1.ListClientsRequest();
    return this._promisifyCall(this.networkClient, 'listClients', request);
  }

  async getClient(clientId) {
    if (!this.networkClient) {
      throw new Error('Network client not initialized');
    }
    const request = new proto.nettestlab.v1.GetClientRequest();
    request.setClientId(clientId);
    return this._promisifyCall(this.networkClient, 'getClient', request);
  }

  async createClient(name, macAddress, profileId) {
    if (!this.networkClient) {
      throw new Error('Network client not initialized');
    }
    const request = new proto.nettestlab.v1.CreateClientRequest();
    request.setName(name);
    request.setMacAddress(macAddress);
    request.setProfileId(profileId);
    return this._promisifyCall(this.networkClient, 'createClient', request);
  }

  async updateClient(clientId, name, profileId) {
    if (!this.networkClient) {
      throw new Error('Network client not initialized');
    }
    const request = new proto.nettestlab.v1.UpdateClientRequest();
    request.setClientId(clientId);
    if (name) request.setName(name);
    if (profileId) request.setProfileId(profileId);
    return this._promisifyCall(this.networkClient, 'updateClient', request);
  }

  async deleteClient(clientId) {
    if (!this.networkClient) {
      throw new Error('Network client not initialized');
    }
    const request = new proto.nettestlab.v1.DeleteClientRequest();
    request.setClientId(clientId);
    return this._promisifyCall(this.networkClient, 'deleteClient', request);
  }

  // Profile Service Methods
  async listProfiles() {
    if (!this.profileClient) {
      throw new Error('Profile client not initialized');
    }
    const request = new proto.nettestlab.v1.ListProfilesRequest();
    return this._promisifyCall(this.profileClient, 'listProfiles', request);
  }

  async getProfile(profileId) {
    if (!this.profileClient) {
      throw new Error('Profile client not initialized');
    }
    const request = new proto.nettestlab.v1.GetProfileRequest();
    request.setProfileId(profileId);
    return this._promisifyCall(this.profileClient, 'getProfile', request);
  }

  async createProfile(name, downloadSpeed, uploadSpeed) {
    if (!this.profileClient) {
      throw new Error('Profile client not initialized');
    }
    const request = new proto.nettestlab.v1.CreateProfileRequest();
    request.setName(name);
    request.setDownloadSpeed(downloadSpeed);
    request.setUploadSpeed(uploadSpeed);
    return this._promisifyCall(this.profileClient, 'createProfile', request);
  }

  async updateProfile(profileId, name, downloadSpeed, uploadSpeed) {
    if (!this.profileClient) {
      throw new Error('Profile client not initialized');
    }
    const request = new proto.nettestlab.v1.UpdateProfileRequest();
    request.setProfileId(profileId);
    if (name) request.setName(name);
    if (downloadSpeed) request.setDownloadSpeed(downloadSpeed);
    if (uploadSpeed) request.setUploadSpeed(uploadSpeed);
    return this._promisifyCall(this.profileClient, 'updateProfile', request);
  }

  async deleteProfile(profileId) {
    if (!this.profileClient) {
      throw new Error('Profile client not initialized');
    }
    const request = new proto.nettestlab.v1.DeleteProfileRequest();
    request.setProfileId(profileId);
    return this._promisifyCall(this.profileClient, 'deleteProfile', request);
  }

  // Health check method
  isConnected() {
    return !!(this.monitoringClient && this.networkClient && this.profileClient);
  }

  // Get connection status
  getConnectionInfo() {
    return {
      baseUrl: this.baseUrl,
      connected: this.isConnected(),
      clients: {
        monitoring: !!this.monitoringClient,
        network: !!this.networkClient,
        profile: !!this.profileClient
      }
    };
  }
}

// Export for use in different environments
if (typeof module !== 'undefined' && module.exports) {
  module.exports = NetTestLabClient;
}
if (typeof window !== 'undefined') {
  window.NetTestLabClient = NetTestLabClient;
}