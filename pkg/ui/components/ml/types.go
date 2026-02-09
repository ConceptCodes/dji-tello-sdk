package ml

import (
	"image"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
)

// TrackVisualization represents a visualized track
type TrackVisualization struct {
	ID         int
	ClassName  string
	Confidence float32
	Box        image.Rectangle
	State      ml.TrackState
	Age        int
	Hits       int
	Misses     int
	Velocity   ml.Point3D
	Prediction image.Rectangle
	Color      string
	History    []image.Point
	LastUpdate time.Time
}

// MLMetrics represents ML pipeline metrics for display
type MLMetrics struct {
	FPS            float64
	Latency        time.Duration
	DroppedFrames  int64
	MemoryUsage    int64
	GPUUsage       float64
	LastUpdate     time.Time
	ProcessorStats map[string]ProcessorStats
}

// ProcessorStats represents statistics for a specific processor
type ProcessorStats struct {
	Name          string
	ProcessTime   time.Duration
	SuccessCount  int64
	ErrorCount    int64
	AvgLatency    time.Duration
	LastProcessed time.Time
	Enabled       bool
}

// MLConfig represents ML configuration for UI
type MLConfig struct {
	Enabled         bool
	ShowTracks      bool
	ShowDetections  bool
	ShowConfidence  bool
	ShowVelocity    bool
	ShowPredictions bool
	MaxTracks       int
	TrackHistory    int
	ColorScheme     string
}

// MLState represents the current state of ML visualization
type MLState struct {
	Tracks     []TrackVisualization
	Metrics    MLMetrics
	Config     MLConfig
	LastUpdate time.Time
	Active     bool
	Error      string
}

// NewDefaultMLConfig creates a default ML configuration
func NewDefaultMLConfig() MLConfig {
	return MLConfig{
		Enabled:         true,
		ShowTracks:      true,
		ShowDetections:  true,
		ShowConfidence:  true,
		ShowVelocity:    false,
		ShowPredictions: false,
		MaxTracks:       20,
		TrackHistory:    10,
		ColorScheme:     "default",
	}
}

// NewDefaultMLState creates a default ML state
func NewDefaultMLState() MLState {
	return MLState{
		Tracks:     make([]TrackVisualization, 0),
		Metrics:    MLMetrics{ProcessorStats: make(map[string]ProcessorStats)},
		Config:     NewDefaultMLConfig(),
		LastUpdate: time.Now(),
		Active:     false,
	}
}

// UpdateFromMLResult updates the ML state from ML pipeline results
func (state *MLState) UpdateFromMLResult(result ml.MLResult) {
	state.LastUpdate = time.Now()
	state.Active = true

	switch r := result.(type) {
	case ml.TrackingResult:
		state.updateFromTrackingResult(r)
	case ml.DetectionResult:
		state.updateFromDetectionResult(r)
	}
}

// UpdateMetrics updates the ML metrics
func (state *MLState) UpdateMetrics(metrics ml.PipelineMetrics) {
	state.Metrics.FPS = metrics.FPS
	state.Metrics.Latency = metrics.Latency
	state.Metrics.DroppedFrames = metrics.DroppedFrames
	state.Metrics.MemoryUsage = metrics.MemoryUsage
	state.Metrics.GPUUsage = metrics.GPUUsage
	state.Metrics.LastUpdate = metrics.LastUpdate

	// Note: metrics.ProcessorStats is map[string]float64, not map[string]ml.ProcessorStats
	// We'll handle processor stats differently if needed
}

// updateFromTrackingResult updates tracks from tracking result
func (state *MLState) updateFromTrackingResult(result ml.TrackingResult) {
	// Clear old tracks if needed
	if len(state.Tracks) > state.Config.MaxTracks {
		state.Tracks = state.Tracks[:state.Config.MaxTracks]
	}

	// Update existing tracks and add new ones
	for _, track := range result.Tracks {
		if track.State == ml.TrackStateDeleted {
			// Remove deleted tracks
			state.removeTrack(track.ID)
			continue
		}

		// Find existing track
		existingIdx := -1
		for i, t := range state.Tracks {
			if t.ID == track.ID {
				existingIdx = i
				break
			}
		}

		if existingIdx >= 0 {
			// Update existing track
			state.updateTrack(existingIdx, track)
		} else {
			// Add new track
			state.addTrack(track)
		}
	}
}

