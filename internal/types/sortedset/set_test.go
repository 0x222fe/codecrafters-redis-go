package sortedset

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("New should return a non-nil SortedSet")
	}
	if s.Len() != 0 {
		t.Errorf("New SortedSet should have length 0, got %d", s.Len())
	}
	if s.top != nil || s.bottom != nil {
		t.Error("New SortedSet should have nil top and bottom layers")
	}
	if len(s.set) != 0 {
		t.Error("New SortedSet should have empty map")
	}
}

func TestSet(t *testing.T) {
	type step struct {
		key   string
		score float64
	}
	type want struct {
		score float64
		ok    bool
	}
	tests := []struct {
		name  string
		steps []step
		check map[string]want
	}{
		{
			name: "insert new key",
			steps: []step{
				{"a", 1.0},
			},
			check: map[string]want{
				"a": {1.0, true},
			},
		},
		{
			name: "update existing key with same score",
			steps: []step{
				{"a", 1.0},
				{"a", 1.0},
			},
			check: map[string]want{
				"a": {1.0, true},
			},
		},
		{
			name: "update existing key with different score",
			steps: []step{
				{"a", 1.0},
				{"a", 2.0},
			},
			check: map[string]want{
				"a": {2.0, true},
			},
		},
		{
			name: "multiple keys",
			steps: []step{
				{"a", 1.0},
				{"b", 2.0},
				{"c", 3.0},
			},
			check: map[string]want{
				"a": {1.0, true},
				"b": {2.0, true},
				"c": {3.0, true},
			},
		},
		{
			name: "remove and re-add key",
			steps: []step{
				{"a", 1.0},
				{"b", 2.0},
				{"a", 3.0},
			},
			check: map[string]want{
				"a": {3.0, true},
				"b": {2.0, true},
			},
		},
		{
			name: "nil receiver",
			steps: []step{
				{"a", 1.0},
			},
			check: map[string]want{
				"a": {0.0, false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s *SortedSet
			if tt.name != "nil receiver" {
				s = New()
			}
			for _, st := range tt.steps {
				if s != nil {
					s.Set(st.key, st.score)
				}
			}
			for k, want := range tt.check {
				var gotScore float64
				var gotOk bool
				if s != nil {
					gotScore, gotOk = s.Get(k)
				}
				if gotScore != want.score || gotOk != want.ok {
					t.Errorf("Get(%q) = (%v, %v), want (%v, %v)", k, gotScore, gotOk, want.score, want.ok)
				}
			}
		})
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name    string
		keys    []string
		scores  []float64
		wantLen int
	}{
		{
			name:    "add one element",
			keys:    []string{"a"},
			scores:  []float64{1.0},
			wantLen: 1,
		},
		{
			name:    "add multiple elements",
			keys:    []string{"a", "b", "c"},
			scores:  []float64{1.0, 2.0, 3.0},
			wantLen: 3,
		},
		{
			name:    "add duplicate keys",
			keys:    []string{"a", "a", "b"},
			scores:  []float64{1.0, 2.0, 3.0},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New()
			for i, key := range tt.keys {
				s.add(key, tt.scores[i])
			}
			if s.Len() != tt.wantLen {
				t.Errorf("After adding elements, Len() = %d, want %d", s.Len(), tt.wantLen)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *SortedSet
		queryKey  string
		wantScore float64
		wantFound bool
	}{
		{
			name: "get existing element",
			setup: func() *SortedSet {
				s := New()
				s.add("a", 1.0)
				s.add("b", 2.0)
				return s
			},
			queryKey:  "b",
			wantScore: 2.0,
			wantFound: true,
		},
		{
			name: "get non-existent element",
			setup: func() *SortedSet {
				s := New()
				s.add("a", 1.0)
				return s
			},
			queryKey:  "b",
			wantScore: 0.0,
			wantFound: false,
		},
		{
			name: "get from empty set",
			setup: func() *SortedSet {
				return New()
			},
			queryKey:  "a",
			wantScore: 0.0,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			score, found := s.Get(tt.queryKey)
			if found != tt.wantFound {
				t.Errorf("Get(%q) found = %v, want %v", tt.queryKey, found, tt.wantFound)
			}
			if score != tt.wantScore {
				t.Errorf("Get(%q) score = %v, want %v", tt.queryKey, score, tt.wantScore)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *SortedSet
		removeKey string
		wantLen   int
	}{
		{
			name: "remove existing element",
			setup: func() *SortedSet {
				s := New()
				s.add("a", 1.0)
				s.add("b", 2.0)
				s.add("c", 3.0)
				return s
			},
			removeKey: "b",
			wantLen:   2,
		},
		{
			name: "remove non-existent element",
			setup: func() *SortedSet {
				s := New()
				s.add("a", 1.0)
				s.add("b", 2.0)
				return s
			},
			removeKey: "c",
			wantLen:   2,
		},
		{
			name: "remove from empty set",
			setup: func() *SortedSet {
				return New()
			},
			removeKey: "a",
			wantLen:   0,
		},
		{
			name: "remove last element",
			setup: func() *SortedSet {
				s := New()
				s.add("a", 1.0)
				return s
			},
			removeKey: "a",
			wantLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			s.Remove(tt.removeKey)

			if s.Len() != tt.wantLen {
				t.Errorf("After Remove(%q), Len() = %d, want %d", tt.removeKey, s.Len(), tt.wantLen)
			}

			_, found := s.Get(tt.removeKey)
			if found {
				t.Errorf("After Remove(%q), key should not be found", tt.removeKey)
			}
		})
	}
}

func TestRangeByScore(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *SortedSet
		min      float64
		max      float64
		wantKeys []string
	}{
		{
			name: "full range",
			setup: func() *SortedSet {
				s := New()
				s.add("a", 1.0)
				s.add("b", 2.0)
				s.add("c", 3.0)
				s.add("d", 4.0)
				s.add("e", 5.0)
				return s
			},
			min:      0.0,
			max:      6.0,
			wantKeys: []string{"a", "b", "c", "d", "e"},
		},
		{
			name: "partial range",
			setup: func() *SortedSet {
				s := New()
				s.add("a", 1.0)
				s.add("b", 2.0)
				s.add("c", 3.0)
				s.add("d", 4.0)
				s.add("e", 5.0)
				return s
			},
			min:      2.0,
			max:      4.0,
			wantKeys: []string{"b", "c", "d"},
		},
		{
			name: "same score diff key",
			setup: func() *SortedSet {
				s := New()
				s.add("a", 1.0)
				s.add("b", 1.0)
				s.add("c", 1.0)
				return s
			},
			min:      1.0,
			max:      1.0,
			wantKeys: []string{"a", "b", "c"},
		},
		{
			name: "empty range",
			setup: func() *SortedSet {
				s := New()
				s.add("a", 1.0)
				s.add("b", 2.0)
				s.add("c", 3.0)
				return s
			},
			min:      5.0,
			max:      6.0,
			wantKeys: []string{},
		},
		{
			name: "empty set",
			setup: func() *SortedSet {
				return New()
			},
			min:      1.0,
			max:      3.0,
			wantKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			result := s.RangeByScore(tt.min, tt.max)
			if !reflect.DeepEqual(result, tt.wantKeys) {
				t.Errorf("RangeByScore(%v, %v) = %v, want %v", tt.min, tt.max, result, tt.wantKeys)
			}
		})
	}
}

func TestRangeByRank(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *SortedSet
		start    int
		stop     int
		wantKeys []string
	}{
		{
			name: "full range",
			setup: func() *SortedSet {
				s := New()
				s.add("a", 1.0)
				s.add("b", 2.0)
				s.add("c", 3.0)
				s.add("d", 4.0)
				s.add("e", 5.0)
				return s
			},
			start:    1,
			stop:     5,
			wantKeys: []string{"a", "b", "c", "d", "e"},
		},
		{
			name: "partial range",
			setup: func() *SortedSet {
				s := New()
				s.add("a", 1.0)
				s.add("b", 2.0)
				s.add("c", 3.0)
				s.add("d", 4.0)
				s.add("e", 5.0)
				return s
			},
			start:    2,
			stop:     4,
			wantKeys: []string{"b", "c", "d"},
		},
		{
			name: "invalid start",
			setup: func() *SortedSet {
				s := New()
				s.add("a", 1.0)
				s.add("b", 2.0)
				s.add("c", 3.0)
				return s
			},
			start:    0,
			stop:     3,
			wantKeys: []string{},
		},
		{
			name: "start beyond size",
			setup: func() *SortedSet {
				s := New()
				s.add("a", 1.0)
				s.add("b", 2.0)
				return s
			},
			start:    3,
			stop:     4,
			wantKeys: []string{},
		},
		{
			name: "empty set",
			setup: func() *SortedSet {
				return New()
			},
			start:    1,
			stop:     3,
			wantKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			result := s.RangeByRank(tt.start, tt.stop)
			if !reflect.DeepEqual(result, tt.wantKeys) {
				t.Errorf("RangeByRank(%v, %v) = %v, want %v", tt.start, tt.stop, result, tt.wantKeys)
			}
		})
	}
}

