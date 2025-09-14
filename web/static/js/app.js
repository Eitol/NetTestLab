/**
 * NetTestLab Web Application
 * Main application logic and UI management
 */

class NetTestLabApp {
    constructor() {
        this.api = window.netTestLabAPI;
        this.currentTab = 'dashboard';
        this.refreshInterval = null;
        this.profiles = [];
        this.interfaces = [];
        this.clients = [];
        
        this.init();
    }

    async init() {
        console.log('Initializing NetTestLab Web Interface...');
        
        // Set up event listeners
        this.setupEventListeners();
        
        // Load initial data
        await this.loadDashboardData();
        
        // Start auto-refresh
        this.startAutoRefresh();
        
        console.log('NetTestLab Web Interface initialized');
    }

    setupEventListeners() {
        // Profile type change handler
        document.getElementById('profileType').addEventListener('change', (e) => {
            this.onProfileTypeChange(e.target.value);
        });

        // Form submission handlers
        document.getElementById('profileForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.saveProfile();
        });
    }

    // Tab Management
    showTab(tabName) {
        // Hide all tabs
        document.querySelectorAll('.tab-content').forEach(tab => {
            tab.style.display = 'none';
        });

        // Show selected tab
        const targetTab = document.getElementById(`${tabName}-tab`);
        if (targetTab) {
            targetTab.style.display = 'block';
        }

        // Update navigation
        document.querySelectorAll('.navbar-nav .nav-link').forEach(link => {
            link.classList.remove('active');
        });
        
        event.target.classList.add('active');
        this.currentTab = tabName;

        // Load tab-specific data
        this.loadTabData(tabName);
    }

    async loadTabData(tabName) {
        switch (tabName) {
            case 'dashboard':
                await this.loadDashboardData();
                break;
            case 'interfaces':
                await this.loadInterfacesData();
                break;
            case 'clients':
                await this.loadClientsData();
                break;
            case 'profiles':
                await this.loadProfilesData();
                break;
        }
    }

    // Dashboard Functions
    async loadDashboardData() {
        try {
            this.updateConnectionStatus('connecting');
            
            // Load system status
            const systemStatus = await this.api.getSystemStatus();
            this.updateSystemStatus(systemStatus);
            
            // Load quick stats
            await this.updateQuickStats();
            
            // Load recent activity
            const activity = await this.api.getRecentActivity();
            this.updateRecentActivity(activity);
            
            this.updateConnectionStatus('connected');
        } catch (error) {
            console.error('Failed to load dashboard data:', error);
            this.updateConnectionStatus('disconnected');
            this.showError('Failed to load dashboard data');
        }
    }

    updateSystemStatus(status) {
        document.getElementById('serverVersion').textContent = status.serverVersion;
        document.getElementById('uptime').textContent = status.uptime;
        document.getElementById('systemLoad').textContent = status.systemLoad;
        document.getElementById('memoryUsage').textContent = status.memoryUsage;
    }

    async updateQuickStats() {
        try {
            const [interfaces, clients, profiles] = await Promise.all([
                this.api.getNetworkInterfaces(),
                this.api.getConnectedClients(),
                this.api.getNetworkProfiles()
            ]);

            const activeInterfaces = interfaces.filter(iface => iface.status === 'up').length;
            const connectedClients = clients.filter(client => client.status === 'online').length;
            const activeProfiles = profiles.filter(profile => profile.active).length;

            document.getElementById('activeInterfaces').textContent = activeInterfaces;
            document.getElementById('connectedClients').textContent = connectedClients;
            document.getElementById('activeProfiles').textContent = activeProfiles;
            document.getElementById('totalProfiles').textContent = profiles.length;
        } catch (error) {
            console.error('Failed to update quick stats:', error);
        }
    }

    updateRecentActivity(activities) {
        const container = document.getElementById('recentActivity');
        
        if (!activities || activities.length === 0) {
            container.innerHTML = '<p class="text-muted text-center">No recent activity</p>';
            return;
        }

        const activityHtml = activities.map(activity => `
            <div class="activity-item d-flex align-items-start">
                <div class="activity-icon activity-${activity.type}">
                    <i class="${activity.icon}"></i>
                </div>
                <div class="flex-grow-1">
                    <div class="activity-message">${activity.message}</div>
                    <div class="activity-time">${this.api.formatTimeAgo(activity.timestamp)}</div>
                </div>
            </div>
        `).join('');

        container.innerHTML = activityHtml;
    }

    // Interfaces Functions
    async loadInterfacesData() {
        try {
            this.showLoading('interfacesList');
            
            this.interfaces = await this.api.getNetworkInterfaces();
            this.renderInterfaces();
        } catch (error) {
            console.error('Failed to load interfaces:', error);
            this.showError('Failed to load network interfaces');
        }
    }

    renderInterfaces() {
        const container = document.getElementById('interfacesList');
        
        if (!this.interfaces || this.interfaces.length === 0) {
            container.innerHTML = '<p class="text-muted text-center">No network interfaces found</p>';
            return;
        }

        const interfacesHtml = this.interfaces.map(iface => `
            <div class="col-md-6 col-lg-4 mb-3">
                <div class="card interface-card ${iface.type.toLowerCase()} ${iface.status === 'down' ? 'disabled' : ''}">
                    <div class="card-body">
                        <div class="d-flex justify-content-between align-items-start mb-2">
                            <h6 class="card-title mb-0">
                                <i class="bi bi-${this.getInterfaceIcon(iface.type)}"></i>
                                ${iface.name}
                            </h6>
                            <span class="interface-status status-${iface.status}">
                                ${iface.status.toUpperCase()}
                            </span>
                        </div>
                        <div class="mb-2">
                            <small class="text-muted">Type:</small>
                            <span class="fw-bold">${iface.type}</span>
                        </div>
                        <div class="mb-2">
                            <small class="text-muted">IP Addresses:</small>
                            <div>${iface.ipAddresses.length > 0 ? iface.ipAddresses.join(', ') : 'None'}</div>
                        </div>
                        ${iface.currentConditions ? `
                            <div class="mb-2">
                                <small class="text-muted">Active Profile:</small>
                                <div class="fw-bold text-primary">${iface.currentConditions.profile}</div>
                                <small class="text-muted">
                                    ${iface.currentConditions.latency} | 
                                    ${iface.currentConditions.packetLoss} | 
                                    ${iface.currentConditions.bandwidth}
                                </small>
                            </div>
                            <button class="btn btn-outline-danger btn-sm" onclick="app.removeProfileFromInterface('${iface.name}')">
                                <i class="bi bi-x-circle"></i> Remove Profile
                            </button>
                        ` : `
                            <button class="btn btn-outline-primary btn-sm" onclick="app.showApplyProfileModal('${iface.name}')">
                                <i class="bi bi-plus-circle"></i> Apply Profile
                            </button>
                        `}
                    </div>
                </div>
            </div>
        `).join('');

        container.innerHTML = interfacesHtml;
    }

    getInterfaceIcon(type) {
        switch (type.toLowerCase()) {
            case 'wifi': return 'wifi';
            case 'ethernet': return 'ethernet';
            case 'bridge': return 'diagram-3';
            default: return 'hdd-network';
        }
    }

    async refreshInterfaces() {
        await this.loadInterfacesData();
        this.showSuccess('Interfaces refreshed');
    }

    // Clients Functions
    async loadClientsData() {
        try {
            this.showLoading('clientsTableBody');
            
            this.clients = await this.api.getConnectedClients();
            this.renderClients();
        } catch (error) {
            console.error('Failed to load clients:', error);
            this.showError('Failed to load connected clients');
        }
    }

    renderClients() {
        const tbody = document.getElementById('clientsTableBody');
        
        if (!this.clients || this.clients.length === 0) {
            tbody.innerHTML = '<tr><td colspan="5" class="text-center text-muted">No connected clients</td></tr>';
            return;
        }

        const clientsHtml = this.clients.map(client => `
            <tr>
                <td>
                    <strong>${client.ipAddress}</strong>
                    <br>
                    <small class="text-muted">${client.macAddress}</small>
                </td>
                <td>${client.deviceName}</td>
                <td>
                    <span class="badge bg-secondary">${client.interface}</span>
                </td>
                <td>
                    ${this.api.formatTimeAgo(client.connectionTime)}
                    <br>
                    <small class="text-muted">${client.bandwidth}</small>
                </td>
                <td>
                    <span class="client-status client-${client.status}">
                        ${client.status.toUpperCase()}
                    </span>
                </td>
            </tr>
        `).join('');

        tbody.innerHTML = clientsHtml;
    }

    async refreshClients() {
        await this.loadClientsData();
        this.showSuccess('Connected clients refreshed');
    }

    // Profiles Functions
    async loadProfilesData() {
        try {
            this.showLoading('profilesList');
            
            this.profiles = await this.api.getNetworkProfiles();
            this.renderProfiles();
        } catch (error) {
            console.error('Failed to load profiles:', error);
            this.showError('Failed to load network profiles');
        }
    }

    renderProfiles() {
        const container = document.getElementById('profilesList');
        
        if (!this.profiles || this.profiles.length === 0) {
            container.innerHTML = '<p class="text-muted text-center">No network profiles found</p>';
            return;
        }

        const profilesHtml = this.profiles.map(profile => `
            <div class="col-md-6 col-lg-4 mb-3">
                <div class="card profile-card ${profile.active ? 'active' : ''}">
                    <div class="profile-header">
                        <div class="d-flex justify-content-between align-items-start">
                            <h6 class="mb-1">${profile.name}</h6>
                            <span class="profile-type-badge">${profile.type.toUpperCase()}</span>
                        </div>
                        <p class="mb-0" style="font-size: 0.875rem; opacity: 0.9;">
                            ${profile.description}
                        </p>
                    </div>
                    <div class="card-body">
                        <div class="row mb-2">
                            <div class="col-6">
                                <small class="text-muted">Latency</small>
                                <div class="fw-bold">${profile.conditions.latency}ms</div>
                            </div>
                            <div class="col-6">
                                <small class="text-muted">Packet Loss</small>
                                <div class="fw-bold">${profile.conditions.packetLoss}%</div>
                            </div>
                        </div>
                        <div class="row mb-3">
                            <div class="col-6">
                                <small class="text-muted">Download</small>
                                <div class="fw-bold">${this.api.formatBandwidth(profile.conditions.downloadBandwidth)}</div>
                            </div>
                            <div class="col-6">
                                <small class="text-muted">Upload</small>
                                <div class="fw-bold">${this.api.formatBandwidth(profile.conditions.uploadBandwidth)}</div>
                            </div>
                        </div>
                        <div class="btn-group w-100" role="group">
                            <button class="btn btn-outline-primary btn-sm" onclick="app.editProfile('${profile.id}')">
                                <i class="bi bi-pencil"></i> Edit
                            </button>
                            <button class="btn btn-outline-success btn-sm" onclick="app.showApplyProfileModalById('${profile.id}')">
                                <i class="bi bi-play"></i> Apply
                            </button>
                            <button class="btn btn-outline-danger btn-sm" onclick="app.deleteProfile('${profile.id}')">
                                <i class="bi bi-trash"></i> Delete
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        `).join('');

        container.innerHTML = profilesHtml;
    }

    // Profile Management
    showCreateProfileModal() {
        document.getElementById('profileModalTitle').textContent = 'Create Network Profile';
        document.getElementById('profileForm').reset();
        document.getElementById('profileId').value = '';
        
        const modal = new bootstrap.Modal(document.getElementById('profileModal'));
        modal.show();
    }

    editProfile(profileId) {
        const profile = this.profiles.find(p => p.id === profileId);
        if (!profile) {
            this.showError('Profile not found');
            return;
        }

        document.getElementById('profileModalTitle').textContent = 'Edit Network Profile';
        document.getElementById('profileId').value = profile.id;
        document.getElementById('profileName').value = profile.name;
        document.getElementById('profileType').value = profile.type;
        document.getElementById('latency').value = profile.conditions.latency;
        document.getElementById('packetLoss').value = profile.conditions.packetLoss;
        document.getElementById('downloadBandwidth').value = profile.conditions.downloadBandwidth;
        document.getElementById('uploadBandwidth').value = profile.conditions.uploadBandwidth;
        document.getElementById('profileDescription').value = profile.description;

        const modal = new bootstrap.Modal(document.getElementById('profileModal'));
        modal.show();
    }

    async saveProfile() {
        try {
            const formData = new FormData(document.getElementById('profileForm'));
            const profileId = document.getElementById('profileId').value;
            
            const profile = {
                name: document.getElementById('profileName').value,
                type: document.getElementById('profileType').value,
                description: document.getElementById('profileDescription').value,
                conditions: {
                    latency: parseInt(document.getElementById('latency').value) || 0,
                    packetLoss: parseFloat(document.getElementById('packetLoss').value) || 0,
                    downloadBandwidth: parseInt(document.getElementById('downloadBandwidth').value) || 0,
                    uploadBandwidth: parseInt(document.getElementById('uploadBandwidth').value) || 0
                }
            };

            if (profileId) {
                await this.api.updateNetworkProfile(profileId, profile);
                this.showSuccess('Profile updated successfully');
            } else {
                await this.api.createNetworkProfile(profile);
                this.showSuccess('Profile created successfully');
            }

            const modal = bootstrap.Modal.getInstance(document.getElementById('profileModal'));
            modal.hide();
            
            await this.loadProfilesData();
        } catch (error) {
            console.error('Failed to save profile:', error);
            this.showError('Failed to save profile');
        }
    }

    async deleteProfile(profileId) {
        if (!confirm('Are you sure you want to delete this profile?')) {
            return;
        }

        try {
            await this.api.deleteNetworkProfile(profileId);
            this.showSuccess('Profile deleted successfully');
            await this.loadProfilesData();
        } catch (error) {
            console.error('Failed to delete profile:', error);
            this.showError('Failed to delete profile');
        }
    }

    onProfileTypeChange(type) {
        // Set default values based on profile type
        const defaults = {
            '2g': { latency: 300, packetLoss: 5.0, download: 56, upload: 28 },
            '3g': { latency: 150, packetLoss: 2.0, download: 384, upload: 128 },
            '4g': { latency: 50, packetLoss: 0.5, download: 10000, upload: 5000 },
            '5g': { latency: 10, packetLoss: 0.1, download: 100000, upload: 50000 },
            'wifi': { latency: 20, packetLoss: 0.1, download: 50000, upload: 50000 },
            'satellite': { latency: 600, packetLoss: 1.0, download: 1000, upload: 256 }
        };

        if (defaults[type]) {
            document.getElementById('latency').value = defaults[type].latency;
            document.getElementById('packetLoss').value = defaults[type].packetLoss;
            document.getElementById('downloadBandwidth').value = defaults[type].download;
            document.getElementById('uploadBandwidth').value = defaults[type].upload;
        }
    }

    // Apply Profile to Interface
    showApplyProfileModal(interfaceName) {
        this.populateInterfaceSelect();
        document.getElementById('targetInterface').value = interfaceName;
        document.getElementById('selectedProfile').value = '';
        
        const modal = new bootstrap.Modal(document.getElementById('applyProfileModal'));
        modal.show();
    }

    showApplyProfileModalById(profileId) {
        const profile = this.profiles.find(p => p.id === profileId);
        if (!profile) {
            this.showError('Profile not found');
            return;
        }

        this.populateInterfaceSelect();
        document.getElementById('selectedProfile').value = profile.name;
        document.getElementById('selectedProfile').dataset.profileId = profileId;
        
        const modal = new bootstrap.Modal(document.getElementById('applyProfileModal'));
        modal.show();
    }

    populateInterfaceSelect() {
        const select = document.getElementById('targetInterface');
        select.innerHTML = this.interfaces.map(iface => 
            `<option value="${iface.name}">${iface.name} (${iface.type})</option>`
        ).join('');
    }

    async applyProfileToInterface() {
        try {
            const interfaceName = document.getElementById('targetInterface').value;
            const profileId = document.getElementById('selectedProfile').dataset.profileId;

            if (!profileId) {
                this.showError('Please select a profile');
                return;
            }

            await this.api.applyProfileToInterface(profileId, interfaceName);
            this.showSuccess(`Profile applied to ${interfaceName}`);
            
            const modal = bootstrap.Modal.getInstance(document.getElementById('applyProfileModal'));
            modal.hide();
            
            await this.loadInterfacesData();
        } catch (error) {
            console.error('Failed to apply profile:', error);
            this.showError('Failed to apply profile to interface');
        }
    }

    async removeProfileFromInterface(interfaceName) {
        if (!confirm(`Remove profile from ${interfaceName}?`)) {
            return;
        }

        try {
            await this.api.removeProfileFromInterface(interfaceName);
            this.showSuccess(`Profile removed from ${interfaceName}`);
            await this.loadInterfacesData();
        } catch (error) {
            console.error('Failed to remove profile:', error);
            this.showError('Failed to remove profile from interface');
        }
    }

    // Utility Functions
    updateConnectionStatus(status) {
        const statusElement = document.getElementById('connectionStatus');
        const iconClass = status === 'connected' ? 'bi-wifi' : 
                         status === 'connecting' ? 'bi-arrow-repeat' : 'bi-wifi-off';
        
        statusElement.className = `badge bg-${status === 'connected' ? 'success' : 
                                                status === 'connecting' ? 'warning' : 'danger'}`;
        statusElement.innerHTML = `<i class="${iconClass}"></i> ${status.charAt(0).toUpperCase() + status.slice(1)}`;
    }

    showLoading(containerId) {
        const container = document.getElementById(containerId);
        container.innerHTML = `
            <div class="loading">
                <div class="spinner-border" role="status">
                    <span class="visually-hidden">Loading...</span>
                </div>
            </div>
        `;
    }

    showSuccess(message) {
        this.showToast(message, 'success');
    }

    showError(message) {
        this.showToast(message, 'danger');
    }

    showToast(message, type = 'info') {
        // Create toast container if it doesn't exist
        let toastContainer = document.getElementById('toastContainer');
        if (!toastContainer) {
            toastContainer = document.createElement('div');
            toastContainer.id = 'toastContainer';
            toastContainer.className = 'toast-container position-fixed top-0 end-0 p-3';
            toastContainer.style.zIndex = '9999';
            document.body.appendChild(toastContainer);
        }

        // Create toast
        const toastId = 'toast-' + Date.now();
        const toastHtml = `
            <div id="${toastId}" class="toast align-items-center text-bg-${type} border-0" role="alert">
                <div class="d-flex">
                    <div class="toast-body">${message}</div>
                    <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
                </div>
            </div>
        `;

        toastContainer.insertAdjacentHTML('beforeend', toastHtml);
        
        const toastElement = document.getElementById(toastId);
        const toast = new bootstrap.Toast(toastElement, { delay: 5000 });
        toast.show();

        // Remove toast element after it's hidden
        toastElement.addEventListener('hidden.bs.toast', () => {
            toastElement.remove();
        });
    }

    startAutoRefresh() {
        // Refresh every 30 seconds
        this.refreshInterval = setInterval(() => {
            if (this.currentTab === 'dashboard') {
                this.loadDashboardData();
            }
        }, 30000);
    }

    stopAutoRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    }
}

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.app = new NetTestLabApp();
});

// Global functions for onclick handlers
function showTab(tabName) {
    window.app.showTab(tabName);
}

function refreshInterfaces() {
    window.app.refreshInterfaces();
}

function refreshClients() {
    window.app.refreshClients();
}

function showCreateProfileModal() {
    window.app.showCreateProfileModal();
}

function saveProfile() {
    window.app.saveProfile();
}