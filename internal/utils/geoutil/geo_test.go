package geoutil

import (
	"math"
	"testing"
)

func almostEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) <= epsilon
}

func TestGenerateScoreAndDecodeScore(t *testing.T) {
	tests := []struct {
		name string
		lo   float64
		la   float64
	}{
		{"Zero", 0, 0},
		{"Min values", minLon, minLat},
		{"Max values", maxLon, maxLat},
		{"Negative values", -77.0365, -12.0432},
		{"Positive values", 120.9842, 14.5995},
		{"Edge case 1", -180, 85.05112878},
		{"Edge case 2", 180, -85.05112878},
	}

	const epsilon = 1e-5

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := GenerateScore(tt.lo, tt.la)
			lo2, la2 := DecodeScore(score)
			if !almostEqual(tt.lo, lo2, epsilon) || !almostEqual(tt.la, la2, epsilon) {
				t.Errorf("Round-trip failed: got (%f, %f), want (%f, %f)", lo2, la2, tt.lo, tt.la)
			}
		})
	}
}