package tracking

import (
	"image"
	"testing"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
)

func TestTrack_GetBoundingBox(t *testing.T) {
	detection := ml.Detection{
		Box: image.Rect(10, 10, 50, 50),
	}

	track := &Track{
		Detections: []ml.Detection{detection},
	}

	box := track.GetBoundingBox()

	if !box.Eq(image.Rect(10, 10, 50, 50)) {
		t.Errorf("Expected box {10,10,50,50}, got %v", box)
	}
}

func TestTrack_GetCenter(t *testing.T) {
	detection := ml.Detection{
		Box: image.Rect(10, 10, 50, 50),
	}

	track := &Track{
		Detections: []ml.Detection{detection},
	}

	center := track.GetCenter()
	expected := image.Point{X: 30, Y: 30}

	if center != expected {
		t.Errorf("Expected center %v, got %v", expected, center)
	}
}

func TestTrack_GetVelocity(t *testing.T) {
	track := &Track{
		KalmanFilter: NewKalmanFilter([]float64{30, 30, 10, 10, 0, 0, 0, 0}),
	}

	vx, vy := track.GetVelocity()

	// Initial velocity should be 0
	if vx != 0 || vy != 0 {
		t.Errorf("Expected initial velocity (0,0), got (%f,%f)", vx, vy)
	}
}

func TestTrack_Update(t *testing.T) {
	detection := ml.Detection{
		Box:        image.Rect(10, 10, 50, 50),
		ClassName:  "person",
		Confidence: 0.8,
		Timestamp:  time.Now(),
	}

	track := &Track{
		ID:         1,
		Detections: []ml.Detection{detection},
		LastUpdate: time.Now(),
		HitStreak:  1,
		Age:        0,
		Class:      detection.ClassName,
		Confidence: detection.Confidence,
		State:      TrackStateTentative,
		Attributes: make(map[string]interface{}),
	}

	// Update with new detection
	newDetection := ml.Detection{
		Box:        image.Rect(15, 15, 55, 55),
		ClassName:  "person",
		Confidence: 0.9,
		Timestamp:  time.Now(),
	}

	track.Update(newDetection)

	if track.HitStreak != 2 {
		t.Errorf("Expected hit streak 2, got %d", track.HitStreak)
	}

	if len(track.Detections) != 2 {
		t.Errorf("Expected 2 detections, got %d", len(track.Detections))
	}
}

func TestTrack_Predict(t *testing.T) {
	detection := ml.Detection{
		Box: image.Rect(10, 10, 50, 50),
	}

	track := &Track{
		KalmanFilter: NewKalmanFilter([]float64{30, 30, 10, 10, 0, 0, 0, 0}),
		Detections:   []ml.Detection{detection},
	}

	predictedBox := track.Predict()

	if predictedBox.Empty() {
		t.Error("Predicted box should not be empty")
	}
}

func TestTrack_IsConfirmed(t *testing.T) {
	track := &Track{State: TrackStateConfirmed}

	if !track.IsConfirmed() {
		t.Error("Track should be confirmed")
	}
}

func TestTrack_IsTentative(t *testing.T) {
	track := &Track{State: TrackStateTentative}

	if !track.IsTentative() {
		t.Error("Track should be tentative")
	}
}

func TestTrack_IsDeleted(t *testing.T) {
	track := &Track{State: TrackStateDeleted}

	if !track.IsDeleted() {
		t.Error("Track should be deleted")
	}
}

func TestTrack_Confirm(t *testing.T) {
	track := &Track{State: TrackStateTentative}

	track.Confirm()

	if track.State != TrackStateConfirmed {
		t.Errorf("Expected state %d, got %d", TrackStateConfirmed, track.State)
	}
}

func TestTrack_Delete(t *testing.T) {
	track := &Track{State: TrackStateConfirmed}

	track.Delete()

	if track.State != TrackStateDeleted {
		t.Errorf("Expected state %d, got %d", TrackStateDeleted, track.State)
	}
}

func TestTrack_GetTimeSinceUpdate(t *testing.T) {
	lastUpdate := time.Now().Add(-5 * time.Second)
	track := &Track{LastUpdate: lastUpdate}

	timeSince := track.GetTimeSinceUpdate()

	if timeSince < 4*time.Second || timeSince > 6*time.Second {
		t.Errorf("Expected time since update ~5s, got %v", timeSince)
	}
}

func TestTrack_GetAverageConfidence(t *testing.T) {
	detections := []ml.Detection{
		{Confidence: 0.8},
		{Confidence: 0.6},
		{Confidence: 1.0},
	}

	track := &Track{Detections: detections}

	avgConf := track.GetAverageConfidence()
	expected := float32(0.8) // (0.8 + 0.6 + 1.0) / 3

	if avgConf != expected {
		t.Errorf("Expected average confidence %f, got %f", expected, avgConf)
	}
}

func TestTrack_Clone(t *testing.T) {
	original := &Track{
		ID:         1,
		HitStreak:  5,
		Age:        10,
		Class:      "person",
		Confidence: 0.8,
		State:      TrackStateConfirmed,
		Attributes: map[string]interface{}{"test": "value"},
	}

	cloned := original.Clone()

	// Check that values are copied
	if cloned.ID != original.ID {
		t.Errorf("Clone ID mismatch: %d vs %d", cloned.ID, original.ID)
	}

	if cloned.Class != original.Class {
		t.Errorf("Clone class mismatch: %s vs %s", cloned.Class, original.Class)
	}

	// Modify original and check clone is unaffected
	original.Attributes["test"] = "modified"

	if cloned.Attributes["test"] == "modified" {
		t.Error("Clone should be independent of original")
	}
}

func TestTrackState_String(t *testing.T) {
	tests := []struct {
		state    TrackState
		expected string
	}{
		{TrackStateTentative, "tentative"},
		{TrackStateConfirmed, "confirmed"},
		{TrackStateDeleted, "deleted"},
		{TrackState(999), "unknown"},
	}

	for _, test := range tests {
		result := test.state.String()
		if result != test.expected {
			t.Errorf("State %d: expected '%s', got '%s'", test.state, test.expected, result)
		}
	}
}

// Benchmark tests
func BenchmarkTrack_GetBoundingBox(b *testing.B) {
	detection := ml.Detection{Box: image.Rect(10, 10, 50, 50)}
	track := &Track{Detections: []ml.Detection{detection}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		track.GetBoundingBox()
	}
}

func BenchmarkTrack_GetCenter(b *testing.B) {
	detection := ml.Detection{Box: image.Rect(10, 10, 50, 50)}
	track := &Track{Detections: []ml.Detection{detection}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		track.GetCenter()
	}
}

func BenchmarkTrack_Update(b *testing.B) {
	detection := ml.Detection{
		Box:        image.Rect(10, 10, 50, 50),
		ClassName:  "person",
		Confidence: 0.8,
		Timestamp:  time.Now(),
	}

	track := &Track{
		ID:         1,
		Detections: []ml.Detection{detection},
		LastUpdate: time.Now(),
		HitStreak:  1,
		Age:        0,
		Class:      detection.ClassName,
		Confidence: detection.Confidence,
		State:      TrackStateTentative,
		Attributes: make(map[string]interface{}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		track.Update(detection)
	}
}
