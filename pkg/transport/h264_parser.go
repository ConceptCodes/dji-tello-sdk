package transport

import (
	"fmt"
)

// H.264 NAL Unit Types
const (
	NALUTypeSlice      = 1
	NALUTypeIDR        = 5
	NALUTypeSPS        = 7
	NALUTypePPS        = 8
	NALUTypeAUD        = 9
	NALUTypeSEI        = 6
)

// H264Parser provides basic H.264 stream parsing capabilities
type H264Parser struct{}

// NALUnit represents a Network Abstraction Layer Unit
type NALUnit struct {
	Type       byte
	RefIDC     byte
	Data       []byte
	StartCode  []byte
	Size       int
	IsKeyFrame bool
}

// NewH264Parser creates a new H.264 parser
func NewH264Parser() *H264Parser {
	return &H264Parser{}
}

// ParseFrame parses H.264 data and extracts NAL units
func (p *H264Parser) ParseFrame(data []byte) ([]NALUnit, error) {
	var nalUnits []NALUnit
	offset := 0

	for offset < len(data) {
		// Find start code
		startCode, startOffset, err := p.findStartCode(data[offset:])
		if err != nil {
			if offset == 0 {
				return nil, fmt.Errorf("no start code found in frame data")
			}
			// No more start codes, we're at the end
			break
		}

		// Find next start code to determine NAL unit boundaries
		nextStartOffset := -1
		if offset+startOffset+len(startCode) < len(data) {
			if _, nextOffset, err := p.findStartCode(data[offset+startOffset+len(startCode):]); err == nil {
				nextStartOffset = offset + startOffset + len(startCode) + nextOffset
			}
		}

		// Extract NAL unit data
		var nalData []byte
		if nextStartOffset != -1 {
			nalData = data[offset+startOffset+len(startCode) : nextStartOffset]
		} else {
			nalData = data[offset+startOffset+len(startCode):]
		}

		if len(nalData) == 0 {
			offset = offset + startOffset + len(startCode)
			continue
		}

		// Parse NAL header
		nalHeader := nalData[0]
		nalType := nalHeader & 0x1F
		refIDC := (nalHeader >> 5) & 0x03

		nalUnit := NALUnit{
			Type:       nalType,
			RefIDC:     refIDC,
			Data:       nalData,
			StartCode:  startCode,
			Size:       len(nalData),
			IsKeyFrame: nalType == NALUTypeIDR,
		}

		nalUnits = append(nalUnits, nalUnit)

		if nextStartOffset != -1 {
			offset = nextStartOffset
		} else {
			break
		}
	}

	return nalUnits, nil
}

// findStartCode finds H.264 start codes (0x000001 or 0x00000001)
func (p *H264Parser) findStartCode(data []byte) ([]byte, int, error) {
	if len(data) < 3 {
		return nil, 0, fmt.Errorf("data too short for start code")
	}

	// Look for 4-byte start code (0x00000001) first - more specific
	for i := 0; i <= len(data)-4; i++ {
		if data[i] == 0x00 && data[i+1] == 0x00 && data[i+2] == 0x00 && data[i+3] == 0x01 {
			return data[i : i+4], i, nil
		}
	}

	// Look for 3-byte start code (0x000001)
	for i := 0; i <= len(data)-3; i++ {
		if data[i] == 0x00 && data[i+1] == 0x00 && data[i+2] == 0x01 {
			return data[i : i+3], i, nil
		}
	}

	return nil, 0, fmt.Errorf("no start code found")
}

// GetNALUTypeName returns the human-readable name for a NAL unit type
func (p *H264Parser) GetNALUTypeName(nalType byte) string {
	switch nalType {
	case NALUTypeSlice:
		return "Slice"
	case NALUTypeIDR:
		return "IDR Frame"
	case NALUTypeSPS:
		return "SPS"
	case NALUTypePPS:
		return "PPS"
	case NALUTypeAUD:
		return "AUD"
	case NALUTypeSEI:
		return "SEI"
	default:
		return fmt.Sprintf("Unknown (%d)", nalType)
	}
}

// ExtractSPS extracts Sequence Parameter Set from NAL units
func (p *H264Parser) ExtractSPS(nalUnits []NALUnit) []byte {
	for _, nalUnit := range nalUnits {
		if nalUnit.Type == NALUTypeSPS {
			return nalUnit.Data
		}
	}
	return nil
}

// ExtractPPS extracts Picture Parameter Set from NAL units
func (p *H264Parser) ExtractPPS(nalUnits []NALUnit) []byte {
	for _, nalUnit := range nalUnits {
		if nalUnit.Type == NALUTypePPS {
			return nalUnit.Data
		}
	}
	return nil
}

// HasKeyFrame checks if the frame contains a key frame (IDR)
func (p *H264Parser) HasKeyFrame(nalUnits []NALUnit) bool {
	for _, nalUnit := range nalUnits {
		if nalUnit.IsKeyFrame {
			return true
		}
	}
	return false
}

// GetFrameInfo returns metadata about the parsed frame
func (p *H264Parser) GetFrameInfo(nalUnits []NALUnit) map[string]interface{} {
	info := make(map[string]interface{})
	
	var nalTypes []string
	var keyFrames []int
	var totalSize int
	
	for i, nalUnit := range nalUnits {
		nalTypes = append(nalTypes, p.GetNALUTypeName(nalUnit.Type))
		if nalUnit.IsKeyFrame {
			keyFrames = append(keyFrames, i)
		}
		totalSize += nalUnit.Size
	}
	
	info["nal_unit_count"] = len(nalUnits)
	info["nal_types"] = nalTypes
	info["total_size"] = totalSize
	info["has_key_frame"] = len(keyFrames) > 0
	info["key_frame_indices"] = keyFrames
	info["sps_present"] = p.ExtractSPS(nalUnits) != nil
	info["pps_present"] = p.ExtractPPS(nalUnits) != nil
	
	return info
}