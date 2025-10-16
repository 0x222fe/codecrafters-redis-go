package sortedset

import (
	"testing"
)

func TestNewLayer(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"new layer returns non-nil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := newLayer()
			if l == nil {
				t.Errorf("newLayer() returned nil")
			}
			if l.size != 0 {
				t.Errorf("newLayer() size = %d, want 0", l.size)
			}
			if l.head != nil {
				t.Errorf("newLayer() head = %v, want nil", l.head)
			}
		})
	}
}

func TestLayerPrepend(t *testing.T) {
	tests := []struct {
		name   string
		values []struct {
			val   string
			score float64
		}
		wantHeadScore float64
		wantSize      int
	}{
		{
			name: "prepend to empty layer",
			values: []struct {
				val   string
				score float64
			}{
				{"a", 1.0},
			},
			wantHeadScore: 1.0,
			wantSize:      1,
		},
		{
			name: "prepend multiple nodes",
			values: []struct {
				val   string
				score float64
			}{
				{"a", 1.0},
				{"b", 2.0},
			},
			wantHeadScore: 2.0,
			wantSize:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := newLayer()
			for _, v := range tt.values {
				l.prepend(v.val, v.score)
			}
			if l.head == nil || l.head.score != tt.wantHeadScore {
				t.Errorf("head.score = %v, want %v", l.head.score, tt.wantHeadScore)
			}
			if l.size != tt.wantSize {
				t.Errorf("size = %d, want %d", l.size, tt.wantSize)
			}
		})
	}
}

func TestLayerSearch(t *testing.T) {
	buildLayer := func(scores []float64) *layer {
		l := newLayer()
		for i := len(scores) - 1; i >= 0; i-- {
			l.prepend("v", scores[i])
		}
		return l
	}

	two, three, five := 2.0, 3.0, 5.0

	tests := []struct {
		name      string
		scores    []float64
		search    float64
		wantScore *float64
	}{
		{
			name:      "empty layer",
			scores:    []float64{},
			search:    1.0,
			wantScore: nil,
		},
		{
			name:      "score less than head",
			scores:    []float64{2.0, 3.0},
			search:    1.0,
			wantScore: nil,
		},
		{
			name:      "score equal to head",
			scores:    []float64{2.0, 3.0},
			search:    2.0,
			wantScore: &two,
		},
		{
			name:      "score between nodes",
			scores:    []float64{2.0, 3.0, 5.0},
			search:    4.0,
			wantScore: &three,
		},
		{
			name:      "score greater than all",
			scores:    []float64{2.0, 3.0, 5.0},
			search:    6.0,
			wantScore: &five,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := buildLayer(tt.scores)
			got := l.search(tt.search)
			if tt.wantScore == nil {
				if got != nil {
					t.Errorf("search(%v) = %v, want nil", tt.search, got)
				}
			} else {
				if got == nil || got.score != *tt.wantScore {
					t.Errorf("search(%v) = %v, want score %v", tt.search, got, *tt.wantScore)
				}
			}
		})
	}
}
