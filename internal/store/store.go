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

type Store struct {
	dataMu sync.RWMutex
	data   map[string]StoreItem

	sortedSetMu      sync.RWMutex
	sortedSetEntries map[string]*sortedSetEntry

	streamMu         sync.RWMutex
	streamRegistries map[string]StreamInsertHandlerRegistry

	listMu         sync.RWMutex
	listRegistries map[string]*ListPushChanRgistry
}

func NewStore() *Store {
	return &Store{
		data:             make(map[string]StoreItem),
		sortedSetEntries: make(map[string]*sortedSetEntry),
		streamRegistries: make(map[string]StreamInsertHandlerRegistry),
		listRegistries:   make(map[string]*ListPushChanRgistry),
	}
}

type StoreItem struct {
	val      any
	valType  ValueType
	expireAt *int64
}

var (
	ERRWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
)

func (store *Store) Get(key string) (any, ValueType, bool) {
	store.dataMu.Lock()
	defer store.dataMu.Unlock()

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
	store.dataMu.Lock()
	defer store.dataMu.Unlock()

	store.data[key] = StoreItem{
		val:      val,
		valType:  valType,
		expireAt: expireAt,
	}
}

func (store *Store) Type(key string) string {
	store.dataMu.RLock()
	defer store.dataMu.RUnlock()

	item, ok := store.data[key]
	if !ok {
		return string(None)
	}

	if item.expireAt != nil && *item.expireAt < time.Now().UnixMilli() {
		store.dataMu.Lock()
		delete(store.data, key)
		store.dataMu.Unlock()
		return string(None)
	}

	return string(item.valType)
}

func (store *Store) Keys() []string {
	store.dataMu.RLock()
	defer store.dataMu.RUnlock()

	keys := make([]string, 0, len(store.data))
	for key := range store.data {
		keys = append(keys, key)
	}
	return keys
}

func (store *Store) RegisterStreamInsertHandler(streamKey string, clientID uuid.UUID, handler StreamInsertHandler) {
	store.streamMu.Lock()
	defer store.streamMu.Unlock()

	registry, ok := store.streamRegistries[streamKey]
	if !ok {
		registry = make(StreamInsertHandlerRegistry)
		store.streamRegistries[streamKey] = registry
	}

	registry[clientID] = handler
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

func (store *Store) RegisterListPushHandler(listKey string, clientID uuid.UUID, ch chan string) {
	store.listMu.Lock()
	defer store.listMu.Unlock()

	registry, ok := store.listRegistries[listKey]
	if !ok {
		registry = orderedmap.New[uuid.UUID, chan string]()
		store.listRegistries[listKey] = registry
	}

	registry.Set(clientID, ch)
}

func (store *Store) UnregisterListPushHandler(listKey string, clientID uuid.UUID) {
	store.listMu.Lock()
	defer store.listMu.Unlock()

	if registry, ok := store.listRegistries[listKey]; ok {
		registry.Delete(clientID)
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
