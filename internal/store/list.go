package store

import "sync"

type RedisList struct {
	mu   sync.RWMutex
	list []string
}

func NewList() *RedisList {
	return &RedisList{
		list: []string{},
	}
}

func (l *RedisList) Push(items ...string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.list = append(l.list, items...)

	return len(l.list)
}

func (l *RedisList) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.list)
}
