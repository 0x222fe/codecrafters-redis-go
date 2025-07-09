package store

import (
	"sync"
	"time"
)

type Store struct {
	mu   sync.RWMutex
	data map[string]storeItem
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]storeItem),
	}
}

type storeItem struct {
	val      string
	expireAt *int64
}

func (store *Store) Get(key string) (string, bool) {
	store.mu.RLock()
	item, ok := store.data[key]
	store.mu.RUnlock()

	if !ok {
		return "", false
	}

	if item.expireAt != nil && *item.expireAt < time.Now().UnixMilli() {
		store.mu.Lock()
		delete(store.data, key)
		store.mu.Unlock()
		return "", false
	}

	return item.val, true
}

func (store *Store) Set(key string, val string, expireAt *int64) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.data[key] = storeItem{
		val:      val,
		expireAt: expireAt,
	}
}
