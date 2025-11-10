package tracking

import (
	"image"
	"math"
	"sort"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
)

// TrackerConfig defines configuration for the object tracker
type TrackerConfig struct {
	MaxDistance      float64 `json:"max_distance"`      // Maximum distance for association
	MaxAge           int     `json:"max_age"`           // Maximum frames without update before deletion
	MinHits          int     `json:"min_hits"`          // Minimum hits before confirming track
	MaxIOUDistance   float64 `json:"max_iou_distance"`  // Maximum IoU distance for association
	UseKalmanFilter  bool    `json:"use_kalman_filter"` // Whether to use Kalman filtering
	EnablePrediction bool    `json:"enable_prediction"` // Whether to predict positions
}

// ObjectTracker manages multiple object tracks
type ObjectTracker struct {
	config     TrackerConfig
	tracks     []*Track
	nextID     int
	frameCount int
}

// NewObjectTracker creates a new object tracker
func NewObjectTracker(config TrackerConfig) *ObjectTracker {
	return &ObjectTracker{
		config: config,
		tracks: make([]*Track, 0),
		nextID: 1,
	}
}

// Update updates the tracker with new detections
func (ot *ObjectTracker) Update(detections []ml.Detection) []*Track {
	ot.frameCount++

	// Predict track positions if enabled
	if ot.config.EnablePrediction {
		for _, track := range ot.tracks {
			if !track.IsDeleted() {
				track.Predict()
			}
		}
	}

	// Associate detections to existing tracks
	matches, unmatchedDetections, unmatchedTracks := ot.associateDetectionsToTracks(detections)

	// Update matched tracks
	for _, match := range matches {
		track := ot.tracks[match.trackIdx]
		detection := detections[match.detectionIdx]

		// Update Kalman filter if enabled
		if ot.config.UseKalmanFilter && track.KalmanFilter != nil {
			measurement := ot.detectionToMeasurement(detection)
			track.KalmanFilter.Update(measurement)
		}

		track.Update(detection)
		track.HitStreak++

		// Confirm track if it has enough hits
		if track.IsTentative() && track.HitStreak >= ot.config.MinHits {
			track.Confirm()
		}
	}

	// Create new tracks for unmatched detections
	for _, detIdx := range unmatchedDetections {
		detection := detections[detIdx]
		newTrack := ot.createNewTrack(detection)
		ot.tracks = append(ot.tracks, newTrack)
	}

	// Mark unmatched tracks as missed
	for _, trackIdx := range unmatchedTracks {
		track := ot.tracks[trackIdx]
		track.HitStreak = 0
		track.Age++

		// Delete old tracks
		if track.Age > ot.config.MaxAge {
			track.Delete()
		}
	}

	// Remove deleted tracks
	ot.tracks = ot.removeDeletedTracks()

	return ot.getActiveTracks()
}

// associateDetectionsToTracks associates detections to existing tracks
func (ot *ObjectTracker) associateDetectionsToTracks(detections []ml.Detection) ([]Match, []int, []int) {
	if len(ot.tracks) == 0 {
		unmatchedDetections := make([]int, len(detections))
		for i := range detections {
			unmatchedDetections[i] = i
		}
		return nil, unmatchedDetections, nil
	}

	// Calculate cost matrix
	costMatrix := ot.calculateCostMatrix(detections)

	// Use Hungarian algorithm for optimal assignment
	matches, unmatchedDetections, unmatchedTracks := ot.hungarianAssignment(costMatrix)

	return matches, unmatchedDetections, unmatchedTracks
}

// Match represents a detection-track match
type Match struct {
	detectionIdx int
	trackIdx     int
	cost         float64
}

// calculateCostMatrix calculates cost matrix for assignment
func (ot *ObjectTracker) calculateCostMatrix(detections []ml.Detection) [][]float64 {
	costMatrix := make([][]float64, len(detections))

	for i := range costMatrix {
		costMatrix[i] = make([]float64, len(ot.tracks))
	}

	for i, detection := range detections {
		for j, track := range ot.tracks {
			if track.IsDeleted() {
				costMatrix[i][j] = math.Inf(1)
				continue
			}

			cost := ot.calculateCost(detection, track)
			costMatrix[i][j] = cost
		}
	}

	return costMatrix
}

// calculateCost calculates cost between detection and track
func (ot *ObjectTracker) calculateCost(detection ml.Detection, track *Track) float64 {
	// Use IoU distance as primary cost
	iou := calculateIoU(detection.Box, track.GetBoundingBox())
	iouCost := 1.0 - iou

	// Use center distance as secondary cost
	detCenter := image.Point{
		X: detection.Box.Min.X + detection.Box.Dx()/2,
		Y: detection.Box.Min.Y + detection.Box.Dy()/2,
	}
	trackCenter := track.GetCenter()

	dx := float64(detCenter.X - trackCenter.X)
	dy := float64(detCenter.Y - trackCenter.Y)
	euclideanDist := math.Sqrt(dx*dx + dy*dy)

	// Normalize distance by image size (assuming max dimension 1000)
	normalizedDist := euclideanDist / 1000.0

	// Combine costs
	combinedCost := 0.7*float64(iouCost) + 0.3*normalizedDist

	// Apply class matching penalty
	if detection.ClassName != track.Class {
		combinedCost += 0.5
	}

	return combinedCost
}

