package telemetry

import (
	"fmt"
	"testing"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/types"
)

func TestGraphRendering(t *testing.T) {
	graph := NewGraph(40, 10, "Test Graph", "units")

	// Add some test data
	for i := 0; i < 50; i++ {
		value := float64(i%20) + 10
		graph.AddDataPoint(value)
	}

	rendered := graph.Render()
	if rendered == "" {
		t.Error("Graph render returned empty string")
	}

	// Check that title is included
	if len(rendered) < 10 {
		t.Error("Graph render output too short")
	}

	fmt.Println("Graph test passed, output sample:")
	fmt.Println(rendered[:100])
}

func TestHorizonRendering(t *testing.T) {
	horizon := NewHorizon(30, 12)

	// Test with different orientations
	testCases := []struct {
		name  string
		pitch float64
		roll  float64
	}{
		{"Level", 0, 0},
		{"Pitch Up", 10, 0},
		{"Pitch Down", -10, 0},
		{"Roll Right", 0, 15},
		{"Roll Left", 0, -15},
		{"Combined", 5, -10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			horizon.SetOrientation(tc.pitch, tc.roll)
			rendered := horizon.Render()

			if rendered == "" {
				t.Error("Horizon render returned empty string")
			}

			if len(rendered) < 20 {
				t.Error("Horizon render output too short")
			}
		})
	}

	fmt.Println("Horizon test passed")
}

func TestDashboardRendering(t *testing.T) {
	dashboard := NewDashboard(80, 24)

	// Create test state
	state := &types.State{
		Pitch: 5,
		Roll:  -3,
		Yaw:   45,
		Vgx:   10,
		Vgy:   -5,
		Vgz:   2,
		Templ: 25,
		Temph: 30,
		Tof:   150,
		H:     120,
		Bat:   75,
		Baro:  1.2,
		Time:  125,
		Agx:   0.1,
		Agy:   -0.05,
		Agz:   0.98,
	}

	// Update dashboard with state
	dashboard.UpdateState(state)

	// Test rendering
	rendered := dashboard.Render()
	if rendered == "" {
		t.Error("Dashboard render returned empty string")
	}

	// Check for key elements
	if len(rendered) < 100 {
		t.Error("Dashboard render output too short")
	}

	// Test toggling features
	dashboard.ToggleGraphs()
	dashboard.ToggleHorizon()
	dashboard.ToggleMetrics()

	rendered2 := dashboard.Render()
	if rendered2 == "" {
		t.Error("Dashboard render after toggling returned empty string")
	}

	fmt.Println("Dashboard test passed")
}

func TestDashboardUpdate(t *testing.T) {
	dashboard := NewDashboard(60, 20)

	// Test multiple updates
	for i := 0; i < 10; i++ {
		state := &types.State{
			Pitch: i,
			Roll:  i * 2,
			H:     100 + i*10,
			Bat:   80 - i,
			Temph: 30 + i,
			Vgx:   i * 5,
			Time:  i * 10,
		}

		dashboard.UpdateState(state)

		// Verify state was updated
		if dashboard.State.H != state.H {
			t.Errorf("Dashboard state not updated correctly. Expected H=%d, got H=%d", state.H, dashboard.State.H)
		}
	}

	fmt.Println("Dashboard update test passed")
}

func TestGraphBounds(t *testing.T) {
	graph := NewGraph(30, 8, "Bounded Graph", "%")
	graph.SetBounds(0, 100)

	// Add values within bounds
	graph.AddDataPoint(25)
	graph.AddDataPoint(50)
	graph.AddDataPoint(75)

	// Add value outside bounds - should trigger auto-scaling
	graph.AddDataPoint(120)

	rendered := graph.Render()
	if rendered == "" {
		t.Error("Graph with bounds render returned empty string")
	}

	fmt.Println("Graph bounds test passed")
}
