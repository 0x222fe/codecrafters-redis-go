package sortedset

import (
	"math/rand/v2"
)

type node struct {
	prev, next, up, down *node
	score                float64
	val                  string
	layer                *layer
}

func newNode(layer *layer, val string, score float64) *node {
	return &node{
		score: score,
		val:   val,
		layer: layer,
	}
}

func shouldLift() bool {
	return rand.Int32N(2) == 0
}

func (n *node) append(val string, score float64) *node {
	if n == nil {
		return nil
	}

	newNode := newNode(n.layer, val, score)

	if n.next != nil {
		n.next.prev = newNode
	}
	n.next = newNode

	n.next = newNode
	newNode.prev = n

	n.layer.size++
	return newNode
}

func (n *node) detach() {
	if n == nil || n.layer == nil {
		return
	}

	if n.prev != nil {
		n.prev.next = n.next
	} else {
		n.layer.head = n.next
	}

	if n.next != nil {
		n.next.prev = n.prev
	}
	n.prev = nil
	n.next = nil

	if n.up != nil {
		n.up.down = nil
	}
	n.up = nil

	if n.down != nil {
		n.down.up = nil
	}
	n.down = nil

	n.layer.size--

	if n.layer.size == 0 {
		n.layer.head = nil
	}
}

// search starts from the current node (n) and search *HORIZONTALLY* and finds
// the *RIGHTMOST* node whose score is less than or equal to the given score.
// Returns nil if the starting node is nil or if the score is less than the starting node's score.
func (n *node) search(score float64) *node {
	if n == nil || score < n.score {
		return nil
	}

	for n.next != nil && n.next.score <= score {
		n = n.next
	}

	return n
}

func (n *node) setDownNode(down *node) {
	if n == nil || down == nil || down.layer == n.layer {
		return
	}
	n.down = down
	down.up = n
}
