// Mission Control UI - JavaScript Application

class MissionControl {
    constructor() {
        this.themeStorageKey = 'mission-control-theme';
        this.init();
        this.setupEventListeners();
        this.setupKeyboardShortcuts();
        this.setupCardToggles();
        this.setupThemePreferenceWatcher();
        this.startPolling();
    }

    init() {
        // Initialize HTMX defaults
        if (typeof htmx !== 'undefined') {
            htmx.config.globalViewTransitions = true;
            htmx.config.defaultSwapStyle = 'outerHTML';
        }

        // Initialize state
        this.state = {
            mode: 'IDLE',
            recording: false,
            waypointMode: false,
            fullscreen: false,
            zoom: 1.0,
            powerPct: null,
            theme: document.documentElement.getAttribute('data-theme') || 'dark',
            connection: {
                connected: false,
                last_error: null
            }
        };
        const storedTheme = this.getStoredTheme();
        if (storedTheme && storedTheme !== this.state.theme) {
            document.documentElement.setAttribute('data-theme', storedTheme);
            this.state.theme = storedTheme;
        }

        this.flags = {
            chipsErrorLogged: false,
            modelsErrorLogged: false,
        };

        // Setup CSRF token for HTMX
        this.setupCSRF();
        this.updateModeDisplay();
    }

    setupCSRF() {
        // Generate CSRF token
        const csrfToken = this.generateCSRFToken();
        
        // Set HTMX headers
        if (typeof htmx !== 'undefined') {
            htmx.config.headers = {
                'X-CSRF-Token': csrfToken
            };
        }
        
        // Store for form submissions
        document.csrfToken = csrfToken;
    }

    generateCSRFToken() {
        const array = new Uint8Array(32);
        crypto.getRandomValues(array);
        return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
    }

