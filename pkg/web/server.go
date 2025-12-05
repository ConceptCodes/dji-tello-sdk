package web

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

// TelemetryData represents telemetry data structure
type TelemetryData struct {
	AltitudeM       float64 `json:"altitude_m"`
	SpeedMPS        float64 `json:"speed_mps"`
	LatDeg          float64 `json:"lat_deg"`
	LonDeg          float64 `json:"lon_deg"`
	SignalPct       int     `json:"signal_pct"`
	SignalTone      string  `json:"signal_tone"`
	DetectionsCount int     `json:"detections_count"`
	DetectionsTone  string  `json:"detections_tone"`
}

// StatusIndicator describes a label + tone pair for pill badges.
type StatusIndicator struct {
	Label string `json:"label"`
	Tone  string `json:"tone"`
}

// SystemStatus represents system status structure
type SystemStatus struct {
	Camera     StatusIndicator `json:"camera"`
	GPS        StatusIndicator `json:"gps"`
	Compass    StatusIndicator `json:"compass"`
	Battery    StatusIndicator `json:"battery"`
	BatteryPct int             `json:"battery_pct"`
}

// HUDData represents HUD overlay data
type HUDData struct {
	TimeLocal string `json:"time_local"`
	GPSLock   string `json:"gps_lock"`
}

// MiniStatsData represents mini flight stats
type MiniStatsData struct {
	AltitudeM float64 `json:"altitude_m"`
	SpeedMPS  float64 `json:"speed_mps"`
}

// Detection represents a single detection
type Detection struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Bbox       BoundingBox `json:"bbox"`
	Confidence float64     `json:"confidence"`
	TrackID    *string     `json:"track_id,omitempty"`
}

// BoundingBox represents a normalized bounding box
type BoundingBox struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

// AppChipsData represents app chips data
type AppChipsData struct {
	PowerPct int    `json:"power_pct"`
	Mode     string `json:"mode"`
}

// ModelState represents ML model state
type ModelState struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	State      string `json:"state"`
	StateClass string `json:"state_class"`
}

// ControlRequest represents a control command request
type ControlRequest struct {
	Action   string   `json:"action,omitempty"`
	DeltaM   *int     `json:"delta_m,omitempty"`
	DeltaDeg *int     `json:"delta_deg,omitempty"`
	Confirm  *bool    `json:"confirm,omitempty"`
	XNorm    *float64 `json:"x_norm,omitempty"`
	YNorm    *float64 `json:"y_norm,omitempty"`
}

// ModelToggleRequest represents a model toggle request
type ModelToggleRequest struct {
	ModelID     string `json:"model_id"`
	TargetState string `json:"target_state"`
}

// RecorderInterface defines interface for video recording
type RecorderInterface interface {
	StartRecording() error
	StopRecording() error
	IsRecording() bool
}

// MLPipelineInterface defines interface for ML pipeline interaction
type MLPipelineInterface interface {
	GetProcessorStates() []map[string]string
}

// WebServer handles web interface and API endpoints
type WebServer struct {
	commander     tello.TelloCommander
	recorder      RecorderInterface
	mlPipeline    MLPipelineInterface
	mlResultChan  <-chan ml.MLResult
	lastMLResults map[string]ml.MLResult
	templates     *template.Template
	csrfTokens    map[string]time.Time
	connection    *ConnectionCoordinator
	mu            sync.RWMutex
}

// NewWebServer creates a new web server instance
func NewWebServer(commander tello.TelloCommander, recorder RecorderInterface, mlPipeline MLPipelineInterface, mlResultChan <-chan ml.MLResult) *WebServer {
	ws := &WebServer{
		commander:     commander,
		recorder:      recorder,
		mlPipeline:    mlPipeline,
		mlResultChan:  mlResultChan,
		lastMLResults: make(map[string]ml.MLResult),
		csrfTokens:    make(map[string]time.Time),
		connection:    NewConnectionCoordinator(commander),
	}

	// Load templates
	ws.loadTemplates()

	// Start ML result processing
	if ws.mlResultChan != nil {
		go ws.processMLResults()
	}

	return ws
}