func TestRank(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *SortedSet
		queryKey  string
		wantRank  int
		wantFound bool
	}{
		{
			name: "existing key",
			setup: func() *SortedSet {
				s := New()
				s.add("a", 1.0)
				s.add("b", 2.0)
				s.add("c", 3.0)
				s.add("d", 4.0)
				s.add("e", 5.0)
				return s
			},
			queryKey:  "c",
			wantRank:  3,
			wantFound: true,
		},
		{
			name: "first element",
			setup: func() *SortedSet {
				s := New()
				s.add("a", 1.0)
				s.add("b", 2.0)
				return s
			},
			queryKey:  "a",
			wantRank:  1,
			wantFound: true,
		},
		{
			name: "non-existent key",
			setup: func() *SortedSet {
				s := New()
				s.add("a", 1.0)
				s.add("b", 2.0)
				s.add("c", 3.0)
				return s
			},
			queryKey:  "d",
			wantRank:  0,
			wantFound: false,
		},
		{
			name: "empty set",
			setup: func() *SortedSet {
				return New()
			},
			queryKey:  "a",
			wantRank:  0,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			rank, found := s.Rank(tt.queryKey)
			if found != tt.wantFound {
				t.Errorf("Rank(%q) found = %v, want %v", tt.queryKey, found, tt.wantFound)
			}
			if rank != tt.wantRank {
				t.Errorf("Rank(%q) = %v, want %v", tt.queryKey, rank, tt.wantRank)
			}
		})
	}
}

func TestLen(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *SortedSet
		wantLen int
	}{
		{
			name: "empty set",
			setup: func() *SortedSet {
				return New()
			},
			wantLen: 0,
		},
		{
			name: "set with elements",
			setup: func() *SortedSet {
				s := New()
				s.add("a", 1.0)
				s.add("b", 2.0)
				s.add("c", 3.0)
				return s
			},
			wantLen: 3,
		},
		{
			name: "after removing elements",
			setup: func() *SortedSet {
				s := New()
				s.add("a", 1.0)
				s.add("b", 2.0)
				s.add("c", 3.0)
				s.Remove("b")
				return s
			},
			wantLen: 2,
		},
		{
			name: "nil set",
			setup: func() *SortedSet {
				return nil
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			if s.Len() != tt.wantLen {
				t.Errorf("Len() = %v, want %v", s.Len(), tt.wantLen)
			}
		})
	}
}