    setupEventListeners() {
        // Video control buttons
        const fullscreenBtn = document.getElementById('feed-fullscreen-btn');
        const zoomInBtn = document.getElementById('feed-zoom-in-btn');
        const refreshBtn = document.getElementById('feed-refresh-btn');

        if (fullscreenBtn) {
            fullscreenBtn.addEventListener('click', () => this.toggleFullscreen());
        }

        if (zoomInBtn) {
            zoomInBtn.addEventListener('click', () => this.zoomIn());
        }

        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => this.refreshFeed());
        }

        // Control buttons
        const recordBtn = document.getElementById('btn-record');
        const rtlBtn = document.getElementById('btn-rtl');
        const waypointBtn = document.getElementById('btn-waypoint');

        if (recordBtn) {
            recordBtn.addEventListener('click', () => this.toggleRecording());
        }

        if (rtlBtn) {
            rtlBtn.addEventListener('click', () => this.returnToLaunch());
        }

        if (waypointBtn) {
            waypointBtn.addEventListener('click', () => this.toggleWaypointMode());
        }

        // Altitude controls
        const altitudeUp = document.getElementById('ctrl-altitude-up');
        const altitudeDown = document.getElementById('ctrl-altitude-down');

        if (altitudeUp) {
            altitudeUp.addEventListener('click', () => this.adjustAltitude(1));
        }

        if (altitudeDown) {
            altitudeDown.addEventListener('click', () => this.adjustAltitude(-1));
        }

        // Rotation controls
        const rotCCW = document.getElementById('ctrl-rot-ccw');
        const rotCW = document.getElementById('ctrl-rot-cw');

        if (rotCCW) {
            rotCCW.addEventListener('click', () => this.rotate(-5));
        }

        if (rotCW) {
            rotCW.addEventListener('click', () => this.rotate(5));
        }

        // Model toggle rows
        document.querySelectorAll('.model-row').forEach(row => {
            row.addEventListener('click', () => this.toggleModel(row));
        });

        // Detection box clicks
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('detection-box')) {
                this.inspectDetection(e.target);
            }
        });

        // Waypoint mode clicks on video
        const videoSurface = document.getElementById('feed-surface');
        if (videoSurface) {
            videoSurface.addEventListener('click', (e) => {
                if (this.state.waypointMode) {
                    this.setWaypoint(e);
                }
            });
        }

        this.connectionElements = {
            banner: document.getElementById('connection-banner'),
            stateLabel: document.getElementById('connection-state-label'),
            button: document.getElementById('connect-drone-btn')
        };
        this.logContainer = document.getElementById('event-log');

        if (this.connectionElements.banner) {
            this.connectionElements.banner.classList.add('disconnected');
        }

        if (this.connectionElements.button) {
            this.connectionElements.button.addEventListener('click', () => this.connectDrone());
        }

        const themeToggle = document.getElementById('theme-toggle');
        if (themeToggle) {
            themeToggle.addEventListener('click', () => this.toggleTheme());
        }

        this.updateThemeToggle();
    }

    setupKeyboardShortcuts() {
        document.addEventListener('keydown', (e) => {
            // Ignore if typing in input
            if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') {
                return;
            }

            switch (e.key.toLowerCase()) {
                case 'f':
                    e.preventDefault();
                    this.toggleFullscreen();
                    break;
                case '+':
                case '=':
                    e.preventDefault();
                    this.zoomIn();
                    break;
                case 'r':
                    e.preventDefault();
                    this.toggleRecording();
                    break;
                case 'w':
                    e.preventDefault();
                    this.toggleWaypointMode();
                    break;
                case 'escape':
                    e.preventDefault();
                    if (this.state.waypointMode) {
                        this.exitWaypointMode();
                    }
                    break;
                case 'arrowup':
                    e.preventDefault();
                    this.adjustAltitude(1);
                    break;
                case 'arrowdown':
                    e.preventDefault();
                    this.adjustAltitude(-1);
                    break;
                case '[':
                    e.preventDefault();
                    this.rotate(-5);
                    break;
                case ']':
                    e.preventDefault();
                    this.rotate(5);
                    break;
            }
        });
    }

    setupCardToggles() {
        this.collapseMap = new Map();

        document.querySelectorAll('.collapse-toggle').forEach(button => {
            const targetId = button.getAttribute('data-target');
            if (!targetId) {
                return;
            }

            this.collapseMap.set(targetId, button);
            button.dataset.collapsed = 'false';
            button.setAttribute('aria-expanded', 'true');
            button.textContent = '−';

            button.addEventListener('click', () => {
                const body = document.getElementById(targetId);
                if (!body) {
                    return;
                }
                const collapsed = body.classList.toggle('collapsed');
                button.dataset.collapsed = collapsed ? 'true' : 'false';
                button.setAttribute('aria-expanded', (!collapsed).toString());
                button.textContent = collapsed ? '+' : '−';
            });
        });

        document.addEventListener('htmx:afterSwap', (event) => {
            const targetId = event.target.id;
            if (!targetId) {
                return;
            }
            const button = this.collapseMap.get(targetId);
            if (!button) {
                return;
            }
            const collapsed = button.dataset.collapsed === 'true';
            if (collapsed) {
                event.target.classList.add('collapsed');
                button.setAttribute('aria-expanded', 'false');
                button.textContent = '+';
            }
        });
    }

    startPolling() {
        // HTMX handles most polling automatically
        // Add any additional polling logic here
        this.updateTime();
        setInterval(() => this.updateTime(), 1000);
        this.updateConnectionStatus();
        setInterval(() => this.updateConnectionStatus(), 5000);
        this.updateModels();
        setInterval(() => this.updateModels(), 5000);
        this.updateChips();
        setInterval(() => this.updateChips(), 7000);
    }

    // Connection Controls
    async updateConnectionStatus() {
        if (!this.connectionElements || !this.connectionElements.stateLabel) {
            return;
        }

        try {
            const response = await fetch('/api/connection/status');
            if (!response.ok) {
                throw new Error('Unable to fetch status');
            }
            const status = await response.json();
            const previousStatus = this.state.connection || {};
            this.state.connection = status;
            this.renderConnectionStatus(status, previousStatus);
        } catch (err) {
            this.appendLog('Connection status unavailable', 'warn');
            console.error('Connection status error', err);
        }
    }

    renderConnectionStatus(status, previousStatus = {}) {
        if (!this.connectionElements || !this.connectionElements.banner || !this.connectionElements.stateLabel) {
            return;
        }

        const { banner, stateLabel, button } = this.connectionElements;

        if (status.connected) {
            banner.classList.add('connected');
            banner.classList.remove('disconnected');
            stateLabel.textContent = 'Drone Connected';
            if (button) {
                button.disabled = true;
                button.textContent = 'Connected';
            }
            if (!previousStatus.connected) {
                this.appendLog('Drone connected', 'success');
            }
        } else {
            banner.classList.add('disconnected');
            banner.classList.remove('connected');
            stateLabel.textContent = 'Drone Disconnected';
            if (button) {
                button.disabled = false;
                button.textContent = 'Connect Drone';
            }
            if (previousStatus.connected) {
                this.appendLog('Drone disconnected', 'warn');
            }
            if (status.last_error && status.last_error !== previousStatus.last_error) {
                this.appendLog(status.last_error, 'error');
            }
        }
    }

    async connectDrone() {
        if (!this.connectionElements || !this.connectionElements.button) {
            return;
        }

        const button = this.connectionElements.button;
        button.disabled = true;
        button.textContent = 'Connecting...';
        this.appendLog('Attempting to connect to drone…', 'info');

        try {
            const response = await fetch('/api/connection/connect', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': document.csrfToken || ''
                }
            });

            const text = await response.text();
            let payload = {};
            if (text) {
                try {
                    payload = JSON.parse(text);
                } catch (parseErr) {
                    payload = { error: text };
                }
            }

            if (!response.ok) {
                throw new Error(payload.error || 'Failed to connect');
            }

            this.showToast(payload.message || 'Drone connected', 'success');
            this.appendLog(payload.message || 'Drone connected successfully', 'success');
            if (payload.status) {
                const previousStatus = this.state.connection || {};
                this.state.connection = payload.status;
                this.renderConnectionStatus(payload.status, previousStatus);
            } else {
                this.updateConnectionStatus();
            }
        } catch (err) {
            this.appendLog(err.message, 'error');
            this.showToast(err.message, 'error');
            button.disabled = false;
            button.textContent = 'Connect Drone';
        }
    }

    async updateModels() {
        const container = document.getElementById('models-list');
        if (!container) {
            return;
        }

        try {
            const response = await fetch('/api/models');
            if (!response.ok) {
                throw new Error('Unable to fetch models');
            }
            const models = await response.json();
            this.flags.modelsErrorLogged = false;
            this.renderModels(container, models);
        } catch (err) {
            if (!this.flags.modelsErrorLogged) {
                this.appendLog('Model list unavailable', 'warn');
                this.flags.modelsErrorLogged = true;
            }
        }
    }

    renderModels(container, models) {
        container.innerHTML = '';

        if (!models || models.length === 0) {
            container.innerHTML = '<div class="empty-state">No ML models configured yet.</div>';
            return;
        }

        models.forEach(model => {
            const id = model.id || model.ID || model.model_id || model.name || 'model';
            const state = (model.state || model.State || 'UNKNOWN').toUpperCase();
            const stateClass = (model.state_class || model.StateClass || 'neutral').toLowerCase();

            const row = document.createElement('div');
            row.className = 'model-row';
            row.dataset.model = id;
            row.dataset.currentState = state;

            const name = document.createElement('span');
            name.className = 'model-name';
            name.textContent = model.name || model.Name || id;

            const pill = document.createElement('span');
            pill.className = `pill pill-${stateClass}`;
            pill.textContent = state;

            row.appendChild(name);
            row.appendChild(pill);
            row.addEventListener('click', () => this.toggleModel(row));

            container.appendChild(row);
        });
    }

    async updateChips() {
        const powerChip = document.getElementById('app-power-chip');
        const modeChip = document.getElementById('app-mode-chip');

        if (!powerChip && !modeChip) {
            return;
        }

        try {
            const response = await fetch('/api/appchips');
            if (!response.ok) {
                throw new Error('Unable to fetch app chips');
            }

            const chips = await response.json();
            this.flags.chipsErrorLogged = false;

            if (powerChip && typeof chips.power_pct === 'number') {
                this.state.powerPct = chips.power_pct;
                powerChip.textContent = `PWR ${chips.power_pct}%`;
                powerChip.classList.toggle('chip-warning', chips.power_pct <= 25);
            }

            if (chips.mode) {
                this.state.mode = chips.mode.toUpperCase();
            }

            this.updateModeDisplay();
        } catch (err) {
            if (!this.flags.chipsErrorLogged) {
                this.appendLog('Power telemetry unavailable', 'warn');
                this.flags.chipsErrorLogged = true;
            }
        }
    }

    // Video Controls
    toggleFullscreen() {
        const videoSurface = document.getElementById('feed-surface');
        if (!videoSurface) return;

        if (!document.fullscreenElement) {
            videoSurface.requestFullscreen().then(() => {
                this.state.fullscreen = true;
                this.showToast('Entered fullscreen mode', 'success');
            }).catch(err => {
                this.showToast('Failed to enter fullscreen: ' + err.message, 'error');
            });
        } else {
            document.exitFullscreen().then(() => {
                this.state.fullscreen = false;
                this.showToast('Exited fullscreen mode', 'success');
            });
        }
    }

    zoomIn() {
        this.state.zoom = Math.min(this.state.zoom * 1.2, 3.0);
        this.applyZoom();
        this.showToast(`Zoom: ${Math.round(this.state.zoom * 100)}%`, 'success');
    }

    applyZoom() {
        const videoFrame = document.querySelector('.video-frame');
        if (videoFrame) {
            videoFrame.style.transform = `scale(${this.state.zoom})`;
        }
    }

    refreshFeed() {
        const img = document.querySelector('.video-frame');
        if (img) {
            const timestamp = Date.now();
            img.src = `/video.jpg?t=${timestamp}`;
            this.showToast('Refreshing video feed...', 'success');
        }

        // Poke the feed endpoint
        fetch('/api/feed/poke', { method: 'POST' })
            .then(response => response.json())
            .then(data => {
                this.showToast(data.message || 'Feed refreshed', 'success');
            })
            .catch(err => {
                this.showToast('Failed to refresh feed: ' + err.message, 'error');
            });
    }

    // Recording Controls
    toggleRecording() {
        const action = this.state.recording ? 'stop' : 'start';
        
        fetch('/api/controls/record', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.csrfToken
            },
            body: JSON.stringify({ action })
        })
        .then(response => response.json())
        .then(data => {
            this.state.recording = !this.state.recording;
            this.showToast(data.message || `Recording ${action}ed`, 'success');
            
            const recordBtn = document.getElementById('btn-record');
            if (recordBtn) {
                const text = recordBtn.querySelector('.btn-text') || recordBtn;
                text.textContent = this.state.recording ? 'Stop Recording' : 'Start Recording';
                recordBtn.classList.toggle('recording', this.state.recording);
            }

            this.updateModeDisplay();
        })
        .catch(err => {
            this.showToast('Failed to toggle recording: ' + err.message, 'error');
        });
    }

    // Flight Controls
    returnToLaunch() {
        if (!confirm('Return to launch? This will land the drone at its starting position.')) {
            return;
        }

        fetch('/api/controls/rtl', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.csrfToken
            },
            body: JSON.stringify({ confirm: true })
        })
        .then(response => response.json())
        .then(data => {
            this.state.mode = 'RTL';
            this.showToast('Returning to launch...', 'success');
            this.updateModeDisplay();
        })
        .catch(err => {
            this.showToast('Failed to RTL: ' + err.message, 'error');
        });
    }

    toggleWaypointMode() {
        this.state.waypointMode = !this.state.waypointMode;
        
        if (this.state.waypointMode) {
            this.state.mode = 'WAYPOINT_SET';
            this.showToast('Waypoint mode: Click on the video feed to set waypoint', 'success');
            document.body.style.cursor = 'crosshair';
        } else {
            this.exitWaypointMode();
        }
        
        this.updateModeDisplay();
    }

    exitWaypointMode() {
        this.state.waypointMode = false;
        this.state.mode = 'IDLE';
        document.body.style.cursor = 'default';
        this.showToast('Waypoint mode exited', 'success');
        this.updateModeDisplay();
    }

    setWaypoint(event) {
        const videoSurface = document.getElementById('feed-surface');
        if (!videoSurface) return;

        const rect = videoSurface.getBoundingClientRect();
        const x = (event.clientX - rect.left) / rect.width;
        const y = (event.clientY - rect.top) / rect.height;

        fetch('/api/controls/waypoint', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.csrfToken
            },
            body: JSON.stringify({ x_norm: x, y_norm: y })
        })
        .then(response => response.json())
        .then(data => {
            this.showToast('Waypoint set successfully', 'success');
            this.exitWaypointMode();
        })
        .catch(err => {
            this.showToast('Failed to set waypoint: ' + err.message, 'error');
        });
    }

    adjustAltitude(delta) {
        fetch('/api/controls/altitude', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.csrfToken
            },
            body: JSON.stringify({ delta_m: delta })
        })
        .then(response => response.json())
        .then(data => {
            this.showToast(`Altitude ${delta > 0 ? 'increased' : 'decreased'}`, 'success');
        })
        .catch(err => {
            this.showToast('Failed to adjust altitude: ' + err.message, 'error');
        });
    }

    rotate(delta) {
        fetch('/api/controls/rotation', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.csrfToken
            },
            body: JSON.stringify({ delta_deg: delta })
        })
        .then(response => response.json())
        .then(data => {
            this.showToast(`Rotated ${delta > 0 ? 'CW' : 'CCW'} ${Math.abs(delta)}°`, 'success');
        })
        .catch(err => {
            this.showToast('Failed to rotate: ' + err.message, 'error');
        });
    }

    // ML Model Controls
    toggleModel(row) {
        const modelId = row.dataset.model;
        const currentState = row.dataset.currentState;
        const nextState = currentState === 'ACTIVE' ? 'STANDBY' : 'ACTIVE';

        fetch('/api/models/toggle', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.csrfToken
            },
            body: JSON.stringify({ 
                model_id: modelId, 
                target_state: nextState 
            })
        })
        .then(response => response.text())
        .then(html => {
            // HTMX will handle the swap automatically
            this.showToast(`Model ${nextState.toLowerCase()}`, 'success');
            this.updateModels();
        })
        .catch(err => {
            this.showToast('Failed to toggle model: ' + err.message, 'error');
        });
    }

    // Detection Inspection
    inspectDetection(detectionBox) {
        const detectionId = detectionBox.dataset.id;
        
        fetch(`/api/detections/${detectionId}`)
            .then(response => response.json())
            .then(data => {
                this.showDetectionInspect(data);
            })
            .catch(err => {
                this.showToast('Failed to inspect detection: ' + err.message, 'error');
            });
    }

    showDetectionInspect(detection) {
        // Create or update inspect popup
        let inspect = document.getElementById('feed-inspect');
        if (!inspect) {
            inspect = document.createElement('div');
            inspect.id = 'feed-inspect';
            inspect.className = 'inspect-popup';
            document.body.appendChild(inspect);
        }

        inspect.innerHTML = `
            <div class="inspect-content">
                <h4>Detection Details</h4>
                <p><strong>Type:</strong> ${detection.type}</p>
                <p><strong>Confidence:</strong> ${Math.round(detection.confidence * 100)}%</p>
                ${detection.track_id ? `<p><strong>Track ID:</strong> ${detection.track_id}</p>` : ''}
                <button onclick="this.parentElement.parentElement.remove()">Close</button>
            </div>
        `;

        // Position near the detection
        const rect = detectionBox.getBoundingClientRect();
        inspect.style.left = rect.right + 10 + 'px';
        inspect.style.top = rect.top + 'px';
    }

    // Theme Controls
    toggleTheme() {
        const nextTheme = this.state.theme === 'light' ? 'dark' : 'light';
        this.applyTheme(nextTheme);
    }

    applyTheme(theme, persist = true) {
        if (!theme) {
            return;
        }

        document.documentElement.setAttribute('data-theme', theme);
        this.state.theme = theme;

        if (persist) {
            try {
                localStorage.setItem(this.themeStorageKey, theme);
            } catch (err) {
                // Storage might be disabled; ignore
            }
        }

        this.updateThemeToggle();
    }

    updateThemeToggle() {
        const toggle = document.getElementById('theme-toggle');
        if (!toggle) {
            return;
        }

        const currentTheme = this.state.theme || document.documentElement.getAttribute('data-theme') || 'dark';
        const isLight = currentTheme === 'light';

        toggle.setAttribute('aria-pressed', isLight.toString());
        toggle.dataset.theme = currentTheme;

        const icon = toggle.querySelector('.theme-toggle-icon');
        const label = toggle.querySelector('.theme-toggle-label');

        if (icon) {
            icon.textContent = isLight ? '☀️' : '🌙';
        }

        if (label) {
            label.textContent = isLight ? 'Light' : 'Dark';
        }
    }

    getStoredTheme() {
        try {
            return localStorage.getItem(this.themeStorageKey);
        } catch (err) {
            return null;
        }
    }

    setupThemePreferenceWatcher() {
        if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
            return;
        }

        this.systemThemeQuery = window.matchMedia('(prefers-color-scheme: light)');
        this.systemThemeListener = (event) => this.handleSystemThemeChange(event);

        if (typeof this.systemThemeQuery.addEventListener === 'function') {
            this.systemThemeQuery.addEventListener('change', this.systemThemeListener);
        } else if (typeof this.systemThemeQuery.addListener === 'function') {
            this.systemThemeQuery.addListener(this.systemThemeListener);
        }
    }

    handleSystemThemeChange(event) {
        if (this.getStoredTheme()) {
            return;
        }

        const nextTheme = event.matches ? 'light' : 'dark';
        this.applyTheme(nextTheme, false);
    }

    // UI Updates
    updateModeDisplay() {
        const modeChip = document.getElementById('app-mode-chip');
        if (!modeChip) {
            return;
        }

        if (this.state.recording) {
            modeChip.textContent = 'REC';
            modeChip.classList.add('chip-recording');
        } else {
            modeChip.textContent = this.state.mode || 'IDLE';
            modeChip.classList.remove('chip-recording');
        }
    }

    updateTime() {
        const hudElement = document.getElementById('feed-hud');
        if (hudElement) {
            const timeElement = hudElement.querySelector('.time');
            if (timeElement) {
                timeElement.textContent = new Date().toLocaleTimeString();
            }
        }
    }

    // Toast Notifications
    showToast(message, type = 'info') {
        const container = document.querySelector('.toast-container') || this.createToastContainer();
        
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = message;
        
        container.appendChild(toast);
        
        // Auto-remove after 3 seconds
        setTimeout(() => {
            toast.remove();
        }, 3000);
    }

    createToastContainer() {
        const container = document.createElement('div');
        container.className = 'toast-container';
        document.body.appendChild(container);
        return container;
    }

    // Error Handling
    handleHTMXError(event) {
        const error = event.detail.error;
        this.showToast('Request failed: ' + error.message, 'error');
        this.appendLog('HTMX error: ' + error.message, 'error');
    }

    appendLog(message, level = 'info') {
        if (!message || !this.logContainer) {
            return;
        }

        const placeholder = this.logContainer.querySelector('.log-entry.muted');
        if (placeholder) {
            placeholder.remove();
        }

        const entry = document.createElement('div');
        entry.className = `log-entry ${level}`;

        const timeEl = document.createElement('time');
        timeEl.textContent = new Date().toLocaleTimeString();

        const messageEl = document.createElement('span');
        messageEl.className = 'log-entry-message';
        messageEl.textContent = message;

        entry.appendChild(timeEl);
        entry.appendChild(messageEl);

        this.logContainer.prepend(entry);

        const maxEntries = 30;
        while (this.logContainer.children.length > maxEntries) {
            this.logContainer.removeChild(this.logContainer.lastChild);
        }
    }

    // Rate Limiting
    createRateLimiter(maxRequests = 20, windowMs = 1000) {
        const requests = [];
        
        return () => {
            const now = Date.now();
            const windowStart = now - windowMs;
            
            // Remove old requests
            while (requests.length > 0 && requests[0] < windowStart) {
                requests.shift();
            }
            
            if (requests.length >= maxRequests) {
                throw new Error('Rate limit exceeded');
            }
            
            requests.push(now);
        };
    }
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.missionControl = new MissionControl();
    
    // Setup HTMX error handling
    document.body.addEventListener('htmx:responseError', (event) => {
        window.missionControl.handleHTMXError(event);
    });
});

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = MissionControl;
}
