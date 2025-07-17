package store

import (
	"fmt"
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
	mu      sync.RWMutex
	data    map[string]storeItem
	streams map[string]*RedisStream
}

func NewStore() *Store {
	return &Store{
		data:    make(map[string]storeItem),
		streams: make(map[string]*RedisStream),
	}
}

type storeItem struct {
	val      any
	valType  ValueType
	expireAt *int64
}

func (store *Store) GetString(key string) (string, bool) {
	store.mu.RLock()
	item, ok := store.data[key]
	store.mu.RUnlock()

	if !ok {
		return "", false
	}

	if item.valType != String {
		return "", false
	}

	if item.expireAt != nil && *item.expireAt < time.Now().UnixMilli() {
		store.mu.Lock()
		delete(store.data, key)
		store.mu.Unlock()
		return "", false
	}

	strVal, ok := item.val.(string)
	if !ok {
		return "", false
	}
	return strVal, true
}

func (store *Store) GetStream(key string) (*RedisStream, bool) {
	store.mu.RLock()
	stream, ok := store.streams[key]
	store.mu.RUnlock()

	if !ok {
		return nil, false
	}

	return stream, true
}

func (store *Store) AddStream(key string, stream *RedisStream) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	_, ok := store.streams[key]
	if ok {
		return fmt.Errorf("stream already exists: %s", key)
	}

	store.streams[key] = stream
	return nil
}

func (store *Store) Get(key string, valType ValueType) (any, bool) {
	store.mu.RLock()
	item, ok := store.data[key]
	store.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if item.valType != valType {
		return nil, false
	}

	if item.expireAt != nil && *item.expireAt < time.Now().UnixMilli() {
		store.mu.Lock()
		delete(store.data, key)
		store.mu.Unlock()
		return nil, false
	}

	return item.val, true
}

func (store *Store) Set(key string, val any, valType ValueType, expireAt *int64) {
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
