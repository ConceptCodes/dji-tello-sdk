package shared

import (
	"encoding/json"
	"testing"
)

func TestTelloStateJSON(t *testing.T) {
	// Test JSON serialization
	state := &TelloState{
		Pitch: 10, Roll: -5, Yaw: 180,
		Vgx: 20, Vgy: 30, Vgz: 40,
		Templ: 20, Temph: 30,
		Tof: 300, H: 100, Bat: 85,
		Baro: 1013.25, Time: 120,
		Agx: 0.1, Agy: 0.2, Agz: 0.3,
	}
	
	// Marshal to JSON
	data, err := json.Marshal(state)
	if err != nil {
		t.Errorf("Expected no error marshaling state to JSON, got %v", err)
	}
	
	// Unmarshal back
	var unmarshaledState TelloState
	err = json.Unmarshal(data, &unmarshaledState)
	if err != nil {
		t.Errorf("Expected no error unmarshaling JSON to state, got %v", err)
	}
	
	// Compare fields
	if unmarshaledState.Pitch != state.Pitch {
		t.Errorf("Expected Pitch %d, got %d", state.Pitch, unmarshaledState.Pitch)
	}
	if unmarshaledState.Roll != state.Roll {
		t.Errorf("Expected Roll %d, got %d", state.Roll, unmarshaledState.Roll)
	}
	if unmarshaledState.Yaw != state.Yaw {
		t.Errorf("Expected Yaw %d, got %d", state.Yaw, unmarshaledState.Yaw)
	}
	if unmarshaledState.Vgx != state.Vgx {
		t.Errorf("Expected Vgx %d, got %d", state.Vgx, unmarshaledState.Vgx)
	}
	if unmarshaledState.Vgy != state.Vgy {
		t.Errorf("Expected Vgy %d, got %d", state.Vgy, unmarshaledState.Vgy)
	}
	if unmarshaledState.Vgz != state.Vgz {
		t.Errorf("Expected Vgz %d, got %d", state.Vgz, unmarshaledState.Vgz)
	}
	if unmarshaledState.Templ != state.Templ {
		t.Errorf("Expected Templ %d, got %d", state.Templ, unmarshaledState.Templ)
	}
	if unmarshaledState.Temph != state.Temph {
		t.Errorf("Expected Temph %d, got %d", state.Temph, unmarshaledState.Temph)
	}
	if unmarshaledState.Tof != state.Tof {
		t.Errorf("Expected Tof %d, got %d", state.Tof, unmarshaledState.Tof)
	}
	if unmarshaledState.H != state.H {
		t.Errorf("Expected H %d, got %d", state.H, unmarshaledState.H)
	}
	if unmarshaledState.Bat != state.Bat {
		t.Errorf("Expected Bat %d, got %d", state.Bat, unmarshaledState.Bat)
	}
	if unmarshaledState.Baro != state.Baro {
		t.Errorf("Expected Baro %f, got %f", state.Baro, unmarshaledState.Baro)
	}
	if unmarshaledState.Time != state.Time {
		t.Errorf("Expected Time %d, got %d", state.Time, unmarshaledState.Time)
	}
	if unmarshaledState.Agx != state.Agx {
		t.Errorf("Expected Agx %f, got %f", state.Agx, unmarshaledState.Agx)
	}
	if unmarshaledState.Agy != state.Agy {
		t.Errorf("Expected Agy %f, got %f", state.Agy, unmarshaledState.Agy)
	}
	if unmarshaledState.Agz != state.Agz {
		t.Errorf("Expected Agz %f, got %f", state.Agz, unmarshaledState.Agz)
	}
}

