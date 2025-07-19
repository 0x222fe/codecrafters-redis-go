package types

type OrderedMap[K comparable, V any] struct {
	head, tail *node[K, V]
	nodes      map[K]*node[K, V]
	length     int
}

type node[K comparable, V any] struct {
	key   K
	value V
	prev  *node[K, V]
	next  *node[K, V]
}

func NewMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		head:   nil,
		tail:   nil,
		nodes:  make(map[K]*node[K, V]),
		length: 0,
	}
}

func newNode[K comparable, V any](key K, val V) *node[K, V] {
	return &node[K, V]{
		key:   key,
		value: val,
		prev:  nil,
		next:  nil,
	}
}

func (n *node[K, V]) detach() {
	if n.prev != nil {
		n.prev.next = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	}
	n.prev = nil
	n.next = nil
}

func (m *OrderedMap[K, V]) Set(key K, value V) {
	n, ok := m.nodes[key]
	if ok {
		n.value = value
		return
	}

	n = newNode(key, value)
	m.length++
	m.nodes[key] = n

	if m.head == nil {
		m.head, m.tail = n, n
		return
	}

	n.prev = m.tail
	m.tail.next = n
	m.tail = n
}

func (m *OrderedMap[K, V]) Delete(key K) {
	n, ok := m.nodes[key]

	if !ok {
		return
	}

	m.length--
	delete(m.nodes, key)

	if m.head == n {
		m.head = n.next
	}

	if m.tail == n {
		m.tail = n.prev
	}
	n.detach()
}

func (m *OrderedMap[K, V]) Get(key K) (V, bool) {
	n, ok := m.nodes[key]

	if !ok {
		var v V
		return v, false
	}

	return n.value, true
}

func (m *OrderedMap[K, V]) Peek() (V, bool) {
	if m.head == nil {
		var zero V
		return zero, false
	}
	return m.head.value, true
}

func (m *OrderedMap[K, V]) Len() int {
	return m.length
}

func (m *OrderedMap[K, V]) ForEach(handler func(value V) (stop bool)) {
	for n := m.head; n != nil; n = n.next {
		node := n
		if handler(node.value) {
			break
		}
	}
}
