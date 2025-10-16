package sortedset

import (
	"github.com/0x222fe/codecrafters-redis-go/internal/types/stack"
)

type SortedSet struct {
	top, bottom *layer
	set         map[string]*node
}

func New() *SortedSet {
	return &SortedSet{
		top:    nil,
		bottom: nil,
		set:    make(map[string]*node),
	}
}

func (s *SortedSet) Set(key string, score float64) {
	currScore, ok := s.Get(key)
	if ok {
		if currScore == score {
			return
		}
		s.Remove(key)
	}
	s.add(key, score)
}

func (s *SortedSet) Remove(key string) {
	if s == nil || s.top == nil {
		return
	}

	n, ok := s.set[key]
	if !ok {
		return
	}

	delete(s.set, key)

	for n != nil {
		up := n.up
		n.detach()
		n = up
	}

	s.cleanLayers()
}

func (s *SortedSet) Get(key string) (float64, bool) {
	if s == nil || s.top == nil {
		return 0, false
	}

	n, ok := s.set[key]
	if !ok {
		return 0, false
	}

	return n.score, true
}

func (s *SortedSet) RangeByScore(min, max float64) []string {
	result := make([]string, 0)
	if s == nil || s.top == nil {
		return result
	}

	stack := s.search(min)
	n, _ := stack.Peek()
	if n == nil {
		n = s.bottom.head
	}

	if n.score < min {
		return result
	}

	// INFO: n is the rightmost node in the bottom layer with score >= min,
	// so we need to move left to find the leftmost node with the same score
	// to ensure all nodes with that score are included.
	for n.prev != nil && n.prev.score == n.score {
		n = n.prev
	}

	for n != nil && n.score <= max {
		result = append(result, n.val)
		n = n.next
	}

	return result
}

func (s *SortedSet) RangeByRank(start, stop int) []string {
	result := make([]string, 0)
	if start < 1 {
		return result
	}

	if s == nil || s.bottom == nil {
		return result
	}

	l := s.bottom
	if start > l.size {
		return result
	}

	n := l.head

	for i := 1; i < start; i++ {
		n = n.next
	}

	for range stop - start + 1 {
		result = append(result, n.val)
		if n.next == nil {
			break
		}
		n = n.next
	}

	return result
}

func (s *SortedSet) Rank(key string) (int, bool) {
	if s == nil || s.bottom == nil {
		return 0, false
	}

	n, ok := s.set[key]
	if !ok {
		return 0, false
	}

	curr := s.bottom.head

	for i := 1; curr != nil; i++ {
		if curr == n {
			return i, true
		}
		curr = curr.next
	}
	return 0, false
}

func (s *SortedSet) Len() int {
	if s == nil {
		return 0
	}
	return len(s.set)
}

func (s *SortedSet) add(key string, score float64) {
	if s == nil {
		return
	}

	if s.top == nil {
		s.top = newLayer()
		s.bottom = s.top
	}

	stack := s.search(score)

	if stack.IsEmpty() {
		l := s.bottom
		n := l.prepend(key, score)
		s.set[key] = n

		for shouldLift() {
			if l.up == nil {
				l = s.addLayer()
			} else {
				l = l.up
			}
			newN := l.prepend(key, score)
			newN.setDownNode(n)
			n = newN
		}
		return
	}

	prevNode, _ := stack.Pop()

	n := prevNode.append(key, score)
	s.set[key] = n

	for shouldLift() {
		prevNode, _ = stack.Pop()
		var newN *node

		if prevNode == nil {
			l := s.addLayer()
			newN = l.prepend(key, score)
		} else {
			newN = prevNode.append(key, score)
		}
		newN.setDownNode(n)
		n = newN
	}
}

func (s *SortedSet) addLayer() *layer {
	l := newLayer()
	l.down = s.top

	if s.top != nil {
		s.top.up = l
	}

	s.top = l

	return l
}

func (s *SortedSet) cleanLayers() {
	l := s.top
	if l == nil {
		return
	}

	for l != nil {
		if l.size > 0 {
			break
		}

		down := l.down
		l.down = nil

		if down != nil {
			down.up = nil
		}

		l = down
	}

	if l == nil {
		s.bottom = nil
	}
	s.top = l
}

// search traverses the skip list from the top layer down, searching for the *RIGHTMOST* node
// in each layer whose score is less than or equal to the given score.
// It returns a stack containing the found nodes from each layer, starting from the top.
// If no such node exists, the returned stack will be empty.
func (s *SortedSet) search(score float64) *stack.Stack[*node] {
	stack := stack.New[*node]()

	if s == nil || s.top == nil {
		return stack
	}

	var n *node
	l := s.top

	for n == nil && l != nil {
		n = l.search(score)
		l = l.down
	}

	if n == nil {
		return stack
	}

	stack.Push(n)

	for n.down != nil {
		n = n.down
		n = n.search(score)
		stack.Push(n)
	}
	return stack

}