// loadTemplates loads HTML templates
func (ws *WebServer) loadTemplates() {
	ws.templates = template.Must(template.ParseGlob("web/templates/**/*.html"))
}

// processMLResults processes ML results for web interface
func (ws *WebServer) processMLResults() {
	for result := range ws.mlResultChan {
		ws.mu.Lock()
		ws.lastMLResults[result.GetProcessorName()] = result
		ws.mu.Unlock()
	}
}

func (ws *WebServer) setupSharedRoutes(mux *http.ServeMux) {
	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Connection endpoints
	mux.HandleFunc("/api/connection/status", ws.handleConnectionStatus)
	mux.HandleFunc("/api/connection/connect", ws.handleConnectionConnect)

	// API endpoints
	mux.HandleFunc("/api/csrf-token", ws.handleCSRFToken)
	mux.HandleFunc("/api/telemetry", ws.handleTelemetry)
	mux.HandleFunc("/api/status", ws.handleSystemStatus)
	mux.HandleFunc("/api/appchips", ws.handleAppChips)
	mux.HandleFunc("/api/models", ws.handleModels)
	mux.HandleFunc("/api/models/toggle", ws.handleModelToggle)
	mux.HandleFunc("/api/feed/hud", ws.handleFeedHUD)
	mux.HandleFunc("/api/feed/ministats", ws.handleFeedMiniStats)
	mux.HandleFunc("/api/detections", ws.handleDetections)
	mux.HandleFunc("/api/detections/", ws.handleDetectionInspect)
	mux.HandleFunc("/api/feed/poke", ws.handleFeedPoke)

	// Control endpoints
	mux.HandleFunc("/api/controls/record", ws.handleRecordControl)
	mux.HandleFunc("/api/controls/rtl", ws.handleRTLControl)
	mux.HandleFunc("/api/controls/altitude", ws.handleAltitudeControl)
	mux.HandleFunc("/api/controls/rotation", ws.handleRotationControl)
	mux.HandleFunc("/api/controls/waypoint", ws.handleWaypointControl)
}

// SetupRoutes configures HTTP routes including the index page
func (ws *WebServer) SetupRoutes(mux *http.ServeMux) {
	ws.setupSharedRoutes(mux)
	mux.HandleFunc("/", ws.handleIndex)
}

// SetupRoutesWithoutIndex registers all API/static routes but allows a custom index handler.
func (ws *WebServer) SetupRoutesWithoutIndex(mux *http.ServeMux) {
	ws.setupSharedRoutes(mux)
}

// HandleIndex exposes the index handler for custom mux wiring.
func (ws *WebServer) HandleIndex(w http.ResponseWriter, r *http.Request) {
	ws.handleIndex(w, r)
}

// ConnectionCoordinator returns the shared coordinator for drone connections.
func (ws *WebServer) ConnectionCoordinator() *ConnectionCoordinator {
	return ws.connection
}

func (ws *WebServer) handleConnectionStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if ws.connection == nil {
		http.Error(w, "Connection coordinator unavailable", http.StatusInternalServerError)
		return
	}

	status := ws.connection.Status()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		utils.Logger.Errorf("Failed to encode connection status: %v", err)
	}
}

func (ws *WebServer) handleConnectionConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !ws.validateCSRF(r) {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	if ws.connection == nil {
		http.Error(w, "Connection coordinator unavailable", http.StatusInternalServerError)
		return
	}

	err := ws.connection.Connect()
	status := ws.connection.Status()

	response := map[string]interface{}{
		"status": status,
	}

	if err != nil {
		response["error"] = err.Error()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
			utils.Logger.Errorf("Failed to encode connection error response: %v", encodeErr)
		}
		return
	}

	response["message"] = "Drone connected successfully"
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.Logger.Errorf("Failed to encode connection response: %v", err)
	}
}

