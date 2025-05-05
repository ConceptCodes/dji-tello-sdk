package parser

import (
	"fmt"
	"strings"

	"github.com/conceptcodes/dji-tello-sdk-go/tello"
)

func ParseState(data string) (*tello.TelloState, error) {
	state := &tello.TelloState{}

	parts := strings.Split(data, ";")
	for _, part := range parts {
		telemetryData := strings.Split(part, ":")
		if len(telemetryData) != 2 {
			return nil, fmt.Errorf("invalid telemetry data: %s", part)
		}
		key := strings.TrimSpace(telemetryData[0])
		value := strings.TrimSpace(telemetryData[1])
		switch key {
		case "pitch":
			state.Pitch = parseInt(value)
		case "roll":
			state.Roll = parseInt(value)
		case "yaw":
			state.Yaw = parseInt(value)
		case "vgx":
			state.Vgx = parseInt(value)
		case "vgy":
			state.Vgy = parseInt(value)
		case "vgz":
			state.Vgz = parseInt(value)
		case "templ":
			state.Templ = parseInt(value)
		case "temph":
			state.Temph = parseInt(value)
		case "tof":
			state.Tof = parseInt(value)
		case "h":
			state.H = parseInt(value)
		case "bat":
			state.Bat = parseInt(value)
		case "baro":
			state.Baro = parseFloat(value)
		case "time":
			state.Time = parseInt(value)
		case "agx":
			state.Agx = parseFloat(value)
		case "agy":
			state.Agy = parseFloat(value)
		case "agz":
			state.Agz = parseFloat(value)
		default:
			// Ignore unknown keys
			continue
		}
	}

	return state, nil
}

func parseInt(value string) int {
	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return 0 
	}
	return result
}

func parseFloat(value string) float64 {
	var result float64
	_, err := fmt.Sscanf(value, "%f", &result)
	if err != nil {
		return 0.0 
	}
	return result
}
