package tracking

import (
	"math"
)

// KalmanFilter implements a Kalman filter for object tracking
type KalmanFilter struct {
	// State vector: [x, y, w, h, vx, vy, vw, vh]
	// x, y: center position
	// w, h: width and height
	// vx, vy: velocity in x and y
	// vw, vh: velocity of width and height
	state      []float64
	covariance [][]float64

	// Transition matrix (8x8)
	F [][]float64

	// Measurement matrix (4x8) - we only measure position and size
	H [][]float64

	// Process noise covariance (8x8)
	Q [][]float64

	// Measurement noise covariance (4x4)
	R [][]float64

	// Identity matrix (8x8)
	I [][]float64
}

// NewKalmanFilter creates a new Kalman filter for tracking
func NewKalmanFilter(initialBox []float64) *KalmanFilter {
	kf := &KalmanFilter{
		state:      make([]float64, 8),
		covariance: make([][]float64, 8),
		F:          make([][]float64, 8),
		H:          make([][]float64, 4),
		Q:          make([][]float64, 8),
		R:          make([][]float64, 4),
		I:          make([][]float64, 8),
	}

	// Initialize covariance matrix
	for i := range kf.covariance {
		kf.covariance[i] = make([]float64, 8)
	}

	// Initialize matrices
	kf.initializeMatrices()

	// Initialize state with initial bounding box
	// Convert [x1, y1, x2, y2] to [cx, cy, w, h, 0, 0, 0, 0]
	if len(initialBox) >= 4 {
		x1, y1, x2, y2 := initialBox[0], initialBox[1], initialBox[2], initialBox[3]
		kf.state[0] = (x1 + x2) / 2 // center x
		kf.state[1] = (y1 + y2) / 2 // center y
		kf.state[2] = x2 - x1       // width
		kf.state[3] = y2 - y1       // height
		// velocities start at 0
	}

	return kf
}

// initializeMatrices sets up the Kalman filter matrices
func (kf *KalmanFilter) initializeMatrices() {
	// Identity matrix
	for i := range kf.I {
		kf.I[i] = make([]float64, 8)
		kf.I[i][i] = 1.0
	}

	// Transition matrix F (constant velocity model)
	for i := range kf.F {
		kf.F[i] = make([]float64, 8)
		kf.F[i][i] = 1.0
	}
	// Position depends on velocity
	kf.F[0][4] = 1.0 // x depends on vx
	kf.F[1][5] = 1.0 // y depends on vy
	kf.F[2][6] = 1.0 // w depends on vw
	kf.F[3][7] = 1.0 // h depends on vh

	// Measurement matrix H (we measure position and size)
	for i := range kf.H {
		kf.H[i] = make([]float64, 8)
	}
	kf.H[0][0] = 1.0 // measure x
	kf.H[1][1] = 1.0 // measure y
	kf.H[2][2] = 1.0 // measure w
	kf.H[3][3] = 1.0 // measure h

	// Process noise covariance Q
	for i := range kf.Q {
		kf.Q[i] = make([]float64, 8)
	}
	// Add uncertainty to velocity components
	kf.Q[4][4] = 0.01 // vx process noise
	kf.Q[5][5] = 0.01 // vy process noise
	kf.Q[6][6] = 0.01 // vw process noise
	kf.Q[7][7] = 0.01 // vh process noise

	// Measurement noise covariance R
	for i := range kf.R {
		kf.R[i] = make([]float64, 4)
		kf.R[i][i] = 0.1 // measurement noise
	}

	// Initial covariance (high uncertainty)
	for i := range kf.covariance {
		kf.covariance[i][i] = 100.0
	}
}

// Predict predicts the next state
func (kf *KalmanFilter) Predict() {
	// Predict state: x_k = F * x_{k-1}
	kf.state = kf.matrixVectorMultiply(kf.F, kf.state)

	// Predict covariance: P_k = F * P_{k-1} * F^T + Q
	FT := kf.matrixTranspose(kf.F)
	temp := kf.matrixMultiply(kf.F, kf.covariance)
	kf.covariance = kf.matrixMultiply(temp, FT)
	kf.covariance = kf.matrixAdd(kf.covariance, kf.Q)
}

// Update updates the filter with a new measurement
func (kf *KalmanFilter) Update(measurement []float64) {
	if len(measurement) < 4 {
		return
	}

	// Simple update: directly update position and size with some smoothing
	// This is a simplified Kalman filter update
	alpha := 0.7 // smoothing factor

	// Convert measurement to center coordinates if needed
	// Assuming measurement is [x1, y1, x2, y2]
	measCx := (measurement[0] + measurement[2]) / 2
	measCy := (measurement[1] + measurement[3]) / 2
	measW := measurement[2] - measurement[0]
	measH := measurement[3] - measurement[1]

	// Update state with smoothing
	kf.state[0] = alpha*measCx + (1-alpha)*kf.state[0] // center x
	kf.state[1] = alpha*measCy + (1-alpha)*kf.state[1] // center y
	kf.state[2] = alpha*measW + (1-alpha)*kf.state[2]  // width
	kf.state[3] = alpha*measH + (1-alpha)*kf.state[3]  // height

	// Velocities are updated based on position change
	kf.state[4] = (measCx - kf.state[0]) * 0.1 // velocity x
	kf.state[5] = (measCy - kf.state[1]) * 0.1 // velocity y
}

