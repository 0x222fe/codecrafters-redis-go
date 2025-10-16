package stack

import (
	"testing"
)

func TestStackIntOperations(t *testing.T) {
	type op struct {
		name   string
		value  int
		expect int
		ok     bool
	}

	tests := []struct {
		name string
		ops  []op
	}{
		{
			name: "Push and Pop",
			ops: []op{
				{"push", 1, 0, false},
				{"push", 2, 0, false},
				{"pop", 0, 2, true},
				{"pop", 0, 1, true},
				{"pop", 0, 0, false},
			},
		},
		{
			name: "Peek on empty and non-empty",
			ops: []op{
				{"peek", 0, 0, false},
				{"push", 5, 0, false},
				{"peek", 0, 5, true},
				{"pop", 0, 5, true},
				{"peek", 0, 0, false},
			},
		},
		{
			name: "Clear and Size",
			ops: []op{
				{"push", 7, 0, false},
				{"push", 8, 0, false},
				{"size", 0, 2, false},
				{"clear", 0, 0, false},
				{"size", 0, 0, false},
				{"pop", 0, 0, false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New[int]()
			for _, op := range tt.ops {
				switch op.name {
				case "push":
					s.Push(op.value)
				case "pop":
					val, ok := s.Pop()
					if val != op.expect || ok != op.ok {
						t.Errorf("Pop: got (%v, %v), want (%v, %v)", val, ok, op.expect, op.ok)
					}
				case "peek":
					val, ok := s.Peek()
					if val != op.expect || ok != op.ok {
						t.Errorf("Peek: got (%v, %v), want (%v, %v)", val, ok, op.expect, op.ok)
					}
				case "size":
					size := s.Size()
					if size != op.expect {
						t.Errorf("Size: got %v, want %v", size, op.expect)
					}
				case "clear":
					s.Clear()
				}
			}
		})
	}
}

func TestStackStringOperations(t *testing.T) {
	type op struct {
		name   string
		value  string
		expect string
		ok     bool
	}

	tests := []struct {
		name string
		ops  []op
	}{
		{
			name: "Push, Peek, Pop",
			ops: []op{
				{"push", "a", "", false},
				{"push", "b", "", false},
				{"peek", "", "b", true},
				{"pop", "", "b", true},
				{"peek", "", "a", true},
				{"pop", "", "a", true},
				{"pop", "", "", false},
			},
		},
		{
			name: "IsEmpty",
			ops: []op{
				{"push", "x", "", false},
				{"pop", "", "x", true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New[string]()
			for _, op := range tt.ops {
				switch op.name {
				case "push":
					s.Push(op.value)
					if s.IsEmpty() {
						t.Error("Stack should not be empty after push")
					}
				case "pop":
					val, ok := s.Pop()
					if val != op.expect || ok != op.ok {
						t.Errorf("Pop: got (%v, %v), want (%v, %v)", val, ok, op.expect, op.ok)
					}
				case "peek":
					val, ok := s.Peek()
					if val != op.expect || ok != op.ok {
						t.Errorf("Peek: got (%v, %v), want (%v, %v)", val, ok, op.expect, op.ok)
					}
				}
			}
			if s.IsEmpty() != true {
				t.Error("Stack should be empty at end of test")
			}
		})
	}
}

func TestStackIsEmptyAndClear(t *testing.T) {
	s := New[float64]()
	if !s.IsEmpty() {
		t.Error("Stack should be empty initially")
	}
	s.Push(3.14)
	if s.IsEmpty() {
		t.Error("Stack should not be empty after push")
	}
	s.Clear()
	if !s.IsEmpty() {
		t.Error("Stack should be empty after clear")
	}
}

func TestStackPopPeekEmpty(t *testing.T) {
	s := New[bool]()
	val, ok := s.Pop()
	if ok {
		t.Errorf("Pop on empty stack: got ok=true, want ok=false")
	}
	if val != false {
		t.Errorf("Pop on empty stack: got %v, want false (zero value)", val)
	}
	val2, ok2 := s.Peek()
	if ok2 {
		t.Errorf("Peek on empty stack: got ok=true, want ok=false")
	}
	if val2 != false {
		t.Errorf("Peek on empty stack: got %v, want false (zero value)", val2)
	}
}