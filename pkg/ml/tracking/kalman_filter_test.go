package tracking

import (
	"math"
	"testing"
)

func TestNewKalmanFilter(t *testing.T) {
	// Initialize with corner coordinates [x1, y1, x2, y2]
	// Box: [295, 215, 345, 265] -> center [320, 240], size [50, 50]
	measurement := []float64{295, 215, 345, 265}
	kf := NewKalmanFilter(measurement)

	if kf == nil {
		t.Fatal("KalmanFilter should not be nil")
	}

	// Check that state vector has correct dimensions
	if len(kf.state) != 8 {
		t.Errorf("Expected state vector length 8, got %d", len(kf.state))
	}

	// Check initial state (should be center coordinates and size)
	expectedCenterX := 320.0
	expectedCenterY := 240.0
	expectedWidth := 50.0
	expectedHeight := 50.0

	if math.Abs(kf.state[0]-expectedCenterX) > 0.001 {
		t.Errorf("State[0] should be %f, got %f", expectedCenterX, kf.state[0])
	}
	if math.Abs(kf.state[1]-expectedCenterY) > 0.001 {
		t.Errorf("State[1] should be %f, got %f", expectedCenterY, kf.state[1])
	}
	if math.Abs(kf.state[2]-expectedWidth) > 0.001 {
		t.Errorf("State[2] should be %f, got %f", expectedWidth, kf.state[2])
	}
	if math.Abs(kf.state[3]-expectedHeight) > 0.001 {
		t.Errorf("State[3] should be %f, got %f", expectedHeight, kf.state[3])
	}
}

func TestKalmanFilter_Predict(t *testing.T) {
	measurement := []float64{295, 215, 345, 265}
	kf := NewKalmanFilter(measurement)

	// Store initial state
	initialState := make([]float64, 8)
	copy(initialState, kf.state)

	// Predict next state
	kf.Predict()

	// State should have changed (due to velocity - though initially zero)
	// With zero velocity, position shouldn't change much due to process noise
	if math.Abs(kf.state[0]-initialState[0]) > 10 {
		t.Errorf("Position shouldn't change dramatically with zero velocity, got %f -> %f", initialState[0], kf.state[0])
	}
}

func TestKalmanFilter_Update(t *testing.T) {
	measurement := []float64{295, 215, 345, 265}
	kf := NewKalmanFilter(measurement)

	// Predict first
	kf.Predict()

	// Update with new measurement (corner coordinates)
	newMeasurement := []float64{300, 220, 350, 270} // center [325, 245], size [50, 50]
	kf.Update(newMeasurement)

	// State should be closer to new measurement
	if math.Abs(float64(kf.state[0]-325)) > 10 {
		t.Errorf("State should be close to measurement center, got %f", kf.state[0])
	}
}

func TestKalmanFilter_GetState(t *testing.T) {
	measurement := []float64{295, 215, 345, 265}
	kf := NewKalmanFilter(measurement)

	state := kf.GetState()

	if len(state) != 8 {
		t.Errorf("Expected state length 8, got %d", len(state))
	}

	// Check that it returns a copy
	state[0] = 999
	if kf.state[0] == 999 {
		t.Error("GetState should return a copy, not reference")
	}
}

func TestKalmanFilter_GetBoundingBox(t *testing.T) {
	measurement := []float64{295, 215, 345, 265}
	kf := NewKalmanFilter(measurement)

	box := kf.GetBoundingBox()

	// GetBoundingBox returns corner coordinates [x1, y1, x2, y2]
	expectedX1 := 295.0
	expectedY1 := 215.0
	expectedX2 := 345.0
	expectedY2 := 265.0

	if math.Abs(box[0]-expectedX1) > 0.001 {
		t.Errorf("Expected x1 %f, got %f", expectedX1, box[0])
	}

	if math.Abs(box[1]-expectedY1) > 0.001 {
		t.Errorf("Expected y1 %f, got %f", expectedY1, box[1])
	}

	if math.Abs(box[2]-expectedX2) > 0.001 {
		t.Errorf("Expected x2 %f, got %f", expectedX2, box[2])
	}

	if math.Abs(box[3]-expectedY2) > 0.001 {
		t.Errorf("Expected y2 %f, got %f", expectedY2, box[3])
	}
}

