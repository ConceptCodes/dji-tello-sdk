package tracking

import (
	"image"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
)

// Track represents a tracked object
type Track struct {
	ID           int                    `json:"id"`
	KalmanFilter *KalmanFilter          `json:"-"`
	Detections   []ml.Detection         `json:"detections"`
	LastUpdate   time.Time              `json:"last_update"`
	HitStreak    int                    `json:"hit_streak"`
	Age          int                    `json:"age"`
	Class        string                 `json:"class"`
	Confidence   float32                `json:"confidence"`
	State        TrackState             `json:"state"`
	Attributes   map[string]interface{} `json:"attributes,omitempty"`
}

// TrackState represents the state of a track
type TrackState int

const (
	TrackStateTentative TrackState = iota
	TrackStateConfirmed
	TrackStateDeleted
)

// String returns string representation of track state
func (ts TrackState) String() string {
	switch ts {
	case TrackStateTentative:
		return "tentative"
	case TrackStateConfirmed:
		return "confirmed"
	case TrackStateDeleted:
		return "deleted"
	default:
		return "unknown"
	}
}

// GetBoundingBox returns the current bounding box of the track
func (t *Track) GetBoundingBox() image.Rectangle {
	if t.KalmanFilter == nil {
		if len(t.Detections) > 0 {
			return t.Detections[len(t.Detections)-1].Box
		}
		return image.Rectangle{}
	}

	state := t.KalmanFilter.GetState()
	if len(state) >= 4 {
		// State format: [x, y, w, h, vx, vy, vw, vh]
		x := int(state[0])
		y := int(state[1])
		w := int(state[2])
		h := int(state[3])

		return image.Rect(x-w/2, y-h/2, x+w/2, y+h/2)
	}

	return image.Rectangle{}
}

// GetCenter returns the center point of the track
func (t *Track) GetCenter() image.Point {
	box := t.GetBoundingBox()
	return image.Point{
		X: box.Min.X + box.Dx()/2,
		Y: box.Min.Y + box.Dy()/2,
	}
}

// GetVelocity returns the velocity of the track
func (t *Track) GetVelocity() (vx, vy float64) {
	if t.KalmanFilter == nil {
		return 0, 0
	}

	state := t.KalmanFilter.GetState()
	if len(state) >= 6 {
		return state[4], state[5] // vx, vy
	}

	return 0, 0
}

// Update updates the track with a new detection
func (t *Track) Update(detection ml.Detection) {
	t.Detections = append(t.Detections, detection)
	t.LastUpdate = time.Now()
	t.HitStreak++
	t.Class = detection.ClassName
	t.Confidence = detection.Confidence

	// Keep only last N detections to prevent memory growth
	if len(t.Detections) > 50 {
		t.Detections = t.Detections[len(t.Detections)-50:]
	}
}

// Predict predicts the next position of the track
func (t *Track) Predict() image.Rectangle {
	if t.KalmanFilter == nil {
		return t.GetBoundingBox()
	}

	t.KalmanFilter.Predict()
	return t.GetBoundingBox()
}

// IsConfirmed returns whether the track is confirmed
func (t *Track) IsConfirmed() bool {
	return t.State == TrackStateConfirmed
}

// IsTentative returns whether the track is tentative
func (t *Track) IsTentative() bool {
	return t.State == TrackStateTentative
}

// IsDeleted returns whether the track is deleted
func (t *Track) IsDeleted() bool {
	return t.State == TrackStateDeleted
}

// Confirm confirms the track
func (t *Track) Confirm() {
	t.State = TrackStateConfirmed
}

// Delete marks the track as deleted
func (t *Track) Delete() {
	t.State = TrackStateDeleted
}

// GetTimeSinceUpdate returns time since last update
func (t *Track) GetTimeSinceUpdate() time.Duration {
	return time.Since(t.LastUpdate)
}

// GetAverageConfidence returns average confidence of all detections
func (t *Track) GetAverageConfidence() float32 {
	if len(t.Detections) == 0 {
		return 0
	}

	var sum float32
	for _, det := range t.Detections {
		sum += det.Confidence
	}

	return sum / float32(len(t.Detections))
}

// GetTrajectory returns the trajectory points of the track
func (t *Track) GetTrajectory() []image.Point {
	var points []image.Point
	for _, det := range t.Detections {
		center := image.Point{
			X: det.Box.Min.X + det.Box.Dx()/2,
			Y: det.Box.Min.Y + det.Box.Dy()/2,
		}
		points = append(points, center)
	}
	return points
}

// GetSmoothedTrajectory returns smoothed trajectory using Kalman filter
func (t *Track) GetSmoothedTrajectory() []image.Point {
	if t.KalmanFilter == nil {
		return t.GetTrajectory()
	}

	// This would require storing Kalman filter history
	// For now, return current trajectory
	return t.GetTrajectory()
}

// Clone creates a copy of the track
func (t *Track) Clone() *Track {
	clone := &Track{
		ID:         t.ID,
		Detections: make([]ml.Detection, len(t.Detections)),
		LastUpdate: t.LastUpdate,
		HitStreak:  t.HitStreak,
		Age:        t.Age,
		Class:      t.Class,
		Confidence: t.Confidence,
		State:      t.State,
		Attributes: make(map[string]interface{}),
	}

	// Clone KalmanFilter if it exists
	if t.KalmanFilter != nil {
		clone.KalmanFilter = t.KalmanFilter.Clone()
	}

	copy(clone.Detections, t.Detections)

	for k, v := range t.Attributes {
		clone.Attributes[k] = v
	}

	return clone
}
