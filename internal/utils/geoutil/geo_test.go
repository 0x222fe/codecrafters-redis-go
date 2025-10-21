package geoutil

import (
	"testing"
)

var (
	tests = []struct {
		name   string
		lo, la float64
		score  float64
	}{
		{"Bangkok", 100.5252, 13.7220, 3962257306574459.0},
		{"Beijing", 116.3972, 39.9075, 4069885364908765.0},
		{"Berlin", 13.4105, 52.5244, 3673983964876493.0},
		{"Copenhagen", 12.5655, 55.6759, 3685973395504349.0},
		{"New Delhi", 77.2167, 28.6667, 3631527070936756.0},
		{"Kathmandu", 85.3206, 27.7017, 3639507404773204.0},
		{"London", -0.1278, 51.5074, 2163557714755072.0},
		{"New York", -74.0060, 40.7128, 1791873974549446.0},
		{"Paris", 2.3488, 48.8534, 3663832752681684.0},
		{"Sydney", 151.2093, -33.8688, 3252046221964352.0},
		{"Tokyo", 139.6917, 35.6895, 4171231230197045.0},
		{"Vienna", 16.3707, 48.2064, 3673109836391743.0},
	}
)

func TestGenerateScoreMatchesRedis(t *testing.T) {
	const epsilon = 1e-6

	for _, tc := range tests {
		score := GenerateScore(tc.lo, tc.la)
		if diff := score - tc.score; diff < -epsilon || diff > epsilon {
			t.Errorf("%s: Score mismatch: got %.1f, want %.1f", tc.name, score, tc.score)
		}
	}
}

func TestDecodeScoreMatchesInput(t *testing.T) {
	const epsilon = 1e-5

	for _, tc := range tests {
		lo, la := DecodeScore(tc.score)
		if diff := lo - tc.lo; diff < -epsilon || diff > epsilon {
			t.Errorf("%s: Longitude mismatch: got %.6f, want %.6f", tc.name, lo, tc.lo)
		}
		if diff := la - tc.la; diff < -epsilon || diff > epsilon {
			t.Errorf("%s: Latitude mismatch: got %.6f, want %.6f", tc.name, la, tc.la)
		}
	}
}