func TestTelloStateFields(t *testing.T) {
	// Test field types and default values
	state := &TelloState{}
	
	// Test default values (should be zero values)
	if state.Pitch != 0 {
		t.Errorf("Expected default Pitch 0, got %d", state.Pitch)
	}
	if state.Roll != 0 {
		t.Errorf("Expected default Roll 0, got %d", state.Roll)
	}
	if state.Yaw != 0 {
		t.Errorf("Expected default Yaw 0, got %d", state.Yaw)
	}
	if state.Vgx != 0 {
		t.Errorf("Expected default Vgx 0, got %d", state.Vgx)
	}
	if state.Vgy != 0 {
		t.Errorf("Expected default Vgy 0, got %d", state.Vgy)
	}
	if state.Vgz != 0 {
		t.Errorf("Expected default Vgz 0, got %d", state.Vgz)
	}
	if state.Templ != 0 {
		t.Errorf("Expected default Templ 0, got %d", state.Templ)
	}
	if state.Temph != 0 {
		t.Errorf("Expected default Temph 0, got %d", state.Temph)
	}
	if state.Tof != 0 {
		t.Errorf("Expected default Tof 0, got %d", state.Tof)
	}
	if state.H != 0 {
		t.Errorf("Expected default H 0, got %d", state.H)
	}
	if state.Bat != 0 {
		t.Errorf("Expected default Bat 0, got %d", state.Bat)
	}
	if state.Baro != 0.0 {
		t.Errorf("Expected default Baro 0.0, got %f", state.Baro)
	}
	if state.Time != 0 {
		t.Errorf("Expected default Time 0, got %d", state.Time)
	}
	if state.Agx != 0.0 {
		t.Errorf("Expected default Agx 0.0, got %f", state.Agx)
	}
	if state.Agy != 0.0 {
		t.Errorf("Expected default Agy 0.0, got %f", state.Agy)
	}
	if state.Agz != 0.0 {
		t.Errorf("Expected default Agz 0.0, got %f", state.Agz)
	}
	
	// Test field assignments
	state.Pitch = 10
	state.Roll = -5
	state.Yaw = 180
	state.Vgx = 20
	state.Vgy = 30
	state.Vgz = 40
	state.Templ = 25
	state.Temph = 35
	state.Tof = 500
	state.H = 150
	state.Bat = 90
	state.Baro = 1013.25
	state.Time = 300
	state.Agx = 0.1
	state.Agy = -0.2
	state.Agz = 9.8
	
	// Verify assignments
	if state.Pitch != 10 {
		t.Errorf("Expected Pitch 10, got %d", state.Pitch)
	}
	if state.Roll != -5 {
		t.Errorf("Expected Roll -5, got %d", state.Roll)
	}
	if state.Yaw != 180 {
		t.Errorf("Expected Yaw 180, got %d", state.Yaw)
	}
	if state.Vgx != 20 {
		t.Errorf("Expected Vgx 20, got %d", state.Vgx)
	}
	if state.Vgy != 30 {
		t.Errorf("Expected Vgy 30, got %d", state.Vgy)
	}
	if state.Vgz != 40 {
		t.Errorf("Expected Vgz 40, got %d", state.Vgz)
	}
	if state.Templ != 25 {
		t.Errorf("Expected Templ 25, got %d", state.Templ)
	}
	if state.Temph != 35 {
		t.Errorf("Expected Temph 35, got %d", state.Temph)
	}
	if state.Tof != 500 {
		t.Errorf("Expected Tof 500, got %d", state.Tof)
	}
	if state.H != 150 {
		t.Errorf("Expected H 150, got %d", state.H)
	}
	if state.Bat != 90 {
		t.Errorf("Expected Bat 90, got %d", state.Bat)
	}
	if state.Baro != 1013.25 {
		t.Errorf("Expected Baro 1013.25, got %f", state.Baro)
	}
	if state.Time != 300 {
		t.Errorf("Expected Time 300, got %d", state.Time)
	}
	if state.Agx != 0.1 {
		t.Errorf("Expected Agx 0.1, got %f", state.Agx)
	}
	if state.Agy != -0.2 {
		t.Errorf("Expected Agy -0.2, got %f", state.Agy)
	}
	if state.Agz != 9.8 {
		t.Errorf("Expected Agz 9.8, got %f", state.Agz)
	}
}

func TestTelloStateEdgeValues(t *testing.T) {
	// Test edge values that might be received from drone
	state := &TelloState{
		Pitch: -180, Roll: 180, Yaw: 359, // Attitude limits
		Vgx: -100, Vgy: 100, Vgz: 100, // Velocity limits
		Templ: -40, Temph: 85, // Temperature limits
		Tof: 0, H: 3000, Bat: 0, // Sensor limits
		Baro: -500.0, Time: 9999, // Extended limits
		Agx: -10.0, Agy: 10.0, Agz: 20.0, // Acceleration limits
	}
	
	// Verify all edge values are preserved
	if state.Pitch != -180 {
		t.Errorf("Expected Pitch -180, got %d", state.Pitch)
	}
	if state.Roll != 180 {
		t.Errorf("Expected Roll 180, got %d", state.Roll)
	}
	if state.Yaw != 359 {
		t.Errorf("Expected Yaw 359, got %d", state.Yaw)
	}
	if state.Vgx != -100 {
		t.Errorf("Expected Vgx -100, got %d", state.Vgx)
	}
	if state.Vgy != 100 {
		t.Errorf("Expected Vgy 100, got %d", state.Vgy)
	}
	if state.Vgz != 100 {
		t.Errorf("Expected Vgz 100, got %d", state.Vgz)
	}
	if state.Templ != -40 {
		t.Errorf("Expected Templ -40, got %d", state.Templ)
	}
	if state.Temph != 85 {
		t.Errorf("Expected Temph 85, got %d", state.Temph)
	}
	if state.Tof != 0 {
		t.Errorf("Expected Tof 0, got %d", state.Tof)
	}
	if state.H != 3000 {
		t.Errorf("Expected H 3000, got %d", state.H)
	}
	if state.Bat != 0 {
		t.Errorf("Expected Bat 0, got %d", state.Bat)
	}
	if state.Baro != -500.0 {
		t.Errorf("Expected Baro -500.0, got %f", state.Baro)
	}
	if state.Time != 9999 {
		t.Errorf("Expected Time 9999, got %d", state.Time)
	}
	if state.Agx != -10.0 {
		t.Errorf("Expected Agx -10.0, got %f", state.Agx)
	}
	if state.Agy != 10.0 {
		t.Errorf("Expected Agy 10.0, got %f", state.Agy)
	}
	if state.Agz != 20.0 {
		t.Errorf("Expected Agz 20.0, got %f", state.Agz)
	}
}