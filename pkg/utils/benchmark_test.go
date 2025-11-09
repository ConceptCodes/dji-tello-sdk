package utils

import (
	"testing"
)

func BenchmarkParseInt(b *testing.B) {
	testCases := []string{"123", "456", "789", "0", "-123"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testCase := testCases[i%len(testCases)]
		ParseInt(testCase)
	}
}

func BenchmarkParseFloat(b *testing.B) {
	testCases := []string{"123.45", "67.89", "0.0", "-123.45", "1e10"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testCase := testCases[i%len(testCases)]
		ParseFloat(testCase)
	}
}

func BenchmarkParseState(b *testing.B) {
	testData := "pitch:10;roll:-5;yaw:180;vgx:20;vgy:30;vgz:40;templ:20;temph:30;tof:300;h:100;bat:85;baro:1013.25;time:120;agx:0.1;agy:0.2;agz:0.3;"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseState(testData)
	}
}

func BenchmarkParseStatePartial(b *testing.B) {
	testData := "pitch:15;bat:90;time:200;"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseState(testData)
	}
}

func BenchmarkValidateNumberInRange(b *testing.B) {
	testCases := []struct {
		number, min, max int
	}{
		{50, 0, 100},
		{25, 10, 50},
		{75, 0, 1000},
		{5, 1, 10},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testCase := testCases[i%len(testCases)]
		ValidateNumberInRange(testCase.number, testCase.min, testCase.max)
	}
}

func BenchmarkValidateArcRadius(b *testing.B) {
	testCases := []struct {
		x1, x2, y1, y2, z1, z2 int
		min, max               float64
	}{
		{0, 100, 0, 100, 0, 100, 0.1, 1000.0},
		{-50, 50, -50, 50, -50, 50, 0.1, 1000.0},
		{0, 10, 0, 10, 0, 10, 0.1, 100.0},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testCase := testCases[i%len(testCases)]
		ValidateArcRadius(testCase.x1, testCase.x2, testCase.y1, testCase.y2, testCase.z1, testCase.z2, testCase.min, testCase.max)
	}
}
