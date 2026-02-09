package safety

import (
	"fmt"
	"testing"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/safety"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/types"
)

func TestSafetyDashboard(t *testing.T) {
	dashboard := NewDashboard(80, 24)

	// Create test safety status
	status := &safety.SafetyStatus{
		IsSafe:        true,
		SafetyEnabled: true,
		EmergencyMode: false,
		ActiveEvents: []safety.SafetyEvent{
			{
				Timestamp: time.Now(),
				Level:     "warning",
				Type:      "battery",
				Message:   "Battery level low",
			},
		},
		ConfigLevel: safety.SafetyLevelNormal,
	}

	// Create test state
	state := &types.State{
		H:     150,
		Bat:   25,
		Temph: 45,
		Time:  300,
		Vgx:   30,
		Vgy:   20,
		Vgz:   5,
		Tof:   80,
	}

	// Update dashboard
	dashboard.UpdateSafetyStatus(status, state)

	// Test rendering
	rendered := dashboard.Render()
	if rendered == "" {
		t.Error("Dashboard render returned empty string")
	}

	// Test toggling features
	dashboard.ToggleDetails()
	dashboard.ToggleLimits()
	dashboard.ToggleEvents()

	rendered2 := dashboard.Render()
	if rendered2 == "" {
		t.Error("Dashboard render after toggling returned empty string")
	}

	fmt.Println("Safety dashboard test passed")
}

func TestEmergencyVisualization(t *testing.T) {
	emergency := NewEmergencyVisualization(60, 20)

	// Test starting emergency
	emergency.StartEmergency("battery_low", "critical")

	if !emergency.IsActive {
		t.Error("Emergency visualization should be active after StartEmergency")
	}

	if emergency.EmergencyType != "battery_low" {
		t.Error("Emergency type not set correctly")
	}

	// Test rendering
	rendered := emergency.Render()
	if rendered == "" {
		t.Error("Emergency render returned empty string")
	}

	// Test progress update
	emergency.UpdateProgress(3 * time.Second)

	// Test stopping emergency
	emergency.StopEmergency()

	if emergency.IsActive {
		t.Error("Emergency visualization should be inactive after StopEmergency")
	}

	// Test rendering inactive state
	rendered2 := emergency.Render()
	if rendered2 == "" {
		t.Error("Emergency render (inactive) returned empty string")
	}

	fmt.Println("Emergency visualization test passed")
}

func TestProximityWarning(t *testing.T) {
	proximity := NewProximityWarning(70, 20)

	// Create test state
	state := &types.State{
		H:   120,
		Tof: 60,
		Vgx: 10,
		Vgy: 5,
		Vgz: 2,
	}

	proximity.UpdateState(state)

	// Add test obstacles
	obstacles := []Obstacle{
		{
			ID:       "obs1",
			X:        30,
			Y:        40,
			Z:        20,
			Distance: 50,
			Type:     "person",
			Size:     50,
		},
		{
			ID:       "obs2",
			X:        -60,
			Y:        80,
			Z:        10,
			Distance: 120,
			Type:     "vehicle",
			Size:     200,
		},
	}

	for _, obs := range obstacles {
		proximity.AddObstacle(obs)
	}

	if len(proximity.Obstacles) != 2 {
		t.Errorf("Expected 2 obstacles, got %d", len(proximity.Obstacles))
	}

	// Test rendering
	rendered := proximity.Render()
	if rendered == "" {
		t.Error("Proximity warning render returned empty string")
	}

	// Test toggling features
	proximity.ToggleRadar()
	proximity.ToggleWarnings()
	proximity.ToggleDistances()

	rendered2 := proximity.Render()
	if rendered2 == "" {
		t.Error("Proximity warning render after toggling returned empty string")
	}

	// Test obstacle removal
	proximity.RemoveObstacle("obs1")
	if len(proximity.Obstacles) != 1 {
		t.Errorf("Expected 1 obstacle after removal, got %d", len(proximity.Obstacles))
	}

	// Test clearing obstacles
	proximity.ClearObstacles()
	if len(proximity.Obstacles) != 0 {
		t.Errorf("Expected 0 obstacles after clear, got %d", len(proximity.Obstacles))
	}

	fmt.Println("Proximity warning test passed")
}

