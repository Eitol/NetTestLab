import { MonitoringServiceClient } from './nettestlab/v1/monitoring_grpc_web_pb';
import { NetworkServiceClient } from './nettestlab/v1/network_grpc_web_pb';
import { ProfileServiceClient } from './nettestlab/v1/profiles_grpc_web_pb';

import {
  GetSystemStatusRequest,
  GetSystemStatusResponse,
  GetSystemMetricsRequest,
  GetSystemMetricsResponse,
} from './nettestlab/v1/monitoring_pb';

import {
  ListInterfacesRequest,
  ListInterfacesResponse,
  GetInterfaceRequest,
  GetInterfaceResponse,
  UpdateInterfaceRequest,
  UpdateInterfaceResponse,
  ListClientsRequest,
  ListClientsResponse,
  GetClientRequest,
  GetClientResponse,
  CreateClientRequest,
  CreateClientResponse,
  UpdateClientRequest,
  UpdateClientResponse,
  DeleteClientRequest,
  DeleteClientResponse,
} from './nettestlab/v1/network_pb';

import {
  ListProfilesRequest,
  ListProfilesResponse,
  GetProfileRequest,
  GetProfileResponse,
  CreateProfileRequest,
  CreateProfileResponse,
  UpdateProfileRequest,
  UpdateProfileResponse,
  DeleteProfileRequest,
  DeleteProfileResponse,
} from './nettestlab/v1/profiles_pb';

export interface NetTestLabClientConfig {
  baseUrl?: string;
  enableTls?: boolean;
}

/**
 * NetTestLab gRPC-Web Client
 * 
 * Provides a TypeScript interface to the NetTestLab gRPC services
 * for use in web browsers.
 */
export class NetTestLabClient {
  private monitoringClient: MonitoringServiceClient;
  private networkClient: NetworkServiceClient;
  private profileClient: ProfileServiceClient;

  constructor(config: NetTestLabClientConfig = {}) {
    const baseUrl = config.baseUrl || 'http://localhost:8080';
    const enableTls = config.enableTls || false;

    this.monitoringClient = new MonitoringServiceClient(baseUrl, null, null);
    this.networkClient = new NetworkServiceClient(baseUrl, null, null);
    this.profileClient = new ProfileServiceClient(baseUrl, null, null);
  }

  // Monitoring Service Methods
  async getSystemStatus(): Promise<GetSystemStatusResponse> {
    const request = new GetSystemStatusRequest();
    return new Promise((resolve, reject) => {
      this.monitoringClient.getSystemStatus(request, {}, (err, response) => {
        if (err) {
          reject(err);
        } else {
          resolve(response);
        }
      });
    });
  }

  async getSystemMetrics(): Promise<GetSystemMetricsResponse> {
    const request = new GetSystemMetricsRequest();
    return new Promise((resolve, reject) => {
      this.monitoringClient.getSystemMetrics(request, {}, (err, response) => {
        if (err) {
          reject(err);
        } else {
          resolve(response);
        }
      });
    });
  }

  // Network Service Methods
  async listInterfaces(): Promise<ListInterfacesResponse> {
    const request = new ListInterfacesRequest();
    return new Promise((resolve, reject) => {
      this.networkClient.listInterfaces(request, {}, (err, response) => {
        if (err) {
          reject(err);
        } else {
          resolve(response);
        }
      });
    });
  }

  async getInterface(interfaceId: string): Promise<GetInterfaceResponse> {
    const request = new GetInterfaceRequest();
    request.setInterfaceId(interfaceId);
    return new Promise((resolve, reject) => {
      this.networkClient.getInterface(request, {}, (err, response) => {
        if (err) {
          reject(err);
        } else {
          resolve(response);
        }
      });
    });
  }

  async updateInterface(interfaceId: string, enabled: boolean): Promise<UpdateInterfaceResponse> {
    const request = new UpdateInterfaceRequest();
    request.setInterfaceId(interfaceId);
    request.setEnabled(enabled);
    return new Promise((resolve, reject) => {
      this.networkClient.updateInterface(request, {}, (err, response) => {
        if (err) {
          reject(err);
        } else {
          resolve(response);
        }
      });
    });
  }

  async listClients(): Promise<ListClientsResponse> {
    const request = new ListClientsRequest();
    return new Promise((resolve, reject) => {
      this.networkClient.listClients(request, {}, (err, response) => {
        if (err) {
          reject(err);
        } else {
          resolve(response);
        }
      });
    });
  }