// hungarianAssignment performs Hungarian algorithm for assignment
func (ot *ObjectTracker) hungarianAssignment(costMatrix [][]float64) ([]Match, []int, []int) {
	if len(costMatrix) == 0 || len(costMatrix[0]) == 0 {
		return nil, nil, nil
	}

	// Simplified greedy assignment (for now)
	// In production, you'd implement the full Hungarian algorithm
	var matches []Match
	usedDetections := make([]bool, len(costMatrix))
	usedTracks := make([]bool, len(costMatrix[0]))

	// Sort all possible matches by cost
	var allMatches []Match
	for i := range costMatrix {
		for j := range costMatrix[i] {
			if costMatrix[i][j] < ot.config.MaxIOUDistance {
				allMatches = append(allMatches, Match{
					detectionIdx: i,
					trackIdx:     j,
					cost:         costMatrix[i][j],
				})
			}
		}
	}

	sort.Slice(allMatches, func(i, j int) bool {
		return allMatches[i].cost < allMatches[j].cost
	})

	// Greedy assignment
	for _, match := range allMatches {
		if !usedDetections[match.detectionIdx] && !usedTracks[match.trackIdx] {
			matches = append(matches, match)
			usedDetections[match.detectionIdx] = true
			usedTracks[match.trackIdx] = true
		}
	}

	// Find unmatched detections and tracks
	var unmatchedDetections, unmatchedTracks []int
	for i := range usedDetections {
		if !usedDetections[i] {
			unmatchedDetections = append(unmatchedDetections, i)
		}
	}

	for i := range usedTracks {
		if !usedTracks[i] {
			unmatchedTracks = append(unmatchedTracks, i)
		}
	}

	return matches, unmatchedDetections, unmatchedTracks
}

// createNewTrack creates a new track from detection
func (ot *ObjectTracker) createNewTrack(detection ml.Detection) *Track {
	track := &Track{
		ID:         ot.nextID,
		Detections: []ml.Detection{detection},
		LastUpdate: time.Now(),
		HitStreak:  1,
		Age:        0,
		Class:      detection.ClassName,
		Confidence: detection.Confidence,
		State:      TrackStateTentative,
		Attributes: make(map[string]interface{}),
	}

	// Initialize Kalman filter if enabled
	if ot.config.UseKalmanFilter {
		box := ot.detectionToMeasurement(detection)
		track.KalmanFilter = NewKalmanFilter(box)
	}

	ot.nextID++
	return track
}

// detectionToMeasurement converts detection to measurement vector
func (ot *ObjectTracker) detectionToMeasurement(detection ml.Detection) []float64 {
	// Convert bounding box to center-width-height format
	cx := float64(detection.Box.Min.X + detection.Box.Dx()/2)
	cy := float64(detection.Box.Min.Y + detection.Box.Dy()/2)
	w := float64(detection.Box.Dx())
	h := float64(detection.Box.Dy())

	return []float64{cx, cy, w, h}
}

// getActiveTracks returns all active (non-deleted) tracks
func (ot *ObjectTracker) getActiveTracks() []*Track {
	var activeTracks []*Track
	for _, track := range ot.tracks {
		if !track.IsDeleted() {
			activeTracks = append(activeTracks, track)
		}
	}
	return activeTracks
}

// removeDeletedTracks removes deleted tracks from the list
func (ot *ObjectTracker) removeDeletedTracks() []*Track {
	var activeTracks []*Track
	for _, track := range ot.tracks {
		if !track.IsDeleted() {
			activeTracks = append(activeTracks, track)
		}
	}
	return activeTracks
}

// GetTracks returns all current tracks
func (ot *ObjectTracker) GetTracks() []*Track {
	return ot.getActiveTracks()
}

// GetConfirmedTracks returns only confirmed tracks
func (ot *ObjectTracker) GetConfirmedTracks() []*Track {
	var confirmedTracks []*Track
	for _, track := range ot.tracks {
		if track.IsConfirmed() {
			confirmedTracks = append(confirmedTracks, track)
		}
	}
	return confirmedTracks
}

// GetTrackCount returns the number of active tracks
func (ot *ObjectTracker) GetTrackCount() int {
	return len(ot.getActiveTracks())
}

// GetConfirmedTrackCount returns the number of confirmed tracks
func (ot *ObjectTracker) GetConfirmedTrackCount() int {
	return len(ot.GetConfirmedTracks())
}

// Reset resets the tracker
func (ot *ObjectTracker) Reset() {
	ot.tracks = make([]*Track, 0)
	ot.nextID = 1
	ot.frameCount = 0
}

// GetFrameCount returns the current frame count
func (ot *ObjectTracker) GetFrameCount() int {
	return ot.frameCount
}

// GetConfig returns the tracker configuration
func (ot *ObjectTracker) GetConfig() TrackerConfig {
	return ot.config
}

// UpdateConfig updates the tracker configuration
func (ot *ObjectTracker) UpdateConfig(config TrackerConfig) {
	ot.config = config
}
