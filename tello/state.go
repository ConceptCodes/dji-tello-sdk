package tello

type TelloState struct {
	Pitch int     `json:"pitch"` // Pitch angle of the drone in degrees
	Roll  int     `json:"roll"`  // Roll angle of the drone in degrees
	Yaw   int     `json:"yaw"`   // Yaw angle of the drone in degrees
	Vgx   int     `json:"vgx"`   // Velocity in the X direction (cm/s)
	Vgy   int     `json:"vgy"`   // Velocity in the Y direction (cm/s)
	Vgz   int     `json:"vgz"`   // Velocity in the Z direction (cm/s)
	Templ int     `json:"templ"` // Lowest temperature of the drone (°C)
	Temph int     `json:"temph"` // Highest temperature of the drone (°C)
	Tof   int     `json:"tof"`   // Time-of-flight distance (cm)
	H     int     `json:"h"`     // Height of the drone (cm)
	Bat   int     `json:"bat"`   // Battery percentage
	Baro  float64 `json:"baro"`  // Barometer reading (m)
	Time  int     `json:"time"`  // Flight time (s)
	Agx   float64 `json:"agx"`   // Acceleration in the X direction (g)
	Agy   float64 `json:"agy"`   // Acceleration in the Y direction (g)
	Agz   float64 `json:"agz"`   // Acceleration in the Z direction (g)
}
