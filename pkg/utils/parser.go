package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/conceptcodes/dji-tello-sdk-go/shared"
)

func ParseInt(value string) (int, error) {
	return strconv.Atoi(value)
}

func ParseFloat(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

func ParseState(data string) (*shared.TelloState, error) {
	state := &shared.TelloState{}

	parts := strings.Split(data, ";")
	for _, part := range parts {
		telemetryData := strings.Split(part, ":")
		if len(telemetryData) != 2 {
			return nil, fmt.Errorf("invalid telemetry data: %s", part)
		}
		if len(telemetryData[0]) == 0 || len(telemetryData[1]) == 0 {
			return nil, fmt.Errorf("invalid telemetry data: %s", part)
		}

		key := strings.TrimSpace(telemetryData[0])
		value := strings.TrimSpace(telemetryData[1])

		switch key {
		case "pitch":
			val, err := ParseInt(value)
			if err != nil {
				Logger.Debugf("error parsing pitch value: %v", err)
				continue
			}
			state.Pitch = val
		case "roll":
			val, err := ParseInt(value)
			if err != nil {
				Logger.Debugf("error parsing roll value: %v", err)
				continue
			}
			state.Roll = val
		case "yaw":
			val, err := ParseInt(value)
			if err != nil {
				Logger.Debugf("error parsing yaw value: %v", err)
				continue
			}
			state.Yaw = val
		case "vgx":
			val, err := ParseInt(value)
			if err != nil {
				Logger.Debugf("error parsing vgx value: %v", err)
				continue
			}
			state.Vgx = val
		case "vgy":
			val, err := ParseInt(value)
			if err != nil {
				Logger.Debugf("error parsing vgy value: %v", err)
				continue
			}
			state.Vgy = val
		case "vgz":
			val, err := ParseInt(value)
			if err != nil {
				Logger.Debugf("error parsing vgz value: %v", err)
				continue
			}
			state.Vgz = val
		case "templ":
			val, err := ParseInt(value)
			if err != nil {
				Logger.Debugf("error parsing templ value: %v", err)
				continue
			}
			state.Templ = val
		case "temph":
			val, err := ParseInt(value)
			if err != nil {
				Logger.Debugf("error parsing temph value: %v", err)
				continue
			}
			state.Temph = val
		case "tof":
			val, err := ParseInt(value)
			if err != nil {
				Logger.Debugf("error parsing tof value: %v", err)
				continue
			}
			state.Tof = val
		case "h":
			val, err := ParseInt(value)
			if err != nil {
				Logger.Debugf("error parsing h value: %v", err)
				continue
			}
			state.H = val
		case "bat":
			val, err := ParseInt(value)
			if err != nil {
				Logger.Debugf("error parsing bat value: %v", err)
				continue
			}
			state.Bat = val
		case "baro":
			val, err := ParseFloat(value)
			if err != nil {
				Logger.Debugf("error parsing baro value: %v", err)
				continue
			}
			state.Baro = val
		case "time":
			val, err := ParseInt(value)
			if err != nil {
				Logger.Debugf("error parsing time value: %v", err)
				continue
			}
			state.Time = val
		case "agx":
			val, err := ParseFloat(value)
			if err != nil {
				Logger.Debugf("error parsing agx value: %v", err)
				continue
			}
			state.Agx = val
		case "agy":
			val, err := ParseFloat(value)
			if err != nil {
				Logger.Debugf("error parsing agy value: %v", err)
				continue
			}
			state.Agy = val
		case "agz":
			val, err := ParseFloat(value)
			if err != nil {
				Logger.Debugf("error parsing agz value: %v", err)
				continue
			}
			state.Agz = val
		default:
			continue
		}
	}

	return state, nil
}
