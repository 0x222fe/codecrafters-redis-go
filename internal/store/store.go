package store

import (
	"sync"
	"time"

	"github.com/google/uuid"
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

type StreamInsertHandler func(entry *StreamEntry)
type StreamInsertHandlerRegistry map[uuid.UUID]StreamInsertHandler

type Store struct {
	mu               sync.RWMutex
	data             map[string]StoreItem
	streams          map[string]*RedisStream
	streamRegistries map[string]StreamInsertHandlerRegistry
}

func NewStore() *Store {
	return &Store{
		data:             make(map[string]StoreItem),
		streams:          make(map[string]*RedisStream),
		streamRegistries: make(map[string]StreamInsertHandlerRegistry),
	}
}

type StoreItem struct {
	val      any
	valType  ValueType
	expireAt *int64
}

func (store *Store) Get(key string) (any, ValueType, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()

	item, ok := store.data[key]

	if !ok {
		return nil, None, false
	}

	if item.expireAt != nil && *item.expireAt < time.Now().UnixMilli() {
		delete(store.data, key)
		return nil, None, false
	}

	return item.val, item.valType, true
}

func (store *Store) GetExact(key string, valType ValueType) (any, bool) {
	item, vType, ok := store.Get(key)
	if !ok {
		return nil, false
	}

	if vType != valType {
		return nil, false
	}

	return item, true
}

func (store *Store) Set(key string, val any, valType ValueType, expireAt *int64) {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.data[key] = StoreItem{
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

func (store *Store) RegisterStreamInsertHandler(streamKey string, handler StreamInsertHandler) uuid.UUID {
	store.mu.Lock()
	defer store.mu.Unlock()

	registry, ok := store.streamRegistries[streamKey]
	if !ok {
		registry = make(StreamInsertHandlerRegistry)
		store.streamRegistries[streamKey] = registry
	}

	id := uuid.New()
	registry[id] = handler
	return id
}

func (store *Store) UnregisterStreamInsertHandler(streamKey string, handlerID uuid.UUID) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if registry, ok := store.streamRegistries[streamKey]; ok {
		delete(registry, handlerID)
	}
}

func (store *Store) IterateStreamInsertHandlers(streamKey string, entry *StreamEntry) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	registry, ok := store.streamRegistries[streamKey]
	if !ok {
		return
	}

	for _, handler := range registry {
		handler(entry)
	}
}