func TestKalmanFilter_GetVelocity(t *testing.T) {
	measurement := []float64{295, 215, 345, 265}
	kf := NewKalmanFilter(measurement)

	// Initially velocity should be 0
	state := kf.GetState()
	vx := state[4] // velocity X
	vy := state[5] // velocity Y

	if vx != 0 {
		t.Errorf("Expected initial velocity X 0, got %f", vx)
	}

	if vy != 0 {
		t.Errorf("Expected initial velocity Y 0, got %f", vy)
	}
}

func TestKalmanFilter_ConstantVelocityModel(t *testing.T) {
	// Test that filter maintains constant velocity
	measurement := []float64{80, 80, 120, 120} // center [100, 100], size [40, 40]
	kf := NewKalmanFilter(measurement)

	// Manually set velocity to test prediction
	kf.state[4] = 10.0 // vx
	kf.state[5] = 5.0  // vy

	// Predict several steps
	for i := 0; i < 5; i++ {
		kf.Predict()
	}

	// Position should have moved according to velocity
	expectedX := 100.0 + 10.0*5 // 150
	expectedY := 100.0 + 5.0*5  // 125

	if math.Abs(kf.state[0]-expectedX) > 1 {
		t.Errorf("Expected X ~%f, got %f", expectedX, kf.state[0])
	}

	if math.Abs(kf.state[1]-expectedY) > 1 {
		t.Errorf("Expected Y ~%f, got %f", expectedY, kf.state[1])
	}
}

func TestKalmanFilter_NoiseHandling(t *testing.T) {
	measurement := []float64{295, 215, 345, 265}
	kf := NewKalmanFilter(measurement)

	// Add some noise to measurements (corner coordinates)
	measurements := [][]float64{
		{297, 217, 347, 267}, // +2, +2
		{293, 213, 343, 263}, // -4, -4
		{300, 220, 350, 270}, // +7, +7
		{290, 210, 340, 260}, // -10, -10
		{295, 215, 345, 265}, // back to original
	}

	for _, m := range measurements {
		kf.Predict()
		kf.Update(m)
	}

	// Final state should be close to original despite noise
	if math.Abs(float64(kf.state[0]-320)) > 10 {
		t.Errorf("State should be close to original despite noise, got %f", kf.state[0])
	}

	if math.Abs(float64(kf.state[1]-240)) > 10 {
		t.Errorf("State should be close to original despite noise, got %f", kf.state[1])
	}
}

func TestKalmanFilter_MissingMeasurements(t *testing.T) {
	measurement := []float64{295, 215, 345, 265}
	kf := NewKalmanFilter(measurement)

	// Manually set velocity
	kf.state[4] = 10.0 // vx
	kf.state[5] = 5.0  // vy

	// Predict without updates for several steps
	for i := 0; i < 10; i++ {
		kf.Predict()
	}

	// Should still have reasonable position based on velocity
	expectedX := 320.0 + 10.0*10 // 420
	expectedY := 240.0 + 5.0*10  // 290

	if math.Abs(kf.state[0]-expectedX) > 20 {
		t.Errorf("Expected X ~%f, got %f", expectedX, kf.state[0])
	}

	if math.Abs(kf.state[1]-expectedY) > 20 {
		t.Errorf("Expected Y ~%f, got %f", expectedY, kf.state[1])
	}
}

// Benchmark tests
func BenchmarkKalmanFilter_Predict(b *testing.B) {
	measurement := []float64{295, 215, 345, 265}
	kf := NewKalmanFilter(measurement)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kf.Predict()
	}
}

func BenchmarkKalmanFilter_Update(b *testing.B) {
	measurement := []float64{295, 215, 345, 265}
	kf := NewKalmanFilter(measurement)
	newMeasurement := []float64{300, 220, 350, 270}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kf.Predict()
		kf.Update(newMeasurement)
	}
}

func BenchmarkKalmanFilter_GetBoundingBox(b *testing.B) {
	measurement := []float64{295, 215, 345, 265}
	kf := NewKalmanFilter(measurement)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kf.GetBoundingBox()
	}
}

func BenchmarkKalmanFilter_GetVelocity(b *testing.B) {
	measurement := []float64{295, 215, 345, 265}
	kf := NewKalmanFilter(measurement)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state := kf.GetState()
		_ = state[4] // velocity X
		_ = state[5] // velocity Y
	}
}
