package store

import (
	"sync"
	"time"
)

type ValueType string

const (
	String ValueType = "string"
	List   ValueType = "list"
	Set    ValueType = "set"
	Hash   ValueType = "hash"
	ZSet   ValueType = "zset"
	Stream ValueType = "stream"
	None   ValueType = "none"
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
	valType  ValueType
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

func (store *Store) Set(key string, val string, valType ValueType, expireAt *int64) {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.data[key] = storeItem{
		val:      val,
		valType:  valType,
		expireAt: expireAt,
	}
}

func (store *Store) Type(key string) string {
	store.mu.RLock()
	defer store.mu.RUnlock()

	item, ok := store.data[key]
	if !ok {
		return string(None)
	}

	if item.expireAt != nil && *item.expireAt < time.Now().UnixMilli() {
		store.mu.Lock()
		delete(store.data, key)
		store.mu.Unlock()
		return string(None)
	}

	return string(item.valType)
}

func (store *Store) Keys() []string {
	store.mu.RLock()
	defer store.mu.RUnlock()

	keys := make([]string, 0, len(store.data))
	for key := range store.data {
		keys = append(keys, key)
	}
	return keys
}
