package store

import (
	"errors"
	"sync"
	"time"

	"github.com/0x222fe/codecrafters-redis-go/internal/types/orderedmap"
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

type ListPushChanRgistry = orderedmap.OrderedMap[uuid.UUID, chan string]

type WatchRegistry map[uuid.UUID]map[string]uint32

type Store struct {
	dataMu sync.RWMutex
	data   map[string]StoreItem

	sortedSetMu      sync.RWMutex
	sortedSetEntries map[string]*sortedSetEntry

	streamMu         sync.RWMutex
	streamRegistries map[string]StreamInsertHandlerRegistry

	listMu         sync.RWMutex
	listRegistries map[string]*ListPushChanRgistry

	watchMu       sync.RWMutex
	watchRegistry WatchRegistry
}

func NewStore() *Store {
	return &Store{
		data:             make(map[string]StoreItem),
		sortedSetEntries: make(map[string]*sortedSetEntry),
		streamRegistries: make(map[string]StreamInsertHandlerRegistry),
		listRegistries:   make(map[string]*ListPushChanRgistry),
		watchRegistry:    make(WatchRegistry),
	}
}

type StoreItem struct {
	val        any
	valType    ValueType
	expireAt   *int64
	modCounter uint32
}

var (
	ERRWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
)

func (store *Store) Get(key string) (any, ValueType, bool) {
	store.dataMu.Lock()
	defer store.dataMu.Unlock()

	item, ok := store.data[key]

	if !ok || item.val == nil {
		return nil, None, false
	}

	if item.expireAt != nil && *item.expireAt < time.Now().UnixMilli() {
		store.deleteLocked(key)
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
	store.dataMu.Lock()
	defer store.dataMu.Unlock()

	var counter uint32
	item, ok := store.data[key]
	if ok {
		counter = item.modCounter
	}

	store.data[key] = StoreItem{
		val:        val,
		valType:    valType,
		expireAt:   expireAt,
		modCounter: counter + 1,
	}
}

func (store *Store) Delete(key string) {
	store.dataMu.Lock()
	defer store.dataMu.Unlock()
	store.deleteLocked(key)
}

func (store *Store) Type(key string) string {
	_, vType, ok := store.Get(key)
	if !ok {
		return string(None)
	}

	return string(vType)
}

func (store *Store) Keys() []string {
	store.dataMu.RLock()
	defer store.dataMu.RUnlock()

	keys := make([]string, 0, len(store.data))
	for key, item := range store.data {
		if item.val != nil {
			keys = append(keys, key)
		}
	}
	return keys
}

func (store *Store) RegisterStreamInsertHandler(streamKey string, connID uuid.UUID, handler StreamInsertHandler) {
	store.streamMu.Lock()
	defer store.streamMu.Unlock()

	registry, ok := store.streamRegistries[streamKey]
	if !ok {
		registry = make(StreamInsertHandlerRegistry)
		store.streamRegistries[streamKey] = registry
	}

	registry[connID] = handler
}

func (store *Store) UnregisterStreamInsertHandler(streamKey string, handlerID uuid.UUID) {
	store.streamMu.Lock()
	defer store.streamMu.Unlock()
	if registry, ok := store.streamRegistries[streamKey]; ok {
		delete(registry, handlerID)
	}
}

func (store *Store) IterateStreamInsertHandlers(streamKey string, entry *StreamEntry) {
	store.streamMu.RLock()
	defer store.streamMu.RUnlock()
	registry, ok := store.streamRegistries[streamKey]
	if !ok {
		return
	}

	for _, handler := range registry {
		handler(entry)
	}
}

func (store *Store) RegisterListPushHandler(listKey string, connID uuid.UUID, ch chan string) {
	store.listMu.Lock()
	defer store.listMu.Unlock()

	registry, ok := store.listRegistries[listKey]
	if !ok {
		registry = orderedmap.New[uuid.UUID, chan string]()
		store.listRegistries[listKey] = registry
	}

	registry.Set(connID, ch)
}

func (store *Store) UnregisterListPushHandler(listKey string, connID uuid.UUID) {
	store.listMu.Lock()
	defer store.listMu.Unlock()

	if registry, ok := store.listRegistries[listKey]; ok {
		registry.Delete(connID)
	}
}

func (store *Store) NotifyListPush(listKey string, value string) {
	store.listMu.RLock()
	defer store.listMu.RUnlock()

	reg, ok := store.listRegistries[listKey]
	if !ok {
		return
	}

	ch, ok := reg.Peek()
	if !ok {
		return
	}

	select {
	case ch <- value:
	default:
	}
}

func (store *Store) Watch(keys []string, connID uuid.UUID) {
	store.watchMu.Lock()
	store.dataMu.RLock()
	defer store.watchMu.Unlock()
	defer store.dataMu.RUnlock()

	clientMap, ok := store.watchRegistry[connID]
	if !ok {
		clientMap = make(map[string]uint32)
		store.watchRegistry[connID] = clientMap
	}

	for _, key := range keys {
		var currCounter uint32
		item, ok := store.data[key]
		if ok {
			currCounter = item.modCounter
		}
		clientMap[key] = currCounter
	}
}

func (store *Store) Unwatch(connID uuid.UUID) {
	store.watchMu.Lock()
	defer store.watchMu.Unlock()

	delete(store.watchRegistry, connID)
}

func (store *Store) WatchesValid(connID uuid.UUID) bool {
	store.watchMu.Lock()
	store.dataMu.RLock()
	defer store.watchMu.Unlock()
	defer store.dataMu.RUnlock()

	clientMap, ok := store.watchRegistry[connID]
	if !ok {
		return true
	}

	for key := range clientMap {
		item := store.data[key]
		if item.modCounter != clientMap[key] {
			return false
		}
	}
	return true
}

func (store *Store) deleteLocked(key string) {
	item, ok := store.data[key]
	if ok {
		item.val = nil
		item.modCounter += 1
		store.data[key] = item
	}
}