func TestSafetyManager(t *testing.T) {
	manager := NewManager(80, 24)

	// Create test safety status
	status := &safety.SafetyStatus{
		IsSafe:        true,
		SafetyEnabled: true,
		EmergencyMode: false,
		ActiveEvents:  make([]safety.SafetyEvent, 0),
		ConfigLevel:   safety.SafetyLevelNormal,
	}

	// Create test state
	state := &types.State{
		H:     100,
		Bat:   50,
		Temph: 35,
		Time:  200,
	}

	// Update manager
	manager.UpdateSafety(status, state)

	// Test rendering in different modes
	modes := []string{"dashboard", "emergency", "proximity", "combined"}
	for _, mode := range modes {
		manager.CurrentMode = mode
		rendered := manager.Render()

		if rendered == "" && manager.ShowSafety {
			t.Errorf("Manager render returned empty string for mode %s", mode)
		}
	}

	// Test emergency handling
	manager.StartEmergency("connection_lost", "critical")

	if !manager.IsEmergencyActive() {
		t.Error("Manager should report emergency as active")
	}

	if manager.CurrentMode != "emergency" {
		t.Error("Manager should switch to emergency mode")
	}

	// Test emergency stop
	manager.StopEmergency()

	if manager.IsEmergencyActive() {
		t.Error("Manager should report emergency as inactive after stop")
	}

	// Test mode toggling
	initialMode := manager.GetCurrentMode()
	manager.ToggleMode()

	if manager.GetCurrentMode() == initialMode {
		t.Error("ToggleMode should change the current mode")
	}

	// Test safety toggle
	initialShow := manager.ShowSafety
	manager.ToggleSafety()

	if manager.ShowSafety == initialShow {
		t.Error("ToggleSafety should change ShowSafety")
	}

	// Test key handling
	manager.HandleKey("f6")
	manager.HandleKey("f7")
	manager.HandleKey("f8")
	manager.HandleKey("f9") // This should start an emergency

	if !manager.IsEmergencyActive() {
		t.Error("F9 key should start emergency procedure")
	}

	// Test obstacle addition
	obstacle := Obstacle{
		ID:       "test_obstacle",
		X:        50,
		Y:        30,
		Z:        10,
		Distance: 60,
		Type:     "tree",
	}

	manager.AddObstacle(obstacle)

	// Test clearing all
	manager.ClearAll()

	if manager.IsEmergencyActive() {
		t.Error("ClearAll should stop any active emergency")
	}

	fmt.Println("Safety manager test passed")
}

func TestSafetyEventHandling(t *testing.T) {
	manager := NewManager(80, 24)

	// Test with critical safety event
	status := &safety.SafetyStatus{
		IsSafe:        false,
		SafetyEnabled: true,
		EmergencyMode: false,
		ActiveEvents: []safety.SafetyEvent{
			{
				Timestamp: time.Now(),
				Level:     "critical",
				Type:      "battery",
				Message:   "Battery critically low",
			},
		},
		ConfigLevel: safety.SafetyLevelNormal,
	}

	state := &types.State{
		H:   100,
		Bat: 10,
	}

	manager.UpdateSafety(status, state)

	// Critical event should trigger emergency mode
	if !manager.IsEmergencyActive() {
		t.Error("Critical safety event should trigger emergency mode")
	}

	fmt.Println("Safety event handling test passed")
}

func TestProximityIntegration(t *testing.T) {
	manager := NewManager(80, 24)

	// Add multiple obstacles at different distances
	obstacles := []Obstacle{
		{ID: "close", Distance: 30, Type: "person"},
		{ID: "medium", Distance: 80, Type: "vehicle"},
		{ID: "far", Distance: 150, Type: "building"},
	}

	for _, obs := range obstacles {
		manager.AddObstacle(obs)
	}

	// Switch to proximity mode
	manager.CurrentMode = "proximity"
	rendered := manager.Render()

	if rendered == "" {
		t.Error("Proximity mode render returned empty string")
	}

	// Test proximity warnings
	manager.HandleKey("f10") // Toggle radar
	manager.HandleKey("f11") // Toggle warnings
	manager.HandleKey("f12") // Toggle distances

	rendered2 := manager.Render()
	if rendered2 == "" {
		t.Error("Proximity mode render after toggling returned empty string")
	}

	fmt.Println("Proximity integration test passed")
}
