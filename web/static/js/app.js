/**
 * NetTestLab Web Application using Connect protocol
 */

// Import the Connect client
import { NetTestLabConnectClient } from './connect-client.js';

class NetTestLabApp {
    constructor() {
        this.connectClient = null;
        this.currentCaptureId = null;
        this.isCapturing = false;
        // Don't call init here - it's called from the DOMContentLoaded listener
    }

    checkAndRedirectFromGenericHost() {
        const currentHost = window.location.hostname;
        const currentPort = window.location.port;
        
        // Check if we're accessing via 0.0.0.0
        if (currentHost === '0.0.0.0') {
            console.log('Detected access via 0.0.0.0, attempting to redirect to local IP...');
            this.redirectToLocalIP(currentPort);
        }
    }

    async redirectToLocalIP(port) {
        try {
            // Try to get the local IP from the API first
            const systemStatus = await this.connectClient.getSystemStatus();
            let localIP = null;

            // Extract IP from interfaces
            if (systemStatus.interfaces && systemStatus.interfaces.length > 0) {
                for (const iface of systemStatus.interfaces) {
                    if (iface.ipAddresses && iface.ipAddresses.length > 0 && iface.isUp) {
                        // Prefer non-loopback interfaces
                        if (iface.type !== 'INTERFACE_TYPE_LOOPBACK') {
                            localIP = iface.ipAddresses[0];
                            break;
                        }
                    }
                }
            }

            if (localIP) {
                const newUrl = `http://${localIP}:${port}${window.location.pathname}${window.location.search}${window.location.hash}`;
                console.log(`Redirecting to: ${newUrl}`);
                window.location.replace(newUrl);
                return;
            }
        } catch (error) {
            console.warn('Could not get system status for IP detection:', error);
        }

        // Fallback: try common local IP ranges
        this.tryCommonLocalIPs(port);
    }

    async tryCommonLocalIPs(port) {
        const commonIPs = [
            '192.168.1.1',
            '192.168.0.1',
            '10.0.0.1',
            '172.16.0.1'
        ];

        for (const ip of commonIPs) {
            try {
                const testUrl = `http://${ip}:${port}/nettestlab.v1.MonitoringService/GetHealth`;
                const response = await fetch(testUrl, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({}),
                    timeout: 3000
                });

                if (response.ok) {
                    const newUrl = `http://${ip}:${port}${window.location.pathname}${window.location.search}${window.location.hash}`;
                    console.log(`Found working IP, redirecting to: ${newUrl}`);
                    window.location.replace(newUrl);
                    return;
                }
            } catch (error) {
                console.log(`IP ${ip} not reachable:`, error.message);
            }
        }

