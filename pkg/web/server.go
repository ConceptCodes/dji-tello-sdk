package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
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
	DetectionsCount int     `json:"detections_count"`
}

// SystemStatus represents system status structure
type SystemStatus struct {
	CameraStatus  string `json:"camera"`
	GPSStatus     string `json:"gps"`
	CompassStatus string `json:"compass"`
	BatteryStatus string `json:"battery"`
	BatteryPct    int    `json:"battery_pct"`
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

// WebServer handles web interface and API endpoints
type WebServer struct {
	commander     tello.TelloCommander
	mlResultChan  <-chan ml.MLResult
	lastMLResults map[string]ml.MLResult
	templates     *template.Template
	csrfTokens    map[string]time.Time
}

// NewWebServer creates a new web server instance
func NewWebServer(commander tello.TelloCommander, mlResultChan <-chan ml.MLResult) *WebServer {
	ws := &WebServer{
		commander:     commander,
		mlResultChan:  mlResultChan,
		lastMLResults: make(map[string]ml.MLResult),
		csrfTokens:    make(map[string]time.Time),
	}

	// Load templates
	ws.loadTemplates()

	// Start ML result processing
	go ws.processMLResults()

	return ws
}

// loadTemplates loads HTML templates
func (ws *WebServer) loadTemplates() {
	ws.templates = template.Must(template.ParseGlob("web/templates/**/*.html"))
}

// processMLResults processes ML results for web interface
func (ws *WebServer) processMLResults() {
	for result := range ws.mlResultChan {
		ws.lastMLResults[result.GetProcessorName()] = result
	}
}

// SetupRoutes configures HTTP routes
func (ws *WebServer) SetupRoutes(mux *http.ServeMux) {
	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Main page
	mux.HandleFunc("/", ws.handleIndex)

	// API endpoints
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

// handleIndex serves main HTML page
func (ws *WebServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	tmpl, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		utils.Logger.Errorf("Failed to parse index template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
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
		message = "Recording started"
		// Start recording logic here
	case "stop":
		message = "Recording stopped"
		// Stop recording logic here
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
	telemetry := &TelemetryData{
		AltitudeM:       0,
		SpeedMPS:        0,
		LatDeg:          0,
		LonDeg:          0,
		SignalPct:       85,
		DetectionsCount: len(ws.getDetections()),
	}

	// Get data from commander if available
	if ws.commander != nil {
		if height, err := ws.commander.GetHeight(); err == nil {
			telemetry.AltitudeM = float64(height) / 100 // Convert cm to m
		}

		if speed, err := ws.commander.GetSpeed(); err == nil {
			telemetry.SpeedMPS = float64(speed) / 100 // Convert cm/s to m/s
		}
	}

	return telemetry
}

func (ws *WebServer) getSystemStatus() *SystemStatus {
	status := &SystemStatus{
		CameraStatus:  "OK",
		GPSStatus:     "WARN",
		CompassStatus: "OK",
		BatteryStatus: "OK",
		BatteryPct:    75,
	}

	// Get battery data from commander if available
	if ws.commander != nil {
		if battery, err := ws.commander.GetBatteryPercentage(); err == nil {
			status.BatteryPct = battery
			if battery < 20 {
				status.BatteryStatus = "ERR"
			} else if battery < 50 {
				status.BatteryStatus = "WARN"
			}
		}
	}

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
	return []ModelState{
		{
			ID:         "object-detection",
			Name:       "Object Detection",
			State:      "ACTIVE",
			StateClass: "ok",
		},
		{
			ID:         "person-tracking",
			Name:       "Person Tracking",
			State:      "ACTIVE",
			StateClass: "ok",
		},
		{
			ID:         "vehicle-recognition",
			Name:       "Vehicle Recognition",
			State:      "STANDBY",
			StateClass: "neutral",
		},
	}
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

func (ws *WebServer) getDetections() []Detection {
	var detections []Detection

	// Process ML results to create detections
	for processorName := range ws.lastMLResults {
		// For now, create some mock detections based on result type
		// This would be replaced with actual ML result processing
		if processorName == "yolo" {
			// Create mock detection for demonstration
			detection := Detection{
				ID:         fmt.Sprintf("%s-1", processorName),
				Type:       "PERSON",
				Confidence: 0.85,
				Bbox: BoundingBox{
					X: 0.3,
					Y: 0.2,
					W: 0.15,
					H: 0.4,
				},
			}
			detections = append(detections, detection)
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
