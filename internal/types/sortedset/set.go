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

func (s *SortedSet) Set(key string, score float64) int {
	currScore, ok := s.Get(key)
	var count int
	if ok {
		count = 0
		if currScore == score {
			return count
		}
		s.Remove(key)
	} else {
		count = 1
	}

	s.add(key, score)
	return count
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

	if start < -s.Len() {
		start = 0
	} else if start < 0 {
		start = s.Len() + start
	}

	if stop < -s.Len() {
		stop = 0
	} else if stop < 0 {
		stop = s.Len() + stop
	}

	if s == nil || s.bottom == nil {
		return result
	}

	l := s.bottom
	if start > l.size {
		return result
	}

	n := l.head

	for range start {
		if n.next == nil {
			return result
		}
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
		return -1, false
	}

	n, ok := s.set[key]
	if !ok {
		return -1, false
	}

	curr := s.bottom.head

	for i := 0; curr != nil; i++ {
		if curr == n {
			return i, true
		}
		curr = curr.next
	}
	return -1, false
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

	currNode, _ := stack.Pop()

	if currNode != nil {
		// INFO: 1.currNode is the rightmost node
		// 2.with same score, we should preserve alphabetical order
		for currNode != nil && currNode.score == score && currNode.val > key {
			currNode = currNode.prev
		}
	}

	if currNode == nil {
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

	n := currNode.append(key, score)
	s.set[key] = n

	for shouldLift() {
		currNode, _ = stack.Pop()
		var newN *node

		if currNode == nil {
			l := s.addLayer()
			newN = l.prepend(key, score)
		} else {
			newN = currNode.append(key, score)
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
