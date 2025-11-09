package transport

import (
	"testing"
	"time"
)

func TestH264Parser_ParseFrame(t *testing.T) {
	parser := NewH264Parser()

	// Test with valid H.264 data with start codes
	testData := []byte{0x00, 0x00, 0x01, 0x67, 0x42, 0x00, 0x1E, 0x8D, 0x40, 0x50, 0x17, 0xFC, 0xB0, 0x0F, 0x08, 0x84, 0x6A}

	nalUnits, err := parser.ParseFrame(testData)
	if err != nil {
		t.Errorf("Expected no error parsing valid H.264 data, got %v", err)
	}

	if len(nalUnits) == 0 {
		t.Error("Expected at least one NAL unit, got none")
	}

	// Check first NAL unit
	firstNAL := nalUnits[0]
	if firstNAL.Type != NALUTypeSPS {
		t.Errorf("Expected SPS NAL unit type (7), got %d", firstNAL.Type)
	}

	if len(firstNAL.Data) == 0 {
		t.Error("Expected NAL unit data, got empty slice")
	}
}

func TestH264Parser_ParseFrameMultipleNALUnits(t *testing.T) {
	parser := NewH264Parser()

	// Test with multiple NAL units
	testData := []byte{
		0x00, 0x00, 0x01, 0x67, 0x42, 0x00, 0x1E, // SPS
		0x00, 0x00, 0x01, 0x68, 0xCE, 0x3C, 0x80, // PPS
		0x00, 0x00, 0x01, 0x41, 0xEA, 0x20, 0x80, // Slice
	}

	nalUnits, err := parser.ParseFrame(testData)
	if err != nil {
		t.Errorf("Expected no error parsing multiple NAL units, got %v", err)
	}

	if len(nalUnits) != 3 {
		t.Errorf("Expected 3 NAL units, got %d", len(nalUnits))
	}

	// Check NAL unit types
	expectedTypes := []byte{NALUTypeSPS, NALUTypePPS, NALUTypeSlice}
	for i, expectedType := range expectedTypes {
		if nalUnits[i].Type != expectedType {
			t.Errorf("Expected NAL unit %d to have type %d, got %d", i, expectedType, nalUnits[i].Type)
		}
	}
}

func TestH264Parser_FindStartCode(t *testing.T) {
	parser := NewH264Parser()

	// Test 3-byte start code
	data1 := []byte{0x00, 0x00, 0x01, 0x67}
	startCode, offset, err := parser.findStartCode(data1)
	if err != nil {
		t.Errorf("Expected no error finding 3-byte start code, got %v", err)
	}
	if offset != 0 {
		t.Errorf("Expected offset 0, got %d", offset)
	}
	if len(startCode) != 3 || startCode[0] != 0x00 || startCode[1] != 0x00 || startCode[2] != 0x01 {
		t.Error("Expected 3-byte start code 0x000001")
	}

	// Test 4-byte start code
	data2 := []byte{0x00, 0x00, 0x00, 0x01, 0x67}
	startCode, offset, err = parser.findStartCode(data2)
	if err != nil {
		t.Errorf("Expected no error finding 4-byte start code, got %v", err)
	}
	if offset != 0 {
		t.Errorf("Expected offset 0, got %d", offset)
	}
	if len(startCode) != 4 || startCode[0] != 0x00 || startCode[1] != 0x00 || startCode[2] != 0x00 || startCode[3] != 0x01 {
		t.Errorf("Expected 4-byte start code 0x00000001, got %v", startCode)
	}

	// Test no start code
	data3 := []byte{0x01, 0x02, 0x03}
	_, _, err = parser.findStartCode(data3)
	if err == nil {
		t.Error("Expected error finding start code in invalid data")
	}
}

func TestH264Parser_GetNALUTypeName(t *testing.T) {
	parser := NewH264Parser()

	tests := map[byte]string{
		NALUTypeSlice: "Slice",
		NALUTypeIDR:   "IDR Frame",
		NALUTypeSPS:   "SPS",
		NALUTypePPS:   "PPS",
		NALUTypeAUD:   "AUD",
		NALUTypeSEI:   "SEI",
		99:            "Unknown (99)",
	}

	for nalType, expectedName := range tests {
		name := parser.GetNALUTypeName(nalType)
		if name != expectedName {
			t.Errorf("Expected NALU type name '%s' for type %d, got '%s'", expectedName, nalType, name)
		}
	}
}

func TestH264Parser_HasKeyFrame(t *testing.T) {
	parser := NewH264Parser()

	// Test with IDR frame (key frame)
	nalUnitsWithKeyFrame := []NALUnit{
		{Type: NALUTypeSPS, IsKeyFrame: false},
		{Type: NALUTypeIDR, IsKeyFrame: true},
		{Type: NALUTypeSlice, IsKeyFrame: false},
	}

	if !parser.HasKeyFrame(nalUnitsWithKeyFrame) {
		t.Error("Expected HasKeyFrame to return true with IDR frame")
	}

	// Test without key frame
	nalUnitsWithoutKeyFrame := []NALUnit{
		{Type: NALUTypeSPS, IsKeyFrame: false},
		{Type: NALUTypeSlice, IsKeyFrame: false},
	}

	if parser.HasKeyFrame(nalUnitsWithoutKeyFrame) {
		t.Error("Expected HasKeyFrame to return false without IDR frame")
	}
}

func TestH264Parser_GetFrameInfo(t *testing.T) {
	parser := NewH264Parser()

	nalUnits := []NALUnit{
		{Type: NALUTypeSPS, Size: 10, IsKeyFrame: false, Data: []byte{0x67}},
		{Type: NALUTypePPS, Size: 5, IsKeyFrame: false, Data: []byte{0x68}},
		{Type: NALUTypeIDR, Size: 100, IsKeyFrame: true, Data: []byte{0x41}},
	}

	info := parser.GetFrameInfo(nalUnits)

	// Check basic info
	if info["nal_unit_count"] != 3 {
		t.Errorf("Expected nal_unit_count 3, got %v", info["nal_unit_count"])
	}

	if info["total_size"] != 115 {
		t.Errorf("Expected total_size 115, got %v", info["total_size"])
	}

	if info["has_key_frame"] != true {
		t.Error("Expected has_key_frame to be true")
	}

	if info["sps_present"] != true {
		t.Error("Expected sps_present to be true")
	}

	if info["pps_present"] != true {
		t.Error("Expected pps_present to be true")
	}
}

func TestVideoFrame(t *testing.T) {
	frame := VideoFrame{
		Data:       []byte{0x00, 0x01, 0x02},
		Timestamp:  time.Now(),
		Size:       3,
		SeqNum:     1,
		NALUnits:   []NALUnit{{Type: NALUTypeSPS}},
		IsKeyFrame: false,
	}

	if frame.Size != len(frame.Data) {
		t.Error("Expected frame size to match data length")
	}

	if frame.SeqNum != 1 {
		t.Error("Expected sequence number 1")
	}

	if len(frame.NALUnits) != 1 {
		t.Error("Expected 1 NAL unit")
	}

	if frame.IsKeyFrame {
		t.Error("Expected IsKeyFrame to be false")
	}
}