// GetState returns the current state
func (kf *KalmanFilter) GetState() []float64 {
	state := make([]float64, len(kf.state))
	copy(state, kf.state)
	return state
}

// GetBoundingBox returns the predicted bounding box
func (kf *KalmanFilter) GetBoundingBox() []float64 {
	if len(kf.state) < 4 {
		return []float64{0, 0, 0, 0}
	}

	cx, cy, w, h := kf.state[0], kf.state[1], kf.state[2], kf.state[3]

	// Convert from center format to corner format
	x1 := cx - w/2
	y1 := cy - h/2
	x2 := cx + w/2
	y2 := cy + h/2

	return []float64{x1, y1, x2, y2}
}

// Clone creates a copy of the Kalman filter
func (kf *KalmanFilter) Clone() *KalmanFilter {
	clone := &KalmanFilter{
		state:      make([]float64, len(kf.state)),
		covariance: make([][]float64, len(kf.covariance)),
		F:          make([][]float64, len(kf.F)),
		H:          make([][]float64, len(kf.H)),
		Q:          make([][]float64, len(kf.Q)),
		R:          make([][]float64, len(kf.R)),
		I:          make([][]float64, len(kf.I)),
	}

	copy(clone.state, kf.state)

	for i := range kf.covariance {
		clone.covariance[i] = make([]float64, len(kf.covariance[i]))
		copy(clone.covariance[i], kf.covariance[i])
	}

	for i := range kf.F {
		clone.F[i] = make([]float64, len(kf.F[i]))
		copy(clone.F[i], kf.F[i])
	}

	for i := range kf.H {
		clone.H[i] = make([]float64, len(kf.H[i]))
		copy(clone.H[i], kf.H[i])
	}

	for i := range kf.Q {
		clone.Q[i] = make([]float64, len(kf.Q[i]))
		copy(clone.Q[i], kf.Q[i])
	}

	for i := range kf.R {
		clone.R[i] = make([]float64, len(kf.R[i]))
		copy(clone.R[i], kf.R[i])
	}

	for i := range kf.I {
		clone.I[i] = make([]float64, len(kf.I[i]))
		copy(clone.I[i], kf.I[i])
	}

	return clone
}

// Matrix operations helper functions

func (kf *KalmanFilter) matrixMultiply(A, B [][]float64) [][]float64 {
	rowsA, colsA := len(A), len(A[0])
	rowsB, colsB := len(B), len(B[0])

	if colsA != rowsB {
		return nil
	}

	result := make([][]float64, rowsA)
	for i := range result {
		result[i] = make([]float64, colsB)
	}

	for i := 0; i < rowsA; i++ {
		for j := 0; j < colsB; j++ {
			for k := 0; k < colsA; k++ {
				result[i][j] += A[i][k] * B[k][j]
			}
		}
	}

	return result
}

func (kf *KalmanFilter) matrixVectorMultiply(A [][]float64, v []float64) []float64 {
	rows, cols := len(A), len(A[0])

	if cols != len(v) {
		return nil
	}

	result := make([]float64, rows)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			result[i] += A[i][j] * v[j]
		}
	}

	return result
}

func (kf *KalmanFilter) matrixTranspose(A [][]float64) [][]float64 {
	rows, cols := len(A), len(A[0])
	result := make([][]float64, cols)

	for i := range result {
		result[i] = make([]float64, rows)
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			result[j][i] = A[i][j]
		}
	}

	return result
}

func (kf *KalmanFilter) matrixAdd(A, B [][]float64) [][]float64 {
	rows, cols := len(A), len(A[0])
	result := make([][]float64, rows)

	for i := range result {
		result[i] = make([]float64, cols)
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			result[i][j] = A[i][j] + B[i][j]
		}
	}

	return result
}

func (kf *KalmanFilter) matrixSubtract(A, B [][]float64) [][]float64 {
	rows, cols := len(A), len(A[0])
	result := make([][]float64, rows)

	for i := range result {
		result[i] = make([]float64, cols)
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			result[i][j] = A[i][j] - B[i][j]
		}
	}

	return result
}

func (kf *KalmanFilter) matrixInverse(A [][]float64) [][]float64 {
	// Simple 2x2 or 4x4 matrix inverse (for our use case)
	n := len(A)
	if n == 2 {
		return kf.inverse2x2(A)
	} else if n == 4 {
		return kf.inverse4x4(A)
	}

	// For larger matrices, return identity (simplified)
	return kf.I[:n]
}

func (kf *KalmanFilter) inverse2x2(A [][]float64) [][]float64 {
	a, b, c, d := A[0][0], A[0][1], A[1][0], A[1][1]
	det := a*d - b*c

	if math.Abs(det) < 1e-10 {
		return kf.I[:2]
	}

	invDet := 1.0 / det
	return [][]float64{
		{d * invDet, -b * invDet},
		{-c * invDet, a * invDet},
	}
}

func (kf *KalmanFilter) inverse4x4(A [][]float64) [][]float64 {
	// Simplified 4x4 inverse - in practice, you'd use a proper library
	// For now, return identity as approximation
	return kf.I[:4]
}