        console.warn('Could not find a working local IP. Staying on 0.0.0.0');
        this.showNotification('Using 0.0.0.0 - could not detect local IP automatically', 'warning');
    }

    async init() {
        console.log('🚀 Initializing NetTestLab Web Interface...');
        
        // Initialize the Connect client
        this.connectClient = new NetTestLabConnectClient();
        
        // Set up event listeners
        this.setupEventListeners();
        
        // Load initial dashboard data
        await this.loadDashboardData();
        
        console.log('✅ NetTestLab Web Interface initialized');
    }

    setupEventListeners() {
        // Global functions for buttons - ensure they're on window object
        window.showTab = this.showTab.bind(this);
        window.refreshStatus = this.refreshStatus.bind(this);
        window.loadInterfaces = this.loadInterfaces.bind(this);
        window.loadProfiles = this.loadProfiles.bind(this);
        window.loadMetrics = this.loadMetrics.bind(this);
        window.createProfile = this.createProfile.bind(this);
        window.editProfile = this.editProfile.bind(this);
        window.deleteProfile = this.deleteProfile.bind(this);
        window.applyProfileToInterface = this.applyProfileToInterface.bind(this);
        window.showProfileModal = this.showProfileModal.bind(this);
        window.saveProfile = this.saveProfile.bind(this);
        window.resetInterfaceConditions = this.resetInterfaceConditions.bind(this);
        
        // Traffic capture functions
        window.startCapture = this.startCapture.bind(this);
        window.stopCapture = this.stopCapture.bind(this);
        window.createTarget = this.createTarget.bind(this);
        window.createDevice = this.createDevice.bind(this);
        window.editTarget = this.editTarget.bind(this);
        window.editDevice = this.editDevice.bind(this);
        window.updateTarget = this.updateTarget.bind(this);
        window.updateDevice = this.updateDevice.bind(this);
        window.deleteTarget = this.deleteTarget.bind(this);
        window.deleteDevice = this.deleteDevice.bind(this);
        window.registerDiscoveredDevice = this.registerDiscoveredDevice.bind(this);
        window.loadDevicesAndTargets = this.loadDevicesAndTargets.bind(this);
        window.loadDevices = this.loadDevices.bind(this);
        window.loadTargets = this.loadTargets.bind(this);
        window.showCreateTargetModal = this.showCreateTargetModal.bind(this);
        window.showCreateDeviceModal = this.showCreateDeviceModal.bind(this);
        window.showEditTargetModal = this.showEditTargetModal.bind(this);
        window.showEditDeviceModal = this.showEditDeviceModal.bind(this);
        window.updateSelectedCount = this.updateSelectedCount.bind(this);
        
        console.log('✅ Global functions assigned to window object');
    }

    // Tab Management
    showTab(tabName) {
        console.log(`Switching to tab: ${tabName}`);
        
        // Hide all tab content divs
        const allTabs = ['dashboard-tab', 'interfaces-tab', 'profiles-tab', 'traffic-capture-tab'];
        allTabs.forEach(tabId => {
            const tab = document.getElementById(tabId);
            if (tab) {
                tab.style.display = 'none';
            }
        });

        // Remove active class from all nav links
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
        });

        // Show selected tab
        const targetTab = document.getElementById(`${tabName}-tab`);
        if (targetTab) {
            targetTab.style.display = 'block';
            console.log(`Showing tab: ${tabName}-tab`);
        } else {
            console.error(`Tab not found: ${tabName}-tab`);
        }

        // Add active class to selected nav link
        const activeLink = document.querySelector(`[onclick="showTab('${tabName}')"]`);
        if (activeLink) {
            activeLink.classList.add('active');
        }

        this.currentTab = tabName;

        // Auto-load data when switching tabs
        if (tabName === 'interfaces') {
            this.loadInterfaces();
        } else if (tabName === 'profiles') {
            this.loadProfiles();
        } else if (tabName === 'traffic-capture') {
            this.loadDevices();
            this.loadTargets();
        }
    }

    async refreshStatus() {
        await this.loadDashboardData();
        await this.loadMetrics();
        this.showNotification('Status refreshed', 'success');
    }

    async loadMetrics() {
        try {
            const metrics = await this.connectClient.getMetrics();
            
            if (metrics && metrics.system) {
                // Update CPU metrics
                const cpuUsage = parseFloat(metrics.system.cpuUsage);
                document.getElementById('cpuProgressBar').style.width = `${cpuUsage}%`;
                document.getElementById('cpuPercentage').textContent = `${cpuUsage.toFixed(1)}%`;
                
                // Update memory metrics
                const memoryUsage = parseFloat(metrics.system.memoryUsage);
                document.getElementById('memoryProgressBar').style.width = `${memoryUsage}%`;
                document.getElementById('memoryPercentage').textContent = `${memoryUsage.toFixed(1)}%`;
                
                // Update disk metrics
                const diskUsage = parseFloat(metrics.system.diskUsage);
                document.getElementById('diskProgressBar').style.width = `${diskUsage}%`;
                document.getElementById('diskPercentage').textContent = `${diskUsage.toFixed(1)}%`;
                
                // Update load average
                if (metrics.system.loadAverage) {
                    const loadAvg = parseFloat(metrics.system.loadAverage.oneMinute);
                    document.getElementById('loadAverage').textContent = loadAvg.toFixed(2);
                }
                
                // Update network connections
                if (metrics.system.networkConnections) {
                    document.getElementById('networkConnections').textContent = metrics.system.networkConnections;
                }
            }
            
            // Update interface stats
            if (metrics && metrics.interfaces) {
                const activeInterfaces = metrics.interfaces.length;
                document.getElementById('activeInterfaces').textContent = activeInterfaces;
                
                // Calculate average utilization
                const avgUtil = metrics.interfaces.reduce((sum, iface) => {
                    return sum + (parseFloat(iface.bandwidth?.utilizationPercent) || 0);
                }, 0) / activeInterfaces;
                
                document.getElementById('avgUtilization').textContent = `${avgUtil.toFixed(1)}%`;
            }
            
            // Update NetTestLab stats
            if (metrics && metrics.nettestlab) {
                document.getElementById('activeConditions').textContent = metrics.nettestlab.activeConditions || 0;
                document.getElementById('profileApplications').textContent = metrics.nettestlab.profileApplications || 0;
                
                // Calculate success rate
                const totalRequests = parseInt(metrics.nettestlab.totalRequests) || 0;
                const failedRequests = parseInt(metrics.nettestlab.failedRequests) || 0;
                const successRate = totalRequests > 0 ? ((totalRequests - failedRequests) / totalRequests * 100) : 100;
                document.getElementById('successRate').textContent = `${successRate.toFixed(1)}%`;
            }
            
            console.log('✅ Metrics loaded successfully');
        } catch (error) {
            console.error('Failed to load metrics:', error);
            // Set default values on error
            const elementsToReset = [
                'cpuPercentage', 'memoryPercentage', 'diskPercentage', 'loadAverage',
                'activeInterfaces', 'networkConnections', 'avgUtilization',
                'activeConditions', 'profileApplications', 'successRate'
            ];
            
            elementsToReset.forEach(id => {
                const element = document.getElementById(id);
                if (element) element.textContent = '0';
            });
        }
    }

    // gRPC Functions using the gRPC client and ES modules
    async loadInterfaces() {
        try {
            this.showLoadingInElement('interfacesList', 'Loading interfaces via gRPC...');
            
            const systemStatus = await this.connectClient.getSystemStatus();
            const profiles = await this.connectClient.listProfiles();
            
            let html = '<div class="row">';
            
            if (systemStatus.interfaces && systemStatus.interfaces.length > 0) {
                // Sort interfaces with the new logic
                const sortedInterfaces = this.sortInterfaces([...systemStatus.interfaces]);
                
                sortedInterfaces.forEach(iface => {
                    const interfaceTypeText = this.getInterfaceTypeText(iface.type);
                    const interfaceIcon = this.getInterfaceIcon(iface.type);
                    const statusBadge = iface.isUp ? 
                        '<span class="badge bg-success">UP</span>' : 
                        '<span class="badge bg-danger">DOWN</span>';
                    
                    // Show applied profile or no conditions
                    let conditionsBadge;
                    if (iface.appliedProfile && iface.appliedProfile.trim() !== '') {
                        conditionsBadge = `<span class="badge bg-success">Profile: ${iface.appliedProfile}</span>`;
                    } else if (iface.hasConditions) {
                        conditionsBadge = '<span class="badge bg-warning">Custom Conditions</span>';
                    } else {
                        conditionsBadge = '<span class="badge bg-secondary">No Conditions</span>';
                    }
                    
                    html += `
                        <div class="col-md-6 mb-3">
                            <div class="card">
                                <div class="card-header d-flex justify-content-between align-items-center">
                                    <h6 class="mb-0">
                                        <i class="bi ${interfaceIcon} me-2"></i>${iface.name}
                                    </h6>
                                    ${statusBadge}
                                </div>
                                <div class="card-body">
                                    <p class="card-text">
                                        <small class="text-muted">Type:</small> ${interfaceTypeText}<br>
                                        <small class="text-muted">IPs:</small> ${iface.ipAddresses ? iface.ipAddresses.join(', ') : 'None'}<br>
                                        ${conditionsBadge}
                                    </p>
                                    <div class="mt-3">
                                        <select class="form-select form-select-sm mb-2" id="profile-select-${iface.name}">
                                            <option value="">Select a profile to apply...</option>
                                            ${profiles.profiles ? profiles.profiles.map(p => 
                                                `<option value="${p.name}">${p.displayName || p.name}</option>`
                                            ).join('') : ''}
                                        </select>
                                        <button class="btn btn-sm btn-primary" onclick="applyProfileToInterface('${iface.name}')">
                                            <i class="bi bi-play-circle"></i> Apply Profile
                                        </button>
                                        ${iface.hasConditions ? `
                                            <button class="btn btn-sm btn-outline-danger ms-1" onclick="resetInterfaceConditions('${iface.name}')">
                                                <i class="bi bi-x-circle"></i> Reset
                                            </button>
                                        ` : ''}
                                    </div>
                                </div>
                            </div>
                        </div>
                    `;
                });
            } else {
                html += '<div class="col-12"><div class="alert alert-info">No interfaces returned from gRPC call.</div></div>';
            }
            
            html += '</div>';
            document.getElementById('interfacesList').innerHTML = html;
            
        } catch (error) {
            console.error('Failed to load interfaces:', error);
            document.getElementById('interfacesList').innerHTML = `
                <div class="alert alert-danger">
                    <i class="bi bi-exclamation-triangle"></i>
                    <strong>gRPC Error:</strong> ${error.message}
                    <br><small>This is expected if gRPC-Web proxy is not configured.</small>
                </div>
            `;
        }
    }

    async loadProfiles() {
        try {
            this.showLoadingInElement('profilesList', 'Loading profiles via gRPC...');
            
            const profiles = await this.connectClient.listProfiles();
            
            let html = `
                <div class="d-flex justify-content-between align-items-center mb-3">
                    <button class="btn btn-primary" onclick="showProfileModal()">
                        <i class="bi bi-plus-circle"></i> Create Profile
                    </button>
                </div>
            `;
            html += '<div class="row">';
            
            if (profiles.profiles && profiles.profiles.length > 0) {
                profiles.profiles.forEach(profile => {
                    const profileTypeText = this.getProfileTypeText(profile.type);
                    const profileIcon = this.getProfileIcon(profile.type);
                    const isBuiltIn = profile.builtIn;
                    const tags = profile.tags ? profile.tags.join(', ') : 'No tags';
                    const conditions = this.formatNetworkConditions(profile.conditions);
                    
                    html += `
                        <div class="col-md-6 mb-3">
                            <div class="card">
                                <div class="card-header d-flex justify-content-between align-items-center">
                                    <h6 class="mb-0">
                                        <i class="bi ${profileIcon} me-2"></i>${profile.displayName || profile.name}
                                    </h6>
                                    <div>
                                        ${isBuiltIn ? '<span class="badge bg-info">Built-in</span>' : '<span class="badge bg-secondary">Custom</span>'}
                                        <span class="badge bg-light text-dark ms-1">${profileTypeText}</span>
                                    </div>
                                </div>
                                <div class="card-body">
                                    <p class="card-text">
                                        ${profile.description || 'No description'}
                                    </p>
                                    <div class="small text-muted mb-2">
                                        <strong>Conditions:</strong><br>
                                        ${conditions}
                                    </div>
                                    <div class="small text-muted mb-3">
                                        <strong>Tags:</strong> ${tags}
                                    </div>
                                    <div class="btn-group w-100" role="group">
                                        <button class="btn btn-sm btn-outline-primary" onclick="editProfile('${profile.name}')">
                                            <i class="bi bi-pencil"></i> ${isBuiltIn ? 'View' : 'Edit'}
                                        </button>
                                        ${!isBuiltIn ? `
                                            <button class="btn btn-sm btn-outline-danger" onclick="deleteProfile('${profile.name}')">
                                                <i class="bi bi-trash"></i> Delete
                                            </button>
                                        ` : ''}
                                    </div>
                                </div>
                            </div>
                        </div>
                    `;
                });
            } else {
                html += '<div class="col-12"><div class="alert alert-info">No profiles returned from gRPC call.</div></div>';
            }
            
            html += '</div>';
            document.getElementById('profilesList').innerHTML = html;
            
        } catch (error) {
            console.error('Failed to load profiles:', error);
            document.getElementById('profilesList').innerHTML = `
                <div class="alert alert-danger">
                    <i class="bi bi-exclamation-triangle"></i>
                    <strong>gRPC Error:</strong> ${error.message}
                    <br><small>This is expected if gRPC-Web proxy is not configured.</small>
                </div>
            `;
        }
    }

    // Dashboard Data Loading
    async loadDashboardData() {
        try {
            this.updateConnectionStatus('connecting');
            
            // Test gRPC services directly
            try {
                const health = await this.connectClient.getHealth();
                console.log('✅ Health service:', health);
                
                // Update UI with real health data
                document.getElementById('serviceName').textContent = 'NetTestLab';
                
                // Convert gRPC health status to human-readable format
                const healthStatus = this.formatHealthStatus(health.status);
                document.getElementById('serviceStatus').textContent = healthStatus;
                document.getElementById('serviceStatus').className = `fw-bold ${this.getHealthStatusClass(health.status)}`;
                
                // Get system status for device information
                try {
                    const systemStatus = await this.connectClient.getSystemStatus();
                    
                    // Add device IP and model information
                    const deviceInfo = this.extractDeviceInfo(systemStatus);
                    this.updateDeviceInfo(deviceInfo);
                    
                } catch (systemError) {
                    console.log('⚠️ System status not available:', systemError);
                }
                
            } catch (healthError) {
                console.log('⚠️ Health service not available:', healthError);
                document.getElementById('serviceStatus').textContent = 'Unavailable';
                document.getElementById('serviceStatus').className = 'fw-bold text-warning';
            }
            
            // Set version from app
            document.getElementById('serviceVersion').textContent = '1.0.0';
            
            // Update connection status to show Connect client works
            this.updateConnectionStatus('connected');
            
            // Load metrics after successful connection
            await this.loadMetrics();
            
            this.showNotification('Services loaded successfully!', 'success');
            
        } catch (error) {
            console.error('Failed to load dashboard data:', error);
            this.updateConnectionStatus('disconnected');
            this.showNotification('Failed to connect to services', 'error');
        }
    }

    // UI Helper Functions
    updateConnectionStatus(status) {
        const statusElement = document.getElementById('connectionStatus');
        if (!statusElement) return;

        statusElement.className = 'badge';
        
        switch (status) {
            case 'connected':
                statusElement.classList.add('bg-success');
                statusElement.innerHTML = '<i class="bi bi-wifi"></i> Connected';
                break;
            case 'connecting':
                statusElement.classList.add('bg-warning');
                statusElement.innerHTML = '<i class="bi bi-hourglass-split"></i> Connecting';
                break;
            case 'warning':
                statusElement.classList.add('bg-warning');
                statusElement.innerHTML = '<i class="bi bi-exclamation-triangle"></i> Partial';
                break;
            case 'disconnected':
                statusElement.classList.add('bg-danger');
                statusElement.innerHTML = '<i class="bi bi-wifi-off"></i> Disconnected';
                break;
        }
    }

    showLoadingInElement(elementId, message = 'Loading...') {
        document.getElementById(elementId).innerHTML = `
            <div class="d-flex justify-content-center align-items-center p-4">
                <div class="spinner-border spinner-border-sm me-2" role="status">
                    <span class="visually-hidden">Loading...</span>
                </div>
                ${message}
            </div>
        `;
    }

    showResults(title, data) {
        document.getElementById('resultsModalTitle').textContent = title;
        document.getElementById('resultsContent').textContent = JSON.stringify(data, null, 2);
        
        const modal = new bootstrap.Modal(document.getElementById('resultsModal'));
        modal.show();
    }

    // Profile Management Functions
    async createProfile() {
        this.showProfileModal();
    }

    async editProfile(profileName) {
        try {
            const profile = await this.connectClient.getProfile(profileName);
            this.showProfileModal(profile.profile);
        } catch (error) {
            console.error('Failed to load profile for editing:', error);
            this.showNotification(`Failed to load profile: ${error.message}`, 'error');
        }
    }

    async deleteProfile(profileName) {
        if (!confirm(`Are you sure you want to delete profile "${profileName}"?`)) {
            return;
        }

        try {
            const result = await this.connectClient.deleteProfile(profileName);
            if (result.success) {
                this.showNotification(`Profile "${profileName}" deleted successfully`, 'success');
                await this.loadProfiles(); // Reload the list
            } else {
                this.showNotification(`Failed to delete profile: ${result.errorMessage}`, 'error');
            }
        } catch (error) {
            console.error('Failed to delete profile:', error);
            this.showNotification(`Failed to delete profile: ${error.message}`, 'error');
        }
    }

    async applyProfileToInterface(interfaceName) {
        const profileSelect = document.getElementById(`profile-select-${interfaceName}`);
        const profileName = profileSelect.value;

        if (!profileName) {
            this.showNotification('Please select a profile first', 'warning');
            return;
        }

        try {
            const result = await this.connectClient.applyProfile(profileName, interfaceName);
            if (result.success) {
                this.showNotification(`Profile "${profileName}" applied to ${interfaceName}`, 'success');
                await this.loadInterfaces(); // Reload to show updated status
            } else {
                this.showNotification(`Failed to apply profile: ${result.errorMessage}`, 'error');
            }
        } catch (error) {
            console.error('Failed to apply profile:', error);
            this.showNotification(`Failed to apply profile: ${error.message}`, 'error');
        }
    }

    showProfileModal(profile = null) {
        const isEdit = !!profile;
        const modalTitle = isEdit ? (profile.builtIn ? 'View Profile' : 'Edit Profile') : 'Create New Profile';
        const readonly = isEdit && profile.builtIn;

        const profileTypeOptions = [
            { value: 'PROFILE_TYPE_MOBILE', label: 'Mobile' },
            { value: 'PROFILE_TYPE_WIFI', label: 'WiFi' },
            { value: 'PROFILE_TYPE_SATELLITE', label: 'Satellite' },
            { value: 'PROFILE_TYPE_CUSTOM', label: 'Custom' },
            { value: 'PROFILE_TYPE_TESTING', label: 'Testing' }
        ];

        const modalHtml = `
            <div class="modal fade" id="profileModal" tabindex="-1">
                <div class="modal-dialog modal-lg">
                    <div class="modal-content">
                        <div class="modal-header">
                            <h5 class="modal-title">${modalTitle}</h5>
                            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                        </div>
                        <div class="modal-body">
                            <form id="profileForm">
                                <div class="row">
                                    <div class="col-md-6">
                                        <div class="mb-3">
                                            <label class="form-label">
                                                Profile Name
                                                <button type="button" class="btn btn-link btn-sm p-0 ms-1" data-bs-toggle="tooltip" 
                                                        title="Unique identifier for the profile. Used internally by the system.">
                                                    <i class="bi bi-question-circle"></i>
                                                </button>
                                            </label>
                                            <input type="text" class="form-control" id="profileName" 
                                                value="${profile?.name || ''}" ${readonly || isEdit ? 'readonly' : ''} required>
                                        </div>
                                    </div>
                                    <div class="col-md-6">
                                        <div class="mb-3">
                                            <label class="form-label">
                                                Display Name
                                                <button type="button" class="btn btn-link btn-sm p-0 ms-1" data-bs-toggle="tooltip" 
                                                        title="Human-readable name shown in the interface. Can be changed anytime.">
                                                    <i class="bi bi-question-circle"></i>
                                                </button>
                                            </label>
                                            <input type="text" class="form-control" id="profileDisplayName" 
                                                value="${profile?.displayName || ''}" ${readonly ? 'readonly' : ''}>
                                        </div>
                                    </div>
                                </div>
                                
                                <div class="mb-3">
                                    <label class="form-label">
                                        Description
                                        <button type="button" class="btn btn-link btn-sm p-0 ms-1" data-bs-toggle="tooltip" 
                                                title="Brief description of what this profile simulates or its intended use case.">
                                            <i class="bi bi-question-circle"></i>
                                        </button>
                                    </label>
                                    <textarea class="form-control" id="profileDescription" rows="2" ${readonly ? 'readonly' : ''}>${profile?.description || ''}</textarea>
                                </div>
                                
                                <div class="row">
                                    <div class="col-md-6">
                                        <div class="mb-3">
                                            <label class="form-label">
                                                Profile Type
                                                <button type="button" class="btn btn-link btn-sm p-0 ms-1" data-bs-toggle="tooltip" 
                                                        title="Category of network condition this profile represents (Mobile, WiFi, Satellite, etc.).">
                                                    <i class="bi bi-question-circle"></i>
                                                </button>
                                            </label>
                                            <select class="form-select" id="profileType" ${readonly ? 'disabled' : ''}>
                                                ${profileTypeOptions.map(opt => 
                                                    `<option value="${opt.value}" ${profile?.type === opt.value ? 'selected' : ''}>${opt.label}</option>`
                                                ).join('')}
                                            </select>
                                        </div>
                                    </div>
                                    <div class="col-md-6">
                                        <div class="mb-3">
                                            <label class="form-label">
                                                Tags (comma-separated)
                                                <button type="button" class="btn btn-link btn-sm p-0 ms-1" data-bs-toggle="tooltip" 
                                                        title="Keywords for organizing and searching profiles. Example: mobile, slow, testing">
                                                    <i class="bi bi-question-circle"></i>
                                                </button>
                                            </label>
                                            <input type="text" class="form-control" id="profileTags" 
                                                value="${profile?.tags ? profile.tags.join(', ') : ''}" ${readonly ? 'readonly' : ''}>
                                        </div>
                                    </div>
                                </div>

                                <h6 class="border-bottom pb-2">Network Conditions</h6>
                                
                                <!-- Latency -->
                                <div class="card mb-3">
                                    <div class="card-header">
                                        <div class="form-check">
                                            <input class="form-check-input" type="checkbox" id="latencyEnabled" 
                                                ${profile?.conditions?.latency?.enabled ? 'checked' : ''} ${readonly ? 'disabled' : ''}>
                                            <label class="form-check-label">
                                                Latency
                                                <button type="button" class="btn btn-link btn-sm p-0 ms-1" data-bs-toggle="tooltip" 
                                                        title="Network delay - time it takes for data to travel from source to destination. Higher values simulate slower networks.">
                                                    <i class="bi bi-question-circle"></i>
                                                </button>
                                            </label>
                                        </div>
                                    </div>
                                    <div class="card-body">
                                        <div class="mb-3">
                                            <label class="form-label">Delay (ms)</label>
                                            <input type="number" class="form-control" id="latencyDelay" 
                                                value="${profile?.conditions?.latency?.delayMs || 0}" ${readonly ? 'readonly' : ''}>
                                        </div>
                                    </div>
                                </div>

                                <!-- Packet Loss -->
                                <div class="card mb-3">
                                    <div class="card-header">
                                        <div class="form-check">
                                            <input class="form-check-input" type="checkbox" id="packetLossEnabled" 
                                                ${profile?.conditions?.packetLoss?.enabled ? 'checked' : ''} ${readonly ? 'disabled' : ''}>
                                            <label class="form-check-label">
                                                Packet Loss
                                                <button type="button" class="btn btn-link btn-sm p-0 ms-1" data-bs-toggle="tooltip" 
                                                        title="Percentage of data packets that get lost during transmission. Common in wireless and congested networks.">
                                                    <i class="bi bi-question-circle"></i>
                                                </button>
                                            </label>
                                        </div>
                                    </div>
                                    <div class="card-body">
                                        <div class="mb-3">
                                            <label class="form-label">Percentage (%)</label>
                                            <input type="number" class="form-control" id="packetLossPercentage" 
                                                step="0.1" min="0" max="100" value="${profile?.conditions?.packetLoss?.percentage || 0}" ${readonly ? 'readonly' : ''}>
                                        </div>
                                    </div>
                                </div>

                                <!-- Bandwidth -->
                                <div class="card mb-3">
                                    <div class="card-header">
                                        <div class="form-check">
                                            <input class="form-check-input" type="checkbox" id="bandwidthEnabled" 
                                                ${profile?.conditions?.bandwidth?.enabled ? 'checked' : ''} ${readonly ? 'disabled' : ''}>
                                            <label class="form-check-label">
                                                Bandwidth Limiting
                                                <button type="button" class="btn btn-link btn-sm p-0 ms-1" data-bs-toggle="tooltip" 
                                                        title="Restricts the maximum data transfer rate. Simulates slower connections like 3G, satellite internet, etc.">
                                                    <i class="bi bi-question-circle"></i>
                                                </button>
                                            </label>
                                        </div>
                                    </div>
                                    <div class="card-body">
                                        <div class="row">
                                            <div class="col-md-6">
                                                <div class="mb-3">
                                                    <label class="form-label">Download (bps)</label>
                                                    <input type="number" class="form-control" id="bandwidthDownload" 
                                                        value="${profile?.conditions?.bandwidth?.downloadBps || 0}" ${readonly ? 'readonly' : ''}>
                                                </div>
                                            </div>
                                            <div class="col-md-6">
                                                <div class="mb-3">
                                                    <label class="form-label">Upload (bps)</label>
                                                    <input type="number" class="form-control" id="bandwidthUpload" 
                                                        value="${profile?.conditions?.bandwidth?.uploadBps || 0}" ${readonly ? 'readonly' : ''}>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <!-- Jitter -->
                                <div class="card mb-3">
                                    <div class="card-header">
                                        <div class="form-check">
                                            <input class="form-check-input" type="checkbox" id="jitterEnabled" 
                                                ${profile?.conditions?.jitter?.enabled ? 'checked' : ''} ${readonly ? 'disabled' : ''}>
                                            <label class="form-check-label">
                                                Jitter
                                                <button type="button" class="btn btn-link btn-sm p-0 ms-1" data-bs-toggle="tooltip" 
                                                        title="Variation in latency over time. Makes delay inconsistent, simulating unstable network conditions.">
                                                    <i class="bi bi-question-circle"></i>
                                                </button>
                                            </label>
                                        </div>
                                    </div>
                                    <div class="card-body">
                                        <div class="mb-3">
                                            <label class="form-label">Variation (ms)</label>
                                            <input type="number" class="form-control" id="jitterVariation" 
                                                value="${profile?.conditions?.jitter?.variationMs || 0}" ${readonly ? 'readonly' : ''}>
                                        </div>
                                    </div>
                                </div>

                                <!-- Corruption -->
                                <div class="card mb-3">
                                    <div class="card-header">
                                        <div class="form-check">
                                            <input class="form-check-input" type="checkbox" id="corruptionEnabled" 
                                                ${profile?.conditions?.corruption?.enabled ? 'checked' : ''} ${readonly ? 'disabled' : ''}>
                                            <label class="form-check-label">
                                                Packet Corruption
                                                <button type="button" class="btn btn-link btn-sm p-0 ms-1" data-bs-toggle="tooltip" 
                                                        title="Percentage of packets that get corrupted during transmission. Simulates poor quality connections.">
                                                    <i class="bi bi-question-circle"></i>
                                                </button>
                                            </label>
                                        </div>
                                    </div>
                                    <div class="card-body">
                                        <div class="mb-3">
                                            <label class="form-label">Percentage (%)</label>
                                            <input type="number" class="form-control" id="corruptionPercentage" 
                                                step="0.1" min="0" max="100" value="${profile?.conditions?.corruption?.percentage || 0}" ${readonly ? 'readonly' : ''}>
                                        </div>
                                    </div>
                                </div>
                            </form>
                        </div>
                        <div class="modal-footer">
                            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
                            ${!readonly ? `<button type="button" class="btn btn-primary" onclick="saveProfile(${isEdit})">
                                ${isEdit ? 'Update' : 'Create'} Profile
                            </button>` : ''}
                        </div>
                    </div>
                </div>
            </div>
        `;

        // Remove existing modal if any
        const existingModal = document.getElementById('profileModal');
        if (existingModal) {
            existingModal.remove();
        }

        // Add new modal to body
        document.body.insertAdjacentHTML('beforeend', modalHtml);

        // Show modal
        const modal = new bootstrap.Modal(document.getElementById('profileModal'));
        modal.show();

        // Initialize tooltips
        const tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
        tooltipTriggerList.map(function (tooltipTriggerEl) {
            return new bootstrap.Tooltip(tooltipTriggerEl);
        });
    }

    async saveProfile(isEdit = false) {
        const form = document.getElementById('profileForm');
        const formData = new FormData(form);

        try {
            // Collect form data
            const profileData = {
                name: document.getElementById('profileName').value,
                displayName: document.getElementById('profileDisplayName').value,
                description: document.getElementById('profileDescription').value,
                type: document.getElementById('profileType').value,
                tags: document.getElementById('profileTags').value.split(',').map(t => t.trim()).filter(t => t),
                conditions: {
                    latency: {
                        enabled: document.getElementById('latencyEnabled').checked,
                        delayMs: parseInt(document.getElementById('latencyDelay').value) || 0
                    },
                    packetLoss: {
                        enabled: document.getElementById('packetLossEnabled').checked,
                        percentage: parseFloat(document.getElementById('packetLossPercentage').value) || 0
                    },
                    bandwidth: {
                        enabled: document.getElementById('bandwidthEnabled').checked,
                        downloadBps: parseInt(document.getElementById('bandwidthDownload').value) || 0,
                        uploadBps: parseInt(document.getElementById('bandwidthUpload').value) || 0
                    },
                    jitter: {
                        enabled: document.getElementById('jitterEnabled').checked,
                        variationMs: parseInt(document.getElementById('jitterVariation').value) || 0
                    },
                    corruption: {
                        enabled: document.getElementById('corruptionEnabled').checked,
                        percentage: parseFloat(document.getElementById('corruptionPercentage').value) || 0
                    }
                }
            };

            let result;
            if (isEdit) {
                result = await this.connectClient.updateProfile(profileData.name, profileData);
            } else {
                result = await this.connectClient.createProfile(profileData);
            }

            if (result.success) {
                this.showNotification(`Profile ${isEdit ? 'updated' : 'created'} successfully`, 'success');
                
                // Close modal
                const modal = bootstrap.Modal.getInstance(document.getElementById('profileModal'));
                modal.hide();
                
                // Reload profiles
                await this.loadProfiles();
            } else {
                this.showNotification(`Failed to ${isEdit ? 'update' : 'create'} profile: ${result.errorMessage}`, 'error');
            }

        } catch (error) {
            console.error(`Failed to ${isEdit ? 'update' : 'create'} profile:`, error);
            this.showNotification(`Failed to ${isEdit ? 'update' : 'create'} profile: ${error.message}`, 'error');
        }
    }

    async resetInterfaceConditions(interfaceName) {
        if (!confirm(`Are you sure you want to reset network conditions on ${interfaceName}?`)) {
            return;
        }

        try {
            const result = await this.connectClient.resetNetworkConditions(interfaceName);
            if (result.success) {
                this.showNotification(`Network conditions reset on ${interfaceName}`, 'success');
                await this.loadInterfaces(); // Reload to show updated status
            } else {
                this.showNotification(`Failed to reset conditions: ${result.errorMessage}`, 'error');
            }
        } catch (error) {
            console.error('Failed to reset network conditions:', error);
            this.showNotification(`Failed to reset conditions: ${error.message}`, 'error');
        }
    }

    // Helper Functions
    getInterfaceTypeText(type) {
        const types = {
            'INTERFACE_TYPE_ETHERNET': 'Ethernet',
            'INTERFACE_TYPE_WIRELESS': 'Wireless',
            'INTERFACE_TYPE_LOOPBACK': 'Loopback',
            'INTERFACE_TYPE_BRIDGE': 'Bridge',
            'INTERFACE_TYPE_UNSPECIFIED': 'Unknown'
        };
        return types[type] || 'Unknown';
    }

    getInterfaceIcon(type) {
        const icons = {
            'INTERFACE_TYPE_ETHERNET': 'bi-ethernet',
            'INTERFACE_TYPE_WIRELESS': 'bi-wifi',
            'INTERFACE_TYPE_LOOPBACK': 'bi-arrow-repeat',
            'INTERFACE_TYPE_BRIDGE': 'bi-diagram-3',
            'INTERFACE_TYPE_UNSPECIFIED': 'bi-question-circle'
        };
        return icons[type] || 'bi-question-circle';
    }

    getProfileTypeText(type) {
        const types = {
            'PROFILE_TYPE_MOBILE': 'Mobile',
            'PROFILE_TYPE_WIFI': 'WiFi',
            'PROFILE_TYPE_SATELLITE': 'Satellite',
            'PROFILE_TYPE_CUSTOM': 'Custom',
            'PROFILE_TYPE_TESTING': 'Testing',
            'PROFILE_TYPE_UNSPECIFIED': 'Unknown'
        };
        return types[type] || 'Unknown';
    }

    getProfileIcon(type) {
        const icons = {
            'PROFILE_TYPE_MOBILE': 'bi-phone',
            'PROFILE_TYPE_WIFI': 'bi-wifi',
            'PROFILE_TYPE_SATELLITE': 'bi-globe',
            'PROFILE_TYPE_CUSTOM': 'bi-gear',
            'PROFILE_TYPE_TESTING': 'bi-flask',
            'PROFILE_TYPE_UNSPECIFIED': 'bi-question-circle'
        };
        return icons[type] || 'bi-question-circle';
    }

    formatHealthStatus(status) {
        const statusMap = {
            'HEALTH_STATUS_HEALTHY': 'Healthy',
            'HEALTH_STATUS_UNHEALTHY': 'Unhealthy',
            'HEALTH_STATUS_DEGRADED': 'Degraded',
            'HEALTH_STATUS_UNKNOWN': 'Unknown',
            'HEALTH_STATUS_UNSPECIFIED': 'Unknown'
        };
        return statusMap[status] || status || 'Unknown';
    }

    getHealthStatusClass(status) {
        const classMap = {
            'HEALTH_STATUS_HEALTHY': 'text-success',
            'HEALTH_STATUS_UNHEALTHY': 'text-danger',
            'HEALTH_STATUS_DEGRADED': 'text-warning',
            'HEALTH_STATUS_UNKNOWN': 'text-secondary',
            'HEALTH_STATUS_UNSPECIFIED': 'text-secondary'
        };
        return classMap[status] || 'text-secondary';
    }

    extractDeviceInfo(systemStatus) {
        const deviceInfo = {
            ip: 'Unknown',
            model: 'Unknown'
        };

        // Extract IP from interfaces
        if (systemStatus.interfaces && systemStatus.interfaces.length > 0) {
            for (const iface of systemStatus.interfaces) {
                if (iface.ipAddresses && iface.ipAddresses.length > 0 && iface.isUp) {
                    // Prefer non-loopback interfaces
                    if (iface.type !== 'INTERFACE_TYPE_LOOPBACK') {
                        deviceInfo.ip = iface.ipAddresses[0];
                        break;
                    }
                }
            }
        }

        // Extract model information (if available in system status)
        if (systemStatus.systemInfo) {
            deviceInfo.model = systemStatus.systemInfo.model || 'NetTestLab Device';
        } else {
            deviceInfo.model = 'NetTestLab Device';
        }

        return deviceInfo;
    }

    updateDeviceInfo(deviceInfo) {
        // Check if device info elements exist, if not create them
        let deviceIPElement = document.getElementById('deviceIP');
        let deviceModelElement = document.getElementById('deviceModel');

        if (!deviceIPElement || !deviceModelElement) {
            // Add device info to the system status card
            const systemCard = document.querySelector('.card-body');
            if (systemCard) {
                const deviceInfoHtml = `
                    <hr>
                    <div class="row">
                        <div class="col-6">
                            <small class="text-muted">Device IP</small>
                            <div id="deviceIP" class="fw-bold">${deviceInfo.ip}</div>
                        </div>
                        <div class="col-6">
                            <small class="text-muted">Model</small>
                            <div id="deviceModel" class="fw-bold">${deviceInfo.model}</div>
                        </div>
                    </div>
                `;
                systemCard.insertAdjacentHTML('beforeend', deviceInfoHtml);
            }
        } else {
            deviceIPElement.textContent = deviceInfo.ip;
            deviceModelElement.textContent = deviceInfo.model;
        }
    }

    sortInterfaces(interfaces) {
        return interfaces.sort((a, b) => {
            // First: interfaces with IPs and UP
            const aHasIPsAndUp = (a.ipAddresses && a.ipAddresses.length > 0) && a.isUp;
            const bHasIPsAndUp = (b.ipAddresses && b.ipAddresses.length > 0) && b.isUp;
            
            if (aHasIPsAndUp && !bHasIPsAndUp) return -1;
            if (!aHasIPsAndUp && bHasIPsAndUp) return 1;
            
            // Second: interfaces with IPs (regardless of UP status)
            const aHasIPs = a.ipAddresses && a.ipAddresses.length > 0;
            const bHasIPs = b.ipAddresses && b.ipAddresses.length > 0;
            
            if (aHasIPs && !bHasIPs) return -1;
            if (!aHasIPs && bHasIPs) return 1;
            
            // Third: UP interfaces
            if (a.isUp && !b.isUp) return -1;
            if (!a.isUp && b.isUp) return 1;
            
            // Finally: alphabetical order
            return a.name.localeCompare(b.name);
        });
    }

    formatBandwidth(bps) {
        if (!bps || bps === 0) return '0 bps';
        
        const units = [
            { name: 'Gbps', value: 1000000000 },
            { name: 'Mbps', value: 1000000 },
            { name: 'Kbps', value: 1000 },
            { name: 'bps', value: 1 }
        ];
        
        for (const unit of units) {
            if (bps >= unit.value) {
                const value = (bps / unit.value).toFixed(bps >= unit.value * 10 ? 0 : 1);
                return `${value} ${unit.name}`;
            }
        }
        
        return `${bps} bps`;
    }

    formatNetworkConditions(conditions) {
        if (!conditions) return 'None configured';
        
        const parts = [];
        if (conditions.latency?.enabled) parts.push(`Latency: ${conditions.latency.delayMs}ms`);
        if (conditions.packetLoss?.enabled) parts.push(`Loss: ${conditions.packetLoss.percentage}%`);
        if (conditions.bandwidth?.enabled) parts.push(`BW: ${this.formatBandwidth(conditions.bandwidth.downloadBps)}`);
        if (conditions.jitter?.enabled) parts.push(`Jitter: ${conditions.jitter.variationMs}ms`);
        if (conditions.corruption?.enabled) parts.push(`Corruption: ${conditions.corruption.percentage}%`);
        
        return parts.length > 0 ? parts.join(', ') : 'None active';
    }

    showNotification(message, type) {
        // Create a toast notification
        const toast = document.createElement('div');
        toast.className = `alert alert-${type === 'error' ? 'danger' : type} alert-dismissible fade show position-fixed top-0 end-0 m-3`;
        toast.style.zIndex = '9999';
        toast.innerHTML = `
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        `;
        
        document.body.appendChild(toast);
        
        // Auto remove after 5 seconds
        setTimeout(() => {
            if (toast.parentNode) {
                toast.parentNode.removeChild(toast);
            }
        }, 5000);
    }

    // Traffic Capture Functions
    async loadDevicesAndTargets() {
        await this.loadDevices();
        await this.loadTargets();
    }

    async loadDevices() {
        try {
            // Load all devices from the backend (includes connection status)
            const devicesResponse = await this.connectClient.listDevices();
            let allDevices = devicesResponse.devices || [];
            
            // Transform devices to include frontend-specific flags
            const transformedDevices = allDevices.map(device => ({
                ...device,
                isConnected: device.connectionStatus === 'DEVICE_CONNECTION_STATUS_CONNECTED',
                isDeleted: device.isDeleted || false,
                isTemporary: false // All devices from backend are registered
            }));
            
            // Store current devices list for other functions
            this.currentDevicesList = transformedDevices;
            
            this.renderDevicesList(transformedDevices);
            this.updateSelectedCount('devices');
        } catch (error) {
            console.error('Error loading devices:', error);
            this.showNotification('Failed to load devices: ' + error.message, 'error');
            document.getElementById('devicesList').innerHTML = `
                <div class="text-center text-danger p-3">
                    <i class="bi bi-exclamation-triangle"></i><br>
                    Failed to load devices<br>
                    <small>${error.message}</small>
                </div>
            `;
        }
    }

    mergeDevicesLists(registeredDevices, connectedDevices) {
        const deviceMap = new Map();
        
        // Add all registered devices
        registeredDevices.forEach(device => {
            deviceMap.set(device.macAddress, {
                ...device,
                isConnected: false,
                isDeleted: device.isDeleted || false
            });
        });
        
        // Add or update with connected devices
        connectedDevices.forEach(connectedDevice => {
            const existing = deviceMap.get(connectedDevice.macAddress);
            if (existing) {
                // Update existing device with connection status
                existing.isConnected = true;
                existing.ipAddress = connectedDevice.ipAddress; // Update current IP
            } else {
                // Add new connected device (not in database)
                deviceMap.set(connectedDevice.macAddress, {
                    id: `connected-${connectedDevice.macAddress}`,
                    deviceName: connectedDevice.deviceName || `Unknown Device`,
                    ipAddress: connectedDevice.ipAddress,
                    macAddress: connectedDevice.macAddress,
                    interface: connectedDevice.interface || 'unknown',
                    deviceType: 'DEVICE_TYPE_OTHER',
                    isConnected: true,
                    isDeleted: false,
                    isTemporary: true // Flag for devices discovered but not registered
                });
            }
        });
        
        return Array.from(deviceMap.values()).sort((a, b) => {
            // Sort: Connected first, then by name
            if (a.isConnected && !b.isConnected) return -1;
            if (!a.isConnected && b.isConnected) return 1;
            return (a.deviceName || a.macAddress).localeCompare(b.deviceName || b.macAddress);
        });
    }

    async loadTargets() {
        try {
            const targetsResponse = await this.connectClient.listUrlTargets();
            this.renderTargetsList(targetsResponse.targets || []);
            this.updateSelectedCount('targets');
        } catch (error) {
            console.error('Error loading targets:', error);
            this.showNotification('Failed to load targets: ' + error.message, 'error');
            document.getElementById('targetsList').innerHTML = `
                <div class="text-center text-danger p-3">
                    <i class="bi bi-exclamation-triangle"></i><br>
                    Failed to load targets<br>
                    <small>${error.message}</small>
                </div>
            `;
        }
    }

    renderDevicesList(devices) {
        const container = document.getElementById('devicesList');
        if (!devices || devices.length === 0) {
            container.innerHTML = `
                <div class="text-center text-muted p-3">
                    <i class="bi bi-router"></i><br>
                    No devices found<br>
                    <small>Click "Add Device" to register network devices</small>
                </div>
            `;
            return;
        }

        container.innerHTML = devices.map(device => {
            const isConnected = device.isConnected || false;
            const isDeleted = device.isDeleted || false;
            const isTemporary = device.isTemporary || false;
            
            const statusBadge = isConnected ? 
                '<span class="badge bg-success ms-1">Connected</span>' : 
                '<span class="badge bg-secondary ms-1">Offline</span>';
            
            let additionalBadges = '';
            if (isDeleted) additionalBadges += '<span class="badge bg-warning ms-1">Deleted</span>';
            if (isTemporary) additionalBadges += '<span class="badge bg-info ms-1">Discovered</span>';
            
            const canEdit = !isTemporary; // Can't edit temporary discovered devices
            const canDelete = !isDeleted && !isTemporary;
            const canSelect = !isDeleted || isConnected; // Can select if not deleted, or if connected
            
            return `
                <div class="form-check mb-2">
                    <div class="d-flex justify-content-between align-items-start">
                        <div class="d-flex align-items-start flex-grow-1">
                            <input class="form-check-input me-2 mt-1" type="checkbox" id="device-${device.id}" 
                                   value="${device.id}" onchange="updateSelectedCount('devices')" ${!canSelect ? 'disabled' : ''}>
                            <label class="form-check-label flex-grow-1" for="device-${device.id}">
                                <div class="d-flex align-items-center">
                                    <i class="bi bi-${this.getDeviceIcon(device.deviceType)} me-2"></i>
                                    <div>
                                        <strong>${device.deviceName || device.macAddress}</strong>
                                        ${statusBadge}${additionalBadges}
                                        <br>
                                        <small class="text-muted">
                                            ${device.ipAddress || 'undefined'} - ${device.macAddress}
                                            ${device.interface ? ` (${device.interface})` : ''}
                                        </small>
                                        ${isTemporary ? '<br><small class="text-info">Click "+" to register this discovered device</small>' : ''}
                                    </div>
                                </div>
                            </label>
                        </div>
                        <div class="btn-group btn-group-sm ms-2" role="group">
                            ${isTemporary ? `
                                <button class="btn btn-outline-success btn-sm" onclick="registerDiscoveredDevice('${device.macAddress}')" 
                                        title="Register device">
                                    <i class="bi bi-plus"></i>
                                </button>
                            ` : `
                                <button class="btn btn-outline-primary btn-sm" onclick="editDevice('${device.id}')" 
                                        title="Edit device" ${!canEdit ? 'disabled' : ''}>
                                    <i class="bi bi-pencil"></i>
                                </button>
                                <button class="btn btn-outline-danger btn-sm" onclick="deleteDevice('${device.id}')" 
                                        title="Delete device" ${!canDelete ? 'disabled' : ''}>
                                    <i class="bi bi-trash"></i>
                                </button>
                            `}
                        </div>
                    </div>
                </div>
            `;
        }).join('');
    }

    renderTargetsList(targets) {
        const container = document.getElementById('targetsList');
        if (!targets || targets.length === 0) {
            container.innerHTML = `
                <div class="text-center text-muted p-3">
                    <i class="bi bi-bullseye"></i><br>
                    No URL targets found<br>
                    <small>Click "Create Target" to define traffic patterns</small>
                </div>
            `;
            return;
        }

        container.innerHTML = targets.map(target => `
            <div class="form-check mb-2">
                <div class="d-flex justify-content-between align-items-start">
                    <div class="d-flex align-items-start flex-grow-1">
                        <input class="form-check-input me-2 mt-1" type="checkbox" id="target-${target.id}" 
                               value="${target.id}" onchange="updateSelectedCount('targets')" ${target.enabled === false ? 'disabled' : ''}>
                        <label class="form-check-label flex-grow-1" for="target-${target.id}">
                            <div class="d-flex align-items-center">
                                <i class="bi bi-bullseye me-2"></i>
                                <div>
                                    <strong>${target.name}</strong>
                                    ${target.enabled === false ? '<span class="badge bg-secondary ms-1">Disabled</span>' : ''}
                                    <br>
                                    <small class="text-muted">
                                        ${target.hostRegex} - Ports: ${target.ports?.join(', ') || 'any'}
                                    </small>
                                </div>
                            </div>
                        </label>
                    </div>
                    <div class="btn-group btn-group-sm ms-2" role="group">
                        <button class="btn btn-outline-primary btn-sm" onclick="editTarget('${target.id}')" 
                                title="Edit target">
                            <i class="bi bi-pencil"></i>
                        </button>
                        <button class="btn btn-outline-danger btn-sm" onclick="deleteTarget('${target.id}')" 
                                title="Delete target">
                            <i class="bi bi-trash"></i>
                        </button>
                    </div>
                </div>
            </div>
        `).join('');
    }

    async startCapture() {
        try {
            const form = document.getElementById('captureForm');
            const formData = new FormData(form);
            
            // Get selected devices
            const selectedDevices = Array.from(document.querySelectorAll('#devicesList input:checked'))
                .map(input => input.value);
            
            // Get selected targets
            const selectedTargets = Array.from(document.querySelectorAll('#targetsList input:checked'))
                .map(input => input.value);
            
            // Get selected protocols
            const protocolSelect = document.getElementById('protocols');
            const selectedProtocols = Array.from(protocolSelect.selectedOptions)
                .map(option => option.value);

            const captureRequest = {
                capture_name: document.getElementById('captureName').value,
                duration: document.getElementById('captureDuration').value,
                max_size_mb: parseInt(document.getElementById('maxSizeMB').value),
                device_ids: selectedDevices,
                url_target_ids: selectedTargets,
                protocols: selectedProtocols,
                capture_payload: document.getElementById('capturePayload').checked
            };

            console.log('Starting capture with request:', captureRequest);
            const response = await this.connectClient.startCapture(captureRequest);
            
            if (response.success) {
                this.showNotification(`Capture started successfully: ${response.message}`, 'success');
                this.currentCaptureId = response.captureId;
                this.updateCaptureUI(true);
                this.monitorCurrentCapture();
            } else {
                this.showNotification(`Failed to start capture: ${response.message}`, 'error');
            }
            
        } catch (error) {
            console.error('Error starting capture:', error);
            this.showNotification('Failed to start capture: ' + error.message, 'error');
        }
    }

    async stopCapture() {
        if (!this.currentCaptureId) {
            this.showNotification('No active capture to stop', 'warning');
            return;
        }

        try {
            const response = await this.connectClient.stopCapture({ capture_id: this.currentCaptureId });
            this.showNotification('Capture stopped successfully', 'success');
            this.updateCaptureUI(false);
            this.currentCaptureId = null;
            this.loadCaptureHistory();
        } catch (error) {
            console.error('Error stopping capture:', error);
            this.showNotification('Failed to stop capture: ' + error.message, 'error');
        }
    }

    updateCaptureUI(isCapturing) {
        const startBtn = document.getElementById('startCaptureBtn');
        const stopBtn = document.getElementById('stopCaptureBtn');
        const form = document.getElementById('captureForm');
        
        if (isCapturing) {
            startBtn.style.display = 'none';
            stopBtn.style.display = 'block';
            form.querySelectorAll('input, select').forEach(input => input.disabled = true);
        } else {
            startBtn.style.display = 'block';
            stopBtn.style.display = 'none';
            form.querySelectorAll('input, select').forEach(input => input.disabled = false);
        }
    }

    async monitorCurrentCapture() {
        if (!this.currentCaptureId) return;

        try {
            const status = await this.connectClient.getCaptureStatus({ capture_id: this.currentCaptureId });
            const capture = status.capture;
            
            // Update current capture status display
            const statusContainer = document.getElementById('currentCaptureStatus');
            const nameSpan = document.getElementById('currentCaptureName');
            const statusBadge = document.getElementById('currentCaptureStatusBadge');
            const startTimeSpan = document.getElementById('currentCaptureStartTime');
            
            statusContainer.style.display = 'block';
            nameSpan.textContent = capture.name;
            statusBadge.textContent = this.formatCaptureStatus(capture.status);
            statusBadge.className = this.getCaptureStatusClass(capture.status);
            startTimeSpan.textContent = new Date(capture.startedAt).toLocaleString();
            
            // If capture is still running, continue monitoring
            if (capture.status === 'CAPTURE_STATUS_ACTIVE') {
                setTimeout(() => this.monitorCurrentCapture(), 2000);
            } else {
                // Capture finished
                this.updateCaptureUI(false);
                this.currentCaptureId = null;
                this.loadCaptureHistory();
            }
            
        } catch (error) {
            console.error('Error monitoring capture:', error);
        }
    }

    formatCaptureStatus(status) {
        const statusMap = {
            'CAPTURE_STATUS_ACTIVE': 'Running',
            'CAPTURE_STATUS_COMPLETED': 'Completed',
            'CAPTURE_STATUS_FAILED': 'Failed',
            'CAPTURE_STATUS_CANCELLED': 'Cancelled'
        };
        return statusMap[status] || status;
    }

    getCaptureStatusClass(status) {
        const classMap = {
            'CAPTURE_STATUS_ACTIVE': 'badge bg-info',
            'CAPTURE_STATUS_COMPLETED': 'badge bg-success',
            'CAPTURE_STATUS_FAILED': 'badge bg-danger',
            'CAPTURE_STATUS_CANCELLED': 'badge bg-warning'
        };
        return classMap[status] || 'badge bg-secondary';
    }

    async loadCaptureHistory() {
        // This would need to be implemented in the backend
        // For now, just show a placeholder
        const container = document.getElementById('captureHistory');
        container.innerHTML = `
            <div class="alert alert-info">
                <i class="bi bi-info-circle"></i>
                Capture history functionality will be implemented in future versions.
            </div>
        `;
    }

    async createTarget() {
        try {
            const target = {
                name: document.getElementById('targetName').value,
                description: document.getElementById('targetDescription').value,
                host_regex: document.getElementById('hostRegex').value,
                ports: document.getElementById('targetPorts').value.split(',').map(p => parseInt(p.trim())),
                protocol_filter: document.getElementById('protocolFilter').value,
                enabled: document.getElementById('targetEnabled').checked
            };

            console.log('Creating target:', target);
            const response = await this.connectClient.createUrlTarget(target);
            
            if (response.created || response.success) {
                this.showNotification('URL target created successfully', 'success');
                bootstrap.Modal.getInstance(document.getElementById('createTargetModal')).hide();
                this.loadTargets();
                
                // Reset form
                document.getElementById('createTargetForm').reset();
            } else {
                this.showNotification('Failed to create URL target', 'error');
            }
        } catch (error) {
            console.error('Error creating target:', error);
            this.showNotification('Failed to create target: ' + error.message, 'error');
        }
    }

    async editTarget(targetId) {
        try {
            // Get target details
            const response = await this.connectClient.getUrlTarget(targetId);
            const target = response.target;
            
            if (!target) {
                this.showNotification('Target not found', 'error');
                return;
            }
            
            // Populate edit form
            document.getElementById('editTargetName').value = target.name || '';
            document.getElementById('editTargetDescription').value = target.description || '';
            document.getElementById('editHostRegex').value = target.hostRegex || '';
            document.getElementById('editTargetPorts').value = target.ports ? target.ports.join(', ') : '';
            document.getElementById('editProtocolFilter').value = target.protocolFilter || 'ALL';
            document.getElementById('editTargetEnabled').checked = target.enabled !== false;
            
            // Store target ID for update
            document.getElementById('editTargetForm').dataset.targetId = targetId;
            
            // Show edit modal
            this.showEditTargetModal();
            
        } catch (error) {
            console.error('Error loading target for edit:', error);
            this.showNotification('Failed to load target: ' + error.message, 'error');
        }
    }

    async updateTarget() {
        try {
            const form = document.getElementById('editTargetForm');
            const targetId = form.dataset.targetId;
            
            const targetData = {
                name: document.getElementById('editTargetName').value,
                description: document.getElementById('editTargetDescription').value,
                host_regex: document.getElementById('editHostRegex').value,
                ports: document.getElementById('editTargetPorts').value.split(',').map(p => parseInt(p.trim())).filter(p => !isNaN(p)),
                protocol_filter: document.getElementById('editProtocolFilter').value,
                enabled: document.getElementById('editTargetEnabled').checked
            };

            console.log('Updating target:', targetId, targetData);
            const response = await this.connectClient.updateUrlTarget(targetId, targetData);
            
            if (response.updated || response.success) {
                this.showNotification('Target updated successfully', 'success');
                bootstrap.Modal.getInstance(document.getElementById('editTargetModal')).hide();
                this.loadTargets();
            } else {
                this.showNotification('Failed to update target: ' + (response.message || 'Unknown error'), 'error');
            }
            
        } catch (error) {
            console.error('Error updating target:', error);
            this.showNotification('Failed to update target: ' + error.message, 'error');
        }
    }

    async deleteTarget(targetId) {
        if (!confirm('Are you sure you want to delete this URL target?')) {
            return;
        }

        try {
            console.log('Deleting target:', targetId);
            const response = await this.connectClient.deleteUrlTarget(targetId);
            
            if (response.deleted || response.success) {
                this.showNotification('Target deleted successfully', 'success');
                this.loadTargets(); // Reload to show updated list
            } else {
                this.showNotification('Failed to delete target: ' + (response.message || 'Unknown error'), 'error');
            }
            
        } catch (error) {
            console.error('Error deleting target:', error);
            this.showNotification('Failed to delete target: ' + error.message, 'error');
        }
    }

    async createDevice() {
        try {
            const deviceData = {
                device_name: document.getElementById('deviceName').value,
                description: document.getElementById('deviceDescription').value,
                ip_address: document.getElementById('deviceIpAddress').value,
                mac_address: document.getElementById('deviceMacAddress').value,
                interface: document.getElementById('deviceInterface').value,
                device_type: document.getElementById('deviceType').value,
                enabled: document.getElementById('deviceEnabled').checked
            };

            console.log('Creating device:', deviceData);
            const response = await this.connectClient.createDevice(deviceData);
            
            if (response.created || response.success) {
                this.showNotification('Device added successfully', 'success');
                bootstrap.Modal.getInstance(document.getElementById('createDeviceModal')).hide();
                this.loadDevices();
                
                // Reset form
                document.getElementById('createDeviceForm').reset();
            } else {
                this.showNotification('Failed to add device: ' + (response.message || 'Unknown error'), 'error');
            }
            
        } catch (error) {
            console.error('Error creating device:', error);
            this.showNotification('Failed to add device: ' + error.message, 'error');
        }
    }

    async editDevice(deviceId) {
        try {
            // Get device details
            const response = await this.connectClient.getDevice(deviceId);
            const device = response.device;
            
            if (!device) {
                this.showNotification('Device not found', 'error');
                return;
            }
            
            // Populate edit form
            document.getElementById('editDeviceName').value = device.deviceName || '';
            document.getElementById('editDeviceDescription').value = device.description || '';
            document.getElementById('editDeviceIpAddress').value = device.ipAddress || '';
            document.getElementById('editDeviceMacAddress').value = device.macAddress || '';
            document.getElementById('editDeviceInterface').value = device.interface || '';
            document.getElementById('editDeviceType').value = device.deviceType || 'DEVICE_TYPE_OTHER';
            document.getElementById('editDeviceEnabled').checked = device.enabled !== false;
            
            // Store device ID for update
            document.getElementById('editDeviceForm').dataset.deviceId = deviceId;
            
            // Show edit modal
            this.showEditDeviceModal();
            
        } catch (error) {
            console.error('Error loading device for edit:', error);
            this.showNotification('Failed to load device: ' + error.message, 'error');
        }
    }

    async updateDevice() {
        try {
            const form = document.getElementById('editDeviceForm');
            const deviceId = form.dataset.deviceId;
            
            const deviceData = {
                device_name: document.getElementById('editDeviceName').value,
                description: document.getElementById('editDeviceDescription').value,
                ip_address: document.getElementById('editDeviceIpAddress').value,
                mac_address: document.getElementById('editDeviceMacAddress').value,
                interface: document.getElementById('editDeviceInterface').value,
                device_type: document.getElementById('editDeviceType').value,
                enabled: document.getElementById('editDeviceEnabled').checked
            };

            console.log('Updating device:', deviceId, deviceData);
            const response = await this.connectClient.updateDevice(deviceId, deviceData);
            
            if (response.updated || response.success) {
                this.showNotification('Device updated successfully', 'success');
                bootstrap.Modal.getInstance(document.getElementById('editDeviceModal')).hide();
                this.loadDevices();
            } else {
                this.showNotification('Failed to update device: ' + (response.message || 'Unknown error'), 'error');
            }
            
        } catch (error) {
            console.error('Error updating device:', error);
            this.showNotification('Failed to update device: ' + error.message, 'error');
        }
    }

    async deleteDevice(deviceId) {
        if (!confirm('Are you sure you want to delete this device? If the device is currently connected, it will still appear in the list.')) {
            return;
        }

        try {
            console.log('Deleting device:', deviceId);
            const response = await this.connectClient.deleteDevice(deviceId);
            
            if (response.deleted || response.success) {
                this.showNotification('Device deleted successfully', 'success');
                this.loadDevices(); // Reload to show updated list
            } else {
                this.showNotification('Failed to delete device: ' + (response.message || 'Unknown error'), 'error');
            }
            
        } catch (error) {
            console.error('Error deleting device:', error);
            this.showNotification('Failed to delete device: ' + error.message, 'error');
        }
    }

    async registerDiscoveredDevice(macAddress) {
        try {
            // Find the discovered device in the current list
            const devicesList = this.currentDevicesList || [];
            const discoveredDevice = devicesList.find(d => d.macAddress === macAddress && d.isTemporary);
            
            if (!discoveredDevice) {
                this.showNotification('Discovered device not found', 'error');
                return;
            }
            
            // Pre-populate the form with discovered device info
            document.getElementById('deviceName').value = discoveredDevice.deviceName || 'Discovered Device';
            document.getElementById('deviceDescription').value = 'Auto-discovered network device';
            document.getElementById('deviceIpAddress').value = discoveredDevice.ipAddress || '';
            document.getElementById('deviceMacAddress').value = discoveredDevice.macAddress;
            document.getElementById('deviceInterface').value = discoveredDevice.interface || 'unknown';
            document.getElementById('deviceType').value = 'DEVICE_TYPE_OTHER';
            document.getElementById('deviceEnabled').checked = true;
            
            // Show the create device modal
            this.showCreateDeviceModal();
            
        } catch (error) {
            console.error('Error registering discovered device:', error);
            this.showNotification('Failed to register device: ' + error.message, 'error');
        }
    }

    showCreateTargetModal() {
        const modal = new bootstrap.Modal(document.getElementById('createTargetModal'));
        modal.show();
    }

    showCreateDeviceModal() {
        const modal = new bootstrap.Modal(document.getElementById('createDeviceModal'));
        modal.show();
    }

    showEditTargetModal() {
        const modal = new bootstrap.Modal(document.getElementById('editTargetModal'));
        modal.show();
    }

    showEditDeviceModal() {
        const modal = new bootstrap.Modal(document.getElementById('editDeviceModal'));
        modal.show();
    }

    updateSelectedCount(type) {
        if (type === 'devices') {
            const selectedDevices = document.querySelectorAll('#devicesList input:checked').length;
            document.getElementById('selectedDevicesCount').textContent = selectedDevices;
        } else if (type === 'targets') {
            const selectedTargets = document.querySelectorAll('#targetsList input:checked').length;
            document.getElementById('selectedTargetsCount').textContent = selectedTargets;
        }
    }

    getDeviceIcon(deviceType) {
        const icons = {
            'DEVICE_TYPE_MOBILE': 'phone',
            'DEVICE_TYPE_LAPTOP': 'laptop',
            'DEVICE_TYPE_DESKTOP': 'pc-display',
            'DEVICE_TYPE_TABLET': 'tablet',
            'DEVICE_TYPE_IOT': 'cpu',
            'DEVICE_TYPE_OTHER': 'router',
            'DEVICE_TYPE_UNSPECIFIED': 'router'
        };
        return icons[deviceType] || 'router';
    }
}

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', async () => {
    window.netTestLabApp = new NetTestLabApp();
    await window.netTestLabApp.init();
});