package orderedmap

import (
	"reflect"
	"testing"
)

func TestOrderedMap_BasicOps(t *testing.T) {
	type kv struct{ k, v int }
	tests := []struct {
		name   string
		ops    []kv
		getKey int
		want   int
		wantOk bool
	}{
		{
			name:   "set and get",
			ops:    []kv{{1, 10}, {2, 20}, {3, 30}},
			getKey: 2,
			want:   20,
			wantOk: true,
		},
		{
			name:   "overwrite",
			ops:    []kv{{1, 10}, {1, 99}},
			getKey: 1,
			want:   99,
			wantOk: true,
		},
		{
			name:   "missing key",
			ops:    []kv{{1, 10}},
			getKey: 2,
			want:   0,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New[int, int]()
			for _, op := range tt.ops {
				m.Set(op.k, op.v)
			}
			got, ok := m.Get(tt.getKey)
			if got != tt.want || ok != tt.wantOk {
				t.Errorf("Get(%v) = (%v, %v), want (%v, %v)", tt.getKey, got, ok, tt.want, tt.wantOk)
			}
		})
	}
}

func TestOrderedMap_Delete(t *testing.T) {
	type op struct {
		name   string
		set    []struct{ k, v int }
		delete int
		want   map[int]int
	}
	tests := []op{
		{
			name:   "delete middle key",
			set:    []struct{ k, v int }{{1, 10}, {2, 20}, {3, 30}},
			delete: 2,
			want:   map[int]int{1: 10, 3: 30},
		},
		{
			name:   "delete only key",
			set:    []struct{ k, v int }{{1, 10}},
			delete: 1,
			want:   map[int]int{},
		},
		{
			name:   "delete non-existent key",
			set:    []struct{ k, v int }{{1, 10}, {2, 20}},
			delete: 99,
			want:   map[int]int{1: 10, 2: 20},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New[int, int]()
			for _, kv := range tt.set {
				m.Set(kv.k, kv.v)
			}
			m.Delete(tt.delete)
			if m.Len() != len(tt.want) {
				t.Errorf("Len() = %d, want %d", m.Len(), len(tt.want))
			}
			for k, v := range tt.want {
				got, ok := m.Get(k)
				if !ok || got != v {
					t.Errorf("Get(%d) = (%d, %v), want (%d, true)", k, got, ok, v)
				}
			}
		})
	}
}

func TestOrderedMap_KeysValuesOrder(t *testing.T) {
	type pair struct {
		k int
		v string
	}
	tests := []struct {
		name  string
		items []pair
	}{
		{name: "multiple items", items: []pair{{3, "c"}, {1, "a"}, {2, "b"}}},
		{name: "single item", items: []pair{{1, "x"}}},
		{name: "empty", items: []pair{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New[int, string]()
			for _, item := range tt.items {
				m.Set(item.k, item.v)
			}
			for _, item := range tt.items {
				got, ok := m.Get(item.k)
				if !ok || got != item.v {
					t.Errorf("Get(%d) = (%q, %v), want (%q, true)", item.k, got, ok, item.v)
				}
			}
		})
	}
}

func TestOrderedMap_ForEach(t *testing.T) {
	type pair struct {
		k int
		v string
	}
	tests := []struct {
		name     string
		items    []pair
		wantKeys []int
		wantVals []string
	}{
		{
			name:     "multiple items",
			items:    []pair{{1, "a"}, {2, "b"}, {3, "c"}},
			wantKeys: []int{1, 2, 3},
			wantVals: []string{"a", "b", "c"},
		},
		{
			name:     "single item",
			items:    []pair{{5, "x"}},
			wantKeys: []int{5},
			wantVals: []string{"x"},
		},
		{
			name:     "empty",
			items:    []pair{},
			wantKeys: []int{},
			wantVals: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New[int, string]()
			for _, item := range tt.items {
				m.Set(item.k, item.v)
			}
			gotKeys := []int{}
			gotVals := []string{}
			m.ForEach(func(k int, v string) bool {
				gotKeys = append(gotKeys, k)
				gotVals = append(gotVals, v)
				return false
			})
			if !reflect.DeepEqual(gotKeys, tt.wantKeys) || !reflect.DeepEqual(gotVals, tt.wantVals) {
				t.Errorf("ForEach got keys %v vals %v, want keys %v vals %v", gotKeys, gotVals, tt.wantKeys, tt.wantVals)
			}
		})
	}
}