// Mission Control UI - JavaScript Application

class MissionControl {
    constructor() {
        this.init();
        this.setupEventListeners();
        this.setupKeyboardShortcuts();
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
            zoom: 1.0
        };

        // Setup CSRF token for HTMX
        this.setupCSRF();
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

    startPolling() {
        // HTMX handles most polling automatically
        // Add any additional polling logic here
        this.updateTime();
        setInterval(() => this.updateTime(), 1000);
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
            
            // Update button text via HTMX or manually
            const recordBtn = document.getElementById('btn-record');
            if (recordBtn) {
                recordBtn.textContent = this.state.recording ? 'Stop Recording' : 'Start Recording';
                recordBtn.classList.toggle('danger', this.state.recording);
            }
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

    // UI Updates
    updateModeDisplay() {
        const modeChip = document.getElementById('app-mode-chip');
        if (modeChip) {
            modeChip.textContent = this.state.mode;
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

// Rate limiters for control commands
const altitudeRateLimiter = new MissionControl().createRateLimiter();
const rotationRateLimiter = new MissionControl().createRateLimiter();

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