  async getClient(clientId: string): Promise<GetClientResponse> {
    const request = new GetClientRequest();
    request.setClientId(clientId);
    return new Promise((resolve, reject) => {
      this.networkClient.getClient(request, {}, (err, response) => {
        if (err) {
          reject(err);
        } else {
          resolve(response);
        }
      });
    });
  }

  async createClient(name: string, macAddress: string, profileId: string): Promise<CreateClientResponse> {
    const request = new CreateClientRequest();
    request.setName(name);
    request.setMacAddress(macAddress);
    request.setProfileId(profileId);
    return new Promise((resolve, reject) => {
      this.networkClient.createClient(request, {}, (err, response) => {
        if (err) {
          reject(err);
        } else {
          resolve(response);
        }
      });
    });
  }

  async updateClient(clientId: string, name?: string, profileId?: string): Promise<UpdateClientResponse> {
    const request = new UpdateClientRequest();
    request.setClientId(clientId);
    if (name) request.setName(name);
    if (profileId) request.setProfileId(profileId);
    return new Promise((resolve, reject) => {
      this.networkClient.updateClient(request, {}, (err, response) => {
        if (err) {
          reject(err);
        } else {
          resolve(response);
        }
      });
    });
  }

  async deleteClient(clientId: string): Promise<DeleteClientResponse> {
    const request = new DeleteClientRequest();
    request.setClientId(clientId);
    return new Promise((resolve, reject) => {
      this.networkClient.deleteClient(request, {}, (err, response) => {
        if (err) {
          reject(err);
        } else {
          resolve(response);
        }
      });
    });
  }

  // Profile Service Methods
  async listProfiles(): Promise<ListProfilesResponse> {
    const request = new ListProfilesRequest();
    return new Promise((resolve, reject) => {
      this.profileClient.listProfiles(request, {}, (err, response) => {
        if (err) {
          reject(err);
        } else {
          resolve(response);
        }
      });
    });
  }

  async getProfile(profileId: string): Promise<GetProfileResponse> {
    const request = new GetProfileRequest();
    request.setProfileId(profileId);
    return new Promise((resolve, reject) => {
      this.profileClient.getProfile(request, {}, (err, response) => {
        if (err) {
          reject(err);
        } else {
          resolve(response);
        }
      });
    });
  }

  async createProfile(name: string, downloadSpeed: number, uploadSpeed: number): Promise<CreateProfileResponse> {
    const request = new CreateProfileRequest();
    request.setName(name);
    request.setDownloadSpeed(downloadSpeed);
    request.setUploadSpeed(uploadSpeed);
    return new Promise((resolve, reject) => {
      this.profileClient.createProfile(request, {}, (err, response) => {
        if (err) {
          reject(err);
        } else {
          resolve(response);
        }
      });
    });
  }

  async updateProfile(profileId: string, name?: string, downloadSpeed?: number, uploadSpeed?: number): Promise<UpdateProfileResponse> {
    const request = new UpdateProfileRequest();
    request.setProfileId(profileId);
    if (name) request.setName(name);
    if (downloadSpeed) request.setDownloadSpeed(downloadSpeed);
    if (uploadSpeed) request.setUploadSpeed(uploadSpeed);
    return new Promise((resolve, reject) => {
      this.profileClient.updateProfile(request, {}, (err, response) => {
        if (err) {
          reject(err);
        } else {
          resolve(response);
        }
      });
    });
  }

  async deleteProfile(profileId: string): Promise<DeleteProfileResponse> {
    const request = new DeleteProfileRequest();
    request.setProfileId(profileId);
    return new Promise((resolve, reject) => {
      this.profileClient.deleteProfile(request, {}, (err, response) => {
        if (err) {
          reject(err);
        } else {
          resolve(response);
        }
      });
    });
  }
}

// Re-export types for convenience
export {
  GetSystemStatusResponse,
  GetSystemMetricsResponse,
  ListInterfacesResponse,
  GetInterfaceResponse,
  UpdateInterfaceResponse,
  ListClientsResponse,
  GetClientResponse,
  CreateClientResponse,
  UpdateClientResponse,
  DeleteClientResponse,
  ListProfilesResponse,
  GetProfileResponse,
  CreateProfileResponse,
  UpdateProfileResponse,
  DeleteProfileResponse,
};