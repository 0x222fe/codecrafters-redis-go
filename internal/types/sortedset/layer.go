package sortedset

type layer struct {
	head     *node
	up, down *layer
	size     int
}

func newLayer() *layer {
	return &layer{}
}

// search finds the *RIGHTMOST* node in the layer whose score is less than or equal to the given score.
// Returns nil if the layer is empty or no such node exists.
func (l *layer) search(score float64) *node {
	if l == nil || l.size == 0 {
		return nil
	}

	if l.head.score > score {
		return nil
	}

	n := l.head
	for n.next != nil && n.next.score <= score {
		n = n.next
	}
	return n
}

func (l *layer) prepend(val string, score float64) *node {
	if l == nil {
		return nil
	}

	node := newNode(l, val, score)

	l.size++
	node.next = l.head
	if l.head != nil {
		l.head.prev = node
	}

	l.head = node
	return node
}
