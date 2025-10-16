package sortedset

import (
	"testing"
)

func TestNodeAppend(t *testing.T) {
	tests := []struct {
		name     string
		initVals []struct {
			val   string
			score float64
		}
		appendVal   string
		appendScore float64
		wantNextVal string
		wantSize    int
	}{
		{
			name: "append to single node",
			initVals: []struct {
				val   string
				score float64
			}{
				{"a", 1.0},
			},
			appendVal:   "b",
			appendScore: 2.0,
			wantNextVal: "b",
			wantSize:    2,
		},
		{
			name: "append to chain",
			initVals: []struct {
				val   string
				score float64
			}{
				{"a", 1.0},
				{"b", 2.0},
			},
			appendVal:   "c",
			appendScore: 3.0,
			wantNextVal: "c",
			wantSize:    3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := newLayer()
			var n *node
			for _, v := range tt.initVals {
				if n == nil {
					n = l.prepend(v.val, v.score)
				} else {
					n = n.append(v.val, v.score)
				}
			}
			newN := n.append(tt.appendVal, tt.appendScore)

			if n.next == nil || n.next.val != tt.wantNextVal {
				t.Errorf("next.val = %v, want %v", n.next.val, tt.wantNextVal)
			}
			if l.size != tt.wantSize {
				t.Errorf("layer.size = %d, want %d", l.size, tt.wantSize)
			}
			if newN.prev != n {
				t.Errorf("new node prev = %v, want %v", newN.prev, n)
			}
		})
	}
}

func TestNodeDetach(t *testing.T) {
	t.Run("detach head node", func(t *testing.T) {
		l := newLayer()
		n1 := l.prepend("a", 1.0)
		n2 := n1.append("b", 2.0)
		n1.detach()
		if l.head != n2 {
			t.Errorf("head = %v, want %v", l.head, n2)
		}
		if l.size != 1 {
			t.Errorf("size = %d, want 1", l.size)
		}
	})

	t.Run("detach middle node", func(t *testing.T) {
		l := newLayer()
		n1 := l.prepend("a", 1.0)
		n2 := n1.append("b", 2.0)
		n3 := n2.append("c", 3.0)
		n2.detach()
		if n1.next != n3 {
			t.Errorf("n1.next = %v, want %v", n1.next, n3)
		}
		if n3.prev != n1 {
			t.Errorf("n3.prev = %v, want %v", n3.prev, n1)
		}
		if l.size != 2 {
			t.Errorf("size = %d, want 2", l.size)
		}
	})

	t.Run("detach tail node", func(t *testing.T) {
		l := newLayer()
		n1 := l.prepend("a", 1.0)
		n2 := n1.append("b", 2.0)
		n2.detach()
		if n1.next != nil {
			t.Errorf("n1.next = %v, want nil", n1.next)
		}
		if l.size != 1 {
			t.Errorf("size = %d, want 1", l.size)
		}
	})

	t.Run("detach only node", func(t *testing.T) {
		l := newLayer()
		n := l.prepend("a", 1.0)
		n.detach()
		if l.head != nil {
			t.Errorf("head = %v, want nil", l.head)
		}
		if l.size != 0 {
			t.Errorf("size = %d, want 0", l.size)
		}
	})
}

func TestNodeSetDownNode(t *testing.T) {
	t.Run("set down node in different layers", func(t *testing.T) {
		l1 := newLayer()
		l2 := newLayer()
		n1 := l1.prepend("a", 1.0)
		n2 := l2.prepend("b", 2.0)
		n1.setDownNode(n2)
		if n1.down != n2 {
			t.Errorf("n1.down = %v, want %v", n2, n1.down)
		}
		if n2.up != n1 {
			t.Errorf("n2.up = %v, want %v", n1, n2.up)
		}
	})

	t.Run("set down node with nil", func(t *testing.T) {
		l := newLayer()
		n1 := l.prepend("a", 1.0)
		n1.setDownNode(nil)
		if n1.down != nil {
			t.Errorf("n1.down = %v, want nil", n1.down)
		}
	})

	t.Run("set down node in same layer", func(t *testing.T) {
		l := newLayer()
		n1 := l.prepend("a", 1.0)
		n2 := l.prepend("b", 2.0)
		n1.setDownNode(n2)
		if n1.down != nil {
			t.Errorf("n1.down = %v, want nil (same layer)", n1.down)
		}
		if n2.up != nil {
			t.Errorf("n2.up = %v, want nil (same layer)", n2.up)
		}
	})
}

func TestNodeSearch(t *testing.T) {
	n1 := &node{score: 1.0, val: "a"}
	n2 := &node{score: 2.0, val: "b"}
	n3 := &node{score: 3.0, val: "c"}
	n1.next = n2
	n2.next = n3

	tests := []struct {
		name    string
		start   *node
		score   float64
		wantVal string
		wantNil bool
	}{
		{"nil node", nil, 1.0, "", true},
		{"score less than first", n1, 0.5, "", true},
		{"score equal to first", n1, 1.0, "a", false},
		{"score between first and second", n1, 1.5, "a", false},
		{"score equal to second", n1, 2.0, "b", false},
		{"score between second and third", n1, 2.5, "b", false},
		{"score equal to third", n1, 3.0, "c", false},
		{"score greater than all", n1, 4.0, "c", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.start.search(tt.score)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
			} else {
				if got == nil || got.val != tt.wantVal {
					t.Errorf("expected val %q, got %v", tt.wantVal, got)
				}
			}
		})
	}
}

func TestNodeSearch_DuplicateScores(t *testing.T) {
	n1 := &node{score: 1.0, val: "a"}
	n2 := &node{score: 2.0, val: "b"}
	n3 := &node{score: 2.0, val: "c"}
	n4 := &node{score: 3.0, val: "d"}
	n1.next = n2
	n2.next = n3
	n3.next = n4

	tests := []struct {
		name    string
		score   float64
		wantVal string
	}{
		{"score equal to duplicate", 2.0, "c"},
		{"score between duplicates and next", 2.5, "c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := n1.search(tt.score)
			if got == nil || got.val != tt.wantVal {
				t.Errorf("expected val %q, got %v", tt.wantVal, got)
			}
		})
	}
}