package redis

import (
	"sync"
	"time"
)

type storeItem struct {
	val string
	exp time.Time
}

var (
	mu    = &sync.RWMutex{}
	store = make(map[string]storeItem)
)

func getStore(key string) (string, bool) {
	mu.RLock()
	item, ok := store[key]
	mu.RUnlock()

	if !ok {
		return "", false
	}

	if !item.exp.IsZero() && time.Now().After(item.exp) {
		mu.Lock()
		delete(store, key)
		mu.Unlock()
		return "", false
	}

	return item.val, true
}

func setStore(key string, val string, expMillis int64) {
	var exp time.Time
	if expMillis > 0 {
		exp = time.Now().Add(time.Duration(expMillis) * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	store[key] = storeItem{
		val: val,
		exp: exp,
	}
}