// updateFromDetectionResult updates from detection result
func (state *MLState) updateFromDetectionResult(result ml.DetectionResult) {
	// For now, just mark that we have detections
	// In a more advanced implementation, we could visualize detections separately
}

// addTrack adds a new track to the state
func (state *MLState) addTrack(track ml.Track) {
	if len(state.Tracks) >= state.Config.MaxTracks {
		// Remove oldest track
		state.Tracks = state.Tracks[1:]
	}

	center := image.Point{
		X: track.Box.Min.X + track.Box.Dx()/2,
		Y: track.Box.Min.Y + track.Box.Dy()/2,
	}

	visualization := TrackVisualization{
		ID:         track.ID,
		ClassName:  track.ClassName,
		Confidence: track.Confidence,
		Box:        track.Box,
		State:      track.State,
		Age:        track.Age,
		Hits:       track.Hits,
		Misses:     track.Misses,
		Velocity:   track.Velocity,
		Prediction: track.Prediction,
		Color:      state.getColorForTrack(track.ID),
		History:    []image.Point{center},
		LastUpdate: time.Now(),
	}

	state.Tracks = append(state.Tracks, visualization)
}

// updateTrack updates an existing track
func (state *MLState) updateTrack(idx int, track ml.Track) {
	center := image.Point{
		X: track.Box.Min.X + track.Box.Dx()/2,
		Y: track.Box.Min.Y + track.Box.Dy()/2,
	}

	// Update track
	state.Tracks[idx].ClassName = track.ClassName
	state.Tracks[idx].Confidence = track.Confidence
	state.Tracks[idx].Box = track.Box
	state.Tracks[idx].State = track.State
	state.Tracks[idx].Age = track.Age
	state.Tracks[idx].Hits = track.Hits
	state.Tracks[idx].Misses = track.Misses
	state.Tracks[idx].Velocity = track.Velocity
	state.Tracks[idx].Prediction = track.Prediction
	state.Tracks[idx].LastUpdate = time.Now()

	// Add to history
	state.Tracks[idx].History = append(state.Tracks[idx].History, center)
	if len(state.Tracks[idx].History) > state.Config.TrackHistory {
		state.Tracks[idx].History = state.Tracks[idx].History[1:]
	}
}

// removeTrack removes a track by ID
func (state *MLState) removeTrack(id int) {
	for i, track := range state.Tracks {
		if track.ID == id {
			state.Tracks = append(state.Tracks[:i], state.Tracks[i+1:]...)
			return
		}
	}
}

// getColorForTrack returns a color for a track ID
func (state *MLState) getColorForTrack(id int) string {
	// Simple color mapping based on track ID
	colors := []string{
		"red", "green", "blue", "yellow", "magenta", "cyan",
		"bright-red", "bright-green", "bright-blue", "bright-yellow",
		"bright-magenta", "bright-cyan",
	}
	return colors[id%len(colors)]
}

// GetTrackCount returns the number of active tracks
func (state *MLState) GetTrackCount() int {
	return len(state.Tracks)
}

// GetConfirmedTrackCount returns the number of confirmed tracks
func (state *MLState) GetConfirmedTrackCount() int {
	count := 0
	for _, track := range state.Tracks {
		if track.State == ml.TrackStateConfirmed {
			count++
		}
	}
	return count
}

// GetTentativeTrackCount returns the number of tentative tracks
func (state *MLState) GetTentativeTrackCount() int {
	count := 0
	for _, track := range state.Tracks {
		if track.State == ml.TrackStateTentative {
			count++
		}
	}
	return count
}
