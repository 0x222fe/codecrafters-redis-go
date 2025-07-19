package store

import (
	"sync"
)

type RedisList struct {
	mu   sync.RWMutex
	list []string
}

func NewList() *RedisList {
	return &RedisList{
		list: []string{},
	}
}

func (l *RedisList) LPush(items ...string) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	n := len(items)
	newList := make([]string, n+len(l.list))
	for i := range n {
		newList[i] = items[n-1-i]
	}
	copy(newList[n:], l.list)
	l.list = newList
	return len(l.list)
}

func (l *RedisList) RPush(items ...string) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.list = append(l.list, items...)
	return len(l.list)
}

func (l *RedisList) LPop(count int) ([]string, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	n := len(l.list)
	if n == 0 || count <= 0 {
		return nil, false
	}
	if count > n {
		count = n
	}

	items := l.list[:count]
	l.list = l.list[count:]
	return items, true
}

func (l *RedisList) RPop(count int) ([]string, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	n := len(l.list)
	if n == 0 || count <= 0 {
		return nil, false
	}
	if count > n {
		count = n
	}

	items := l.list[n-count:]
	l.list = l.list[:n-count]
	return items, true
}

func (l *RedisList) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.list)
}

func (l *RedisList) GetRange(start, end int) []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	n := len(l.list)
	if n == 0 {
		return []string{}
	}
	if start < 0 {
		start += n
	}
	if end < 0 {
		end += n
	}
	if start < 0 {
		start = 0
	}
	if end >= n {
		end = n - 1
	}
	if start > end || start >= n {
		return []string{}
	}
	return l.list[start : end+1]
}