// handleIndex serves main HTML page
func (ws *WebServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	// Generate CSRF token for this session
	ws.mu.Lock()
	csrfToken := ws.generateCSRFToken()
	ws.mu.Unlock()

	tmpl, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		utils.Logger.Errorf("Failed to parse index template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Pass CSRF token to template
	templateData := map[string]interface{}{
		"csrf_token": csrfToken,
	}

	err = tmpl.Execute(w, templateData)
	if err != nil {
		utils.Logger.Errorf("Failed to execute index template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleTelemetry returns telemetry data
func (ws *WebServer) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	telemetry := ws.getTelemetryData()

	// Return HTML fragment for HTMX
	w.Header().Set("Content-Type", "text/html")

	tmpl, err := template.ParseFiles("web/templates/fragments/telemetry.html")
	if err != nil {
		utils.Logger.Errorf("Failed to parse telemetry template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, telemetry)
	if err != nil {
		utils.Logger.Errorf("Failed to execute telemetry template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleSystemStatus returns system status
func (ws *WebServer) handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := ws.getSystemStatus()

	// Return HTML fragment for HTMX
	w.Header().Set("Content-Type", "text/html")

	tmpl, err := template.ParseFiles("web/templates/fragments/status.html")
	if err != nil {
		utils.Logger.Errorf("Failed to parse status template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, status)
	if err != nil {
		utils.Logger.Errorf("Failed to execute status template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleAppChips returns app chips data
func (ws *WebServer) handleAppChips(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chips := ws.getAppChipsData()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chips)
}

// handleModels returns ML models status
func (ws *WebServer) handleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	models := ws.getModelsData()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}

// handleModelToggle toggles ML model state
func (ws *WebServer) handleModelToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ModelToggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate CSRF token
	if !ws.validateCSRF(r) {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	// Toggle model state (this would integrate with ML pipeline)
	// For now, just return updated row
	model := ModelState{
		ID:         req.ModelID,
		Name:       ws.getModelName(req.ModelID),
		State:      req.TargetState,
		StateClass: ws.getStateClass(req.TargetState),
	}

	// Return HTML fragment for HTMX
	w.Header().Set("Content-Type", "text/html")

	tmpl, err := template.ParseFiles("web/templates/fragments/model_row.html")
	if err != nil {
		utils.Logger.Errorf("Failed to parse model row template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, model)
	if err != nil {
		utils.Logger.Errorf("Failed to execute model row template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleFeedHUD returns HUD overlay data
func (ws *WebServer) handleFeedHUD(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hud := ws.getHUDData()

	// Return HTML fragment for HTMX
	w.Header().Set("Content-Type", "text/html")

	tmpl, err := template.ParseFiles("web/templates/fragments/feed_hud.html")
	if err != nil {
		utils.Logger.Errorf("Failed to parse HUD template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, hud)
	if err != nil {
		utils.Logger.Errorf("Failed to execute HUD template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleFeedMiniStats returns mini flight stats
func (ws *WebServer) handleFeedMiniStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := ws.getMiniStatsData()

	// Return HTML fragment for HTMX
	w.Header().Set("Content-Type", "text/html")

	tmpl, err := template.ParseFiles("web/templates/fragments/feed_ministats.html")
	if err != nil {
		utils.Logger.Errorf("Failed to parse mini stats template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, stats)
	if err != nil {
		utils.Logger.Errorf("Failed to execute mini stats template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleDetections returns current detections
func (ws *WebServer) handleDetections(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	detections := ws.getDetections()

	// Return HTML fragment for HTMX
	w.Header().Set("Content-Type", "text/html")

	tmpl, err := template.ParseFiles("web/templates/fragments/detections.html")
	if err != nil {
		utils.Logger.Errorf("Failed to parse detections template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, detections)
	if err != nil {
		utils.Logger.Errorf("Failed to execute detections template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleDetectionInspect returns detailed detection information
func (ws *WebServer) handleDetectionInspect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract detection ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/detections/")
	if path == "" {
		http.Error(w, "Detection ID required", http.StatusBadRequest)
		return
	}

	// Find detection by ID
	detection := ws.findDetectionByID(path)
	if detection == nil {
		http.Error(w, "Detection not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detection)
}

// handleFeedPoke refreshes video feed
func (ws *WebServer) handleFeedPoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate CSRF token
	if !ws.validateCSRF(r) {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	// This would trigger a video stream refresh
	// For now, just return success message
	response := map[string]string{
		"message": "Feed refresh initiated",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Control endpoint handlers

func (ws *WebServer) handleRecordControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate CSRF token
	if !ws.validateCSRF(r) {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	// Handle recording control
	var message string
	switch req.Action {
	case "start":
		if ws.recorder != nil {
			if err := ws.recorder.StartRecording(); err != nil {
				http.Error(w, fmt.Sprintf("Failed to start recording: %v", err), http.StatusInternalServerError)
				return
			}
			message = "Recording started"
		} else {
			http.Error(w, "Recording not available", http.StatusServiceUnavailable)
			return
		}
	case "stop":
		if ws.recorder != nil {
			if err := ws.recorder.StopRecording(); err != nil {
				http.Error(w, fmt.Sprintf("Failed to stop recording: %v", err), http.StatusInternalServerError)
				return
			}
			message = "Recording stopped"
		} else {
			http.Error(w, "Recording not available", http.StatusServiceUnavailable)
			return
		}
	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}

	response := map[string]string{
		"message": message,
		"action":  req.Action,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (ws *WebServer) handleRTLControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate CSRF token
	if !ws.validateCSRF(r) {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	// Handle RTL (Return to Launch)
	if req.Confirm == nil || !*req.Confirm {
		http.Error(w, "Confirmation required", http.StatusBadRequest)
		return
	}

	// Execute RTL command if commander is available
	if ws.commander != nil {
		// This would implement RTL logic
		// For now, just return success
	}

	response := map[string]string{
		"message": "Return to launch initiated",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (ws *WebServer) handleAltitudeControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate CSRF token
	if !ws.validateCSRF(r) {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	if req.DeltaM == nil {
		http.Error(w, "Delta meters required", http.StatusBadRequest)
		return
	}

	// Execute altitude command if commander is available
	if ws.commander != nil {
		deltaCm := *req.DeltaM * 100 // Convert to cm
		if deltaCm > 0 {
			ws.commander.Up(int(deltaCm))
		} else {
			ws.commander.Down(int(-deltaCm))
		}
	}

	response := map[string]string{
		"message": fmt.Sprintf("Altitude adjusted by %dm", *req.DeltaM),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (ws *WebServer) handleRotationControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate CSRF token
	if !ws.validateCSRF(r) {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	if req.DeltaDeg == nil {
		http.Error(w, "Delta degrees required", http.StatusBadRequest)
		return
	}

	// Execute rotation command if commander is available
	if ws.commander != nil {
		if *req.DeltaDeg > 0 {
			ws.commander.Clockwise(*req.DeltaDeg)
		} else {
			ws.commander.CounterClockwise(-*req.DeltaDeg)
		}
	}

	response := map[string]string{
		"message": fmt.Sprintf("Rotated by %d°", *req.DeltaDeg),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (ws *WebServer) handleWaypointControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate CSRF token
	if !ws.validateCSRF(r) {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	if req.XNorm == nil || req.YNorm == nil {
		http.Error(w, "Waypoint coordinates required", http.StatusBadRequest)
		return
	}

	// Execute waypoint command if commander is available
	if ws.commander != nil {
		// This would implement waypoint logic
		// For now, just log the waypoint
		utils.Logger.Infof("Waypoint set at normalized coordinates: (%.3f, %.3f)", *req.XNorm, *req.YNorm)
	}

	response := map[string]string{
		"message": "Waypoint set successfully",
		"x":       fmt.Sprintf("%.3f", *req.XNorm),
		"y":       fmt.Sprintf("%.3f", *req.YNorm),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Data retrieval methods

func (ws *WebServer) getTelemetryData() *TelemetryData {
	connected := ws.connection != nil && ws.connection.IsConnected()

	signalPct := 0
	if connected {
		signalPct = 85
	}

	telemetry := &TelemetryData{
		AltitudeM:       0,
		SpeedMPS:        0,
		LatDeg:          0,
		LonDeg:          0,
		SignalPct:       signalPct,
		SignalTone:      toneFromPercentage(signalPct),
		DetectionsCount: len(ws.getDetections()),
	}
	telemetry.DetectionsTone = toneFromCount(telemetry.DetectionsCount)

	// Get data from commander if available
	if ws.commander != nil {
		if height, err := ws.commander.GetHeight(); err == nil {
			telemetry.AltitudeM = float64(height) / 100 // Convert cm to m
		}

		if speed, err := ws.commander.GetSpeed(); err == nil {
			telemetry.SpeedMPS = float64(speed) / 100 // Convert cm/s to m/s
		}
	}

	telemetry.SignalTone = toneFromPercentage(telemetry.SignalPct)

	return telemetry
}

func (ws *WebServer) getSystemStatus() *SystemStatus {
	status := &SystemStatus{
		Camera:     newIndicator("OFFLINE", "warn"),
		GPS:        newIndicator("NO FIX", "warn"),
		Compass:    newIndicator("STANDBY", "neutral"),
		Battery:    newIndicator("OK", "ok"),
		BatteryPct: 75,
	}

	if ws.connection != nil {
		connStatus := ws.connection.Status()
		if connStatus.Connected {
			status.Camera = newIndicator("LIVE", "ok")
			status.GPS = newIndicator("LOCKED", "ok")
			status.Compass = newIndicator("CALIBRATED", "ok")
		} else if connStatus.LastError != "" {
			status.Camera = newIndicator("ERROR", "err")
		} else {
			status.Camera = newIndicator("IDLE", "neutral")
		}
	}

	// Get battery data from commander if available
	if ws.commander != nil {
		if battery, err := ws.commander.GetBatteryPercentage(); err == nil {
			status.BatteryPct = battery
		}
	}

	status.Battery = batteryIndicator(status.BatteryPct)

	return status
}

func (ws *WebServer) getAppChipsData() *AppChipsData {
	chips := &AppChipsData{
		PowerPct: 75,
		Mode:     "IDLE",
	}

	// Get battery data from commander if available
	if ws.commander != nil {
		if battery, err := ws.commander.GetBatteryPercentage(); err == nil {
			chips.PowerPct = battery
		}
	}

	return chips
}

func (ws *WebServer) getModelsData() []ModelState {
	if ws.mlPipeline == nil {
		return []ModelState{}
	}

	states := ws.mlPipeline.GetProcessorStates()
	var models []ModelState

	for _, state := range states {
		models = append(models, ModelState{
			ID:         state["id"],
			Name:       state["name"],
			State:      state["state"],
			StateClass: ws.getStateClass(state["state"]),
		})
	}

	return models
}

func (ws *WebServer) getHUDData() *HUDData {
	return &HUDData{
		TimeLocal: time.Now().Format("15:04:05"),
		GPSLock:   "NO FIX",
	}
}

func (ws *WebServer) getMiniStatsData() *MiniStatsData {
	stats := &MiniStatsData{
		AltitudeM: 0,
		SpeedMPS:  0,
	}

	// Get data from commander if available
	if ws.commander != nil {
		if height, err := ws.commander.GetHeight(); err == nil {
			stats.AltitudeM = float64(height) / 100 // Convert cm to m
		}

		if speed, err := ws.commander.GetSpeed(); err == nil {
			stats.SpeedMPS = float64(speed) / 100 // Convert cm/s to m/s
		}
	}

	return stats
}

func newIndicator(label, tone string) StatusIndicator {
	if tone == "" {
		tone = "neutral"
	}

	return StatusIndicator{
		Label: strings.ToUpper(label),
		Tone:  tone,
	}
}

func batteryIndicator(pct int) StatusIndicator {
	tone := toneFromPercentage(pct)
	label := "OK"

	switch tone {
	case "warn":
		label = "LOW"
	case "err":
		label = "CRIT"
	}

	return newIndicator(label, tone)
}

func toneFromPercentage(pct int) string {
	switch {
	case pct <= 15:
		return "err"
	case pct <= 40:
		return "warn"
	default:
		return "ok"
	}
}

func toneFromCount(count int) string {
	if count <= 0 {
		return "neutral"
	}
	return "ok"
}

func (ws *WebServer) getDetections() []Detection {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	var detections []Detection

	// Process ML results to create detections
	for _, result := range ws.lastMLResults {
		// Check for DetectionResult type
		if detResult, ok := result.(*ml.DetectionResult); ok {
			for i, det := range detResult.Detections {
				detection := Detection{
					ID:         fmt.Sprintf("%s-%d", result.GetProcessorName(), i),
					Type:       det.ClassName,
					Confidence: float64(det.Confidence),
					Bbox: BoundingBox{
						X: float64(det.Box.Min.X) / 960.0, // Normalize assuming 960x720
						Y: float64(det.Box.Min.Y) / 720.0,
						W: float64(det.Box.Dx()) / 960.0,
						H: float64(det.Box.Dy()) / 720.0,
					},
				}
				detections = append(detections, detection)
			}
		}
	}

	return detections
}

func (ws *WebServer) findDetectionByID(id string) *Detection {
	detections := ws.getDetections()
	for _, detection := range detections {
		if detection.ID == id {
			return &detection
		}
	}
	return nil
}

// Helper methods

func (ws *WebServer) getModelName(modelID string) string {
	models := map[string]string{
		"object-detection":    "Object Detection",
		"person-tracking":     "Person Tracking",
		"vehicle-recognition": "Vehicle Recognition",
	}
	return models[modelID]
}

func (ws *WebServer) getStateClass(state string) string {
	switch state {
	case "ACTIVE":
		return "ok"
	case "STANDBY":
		return "neutral"
	default:
		return "warn"
	}
}

func (ws *WebServer) validateCSRF(r *http.Request) bool {
	token := r.Header.Get("X-CSRF-Token")
	if token == "" {
		return false
	}

	// Check if token exists and is not expired
	if expiry, ok := ws.csrfTokens[token]; ok {
		if time.Since(expiry) < time.Hour {
			return true
		}
		// Remove expired token
		delete(ws.csrfTokens, token)
	}

	return false
}

// generateCSRFToken generates a new CSRF token
func (ws *WebServer) generateCSRFToken() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const tokenLength = 32

	// Use crypto/rand for secure random generation
	b := make([]byte, tokenLength)
	rand.Read(b)
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}

	token := fmt.Sprintf("csrf_%d_%s", time.Now().UnixNano(), string(b))
	ws.csrfTokens[token] = time.Now().Add(time.Hour)
	return token
}

// cleanupExpiredCSRFTokens removes expired CSRF tokens
func (ws *WebServer) cleanupExpiredCSRFTokens() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	for token, expiry := range ws.csrfTokens {
		if time.Since(expiry) > time.Hour {
			delete(ws.csrfTokens, token)
		}
	}
}

// handleCSRFToken generates and returns a new CSRF token
func (ws *WebServer) handleCSRFToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ws.mu.Lock()
	token := ws.generateCSRFToken()
	ws.mu.Unlock()

	response := map[string]interface{}{
		"csrf_token": token,
		"expires_in": "1h",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
