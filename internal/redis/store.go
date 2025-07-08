package redis

import "sync"

var (
	mu    = &sync.RWMutex{}
	store = make(map[string]string)
)

func getStore(key string) (string, bool) {
	mu.RLock()
	defer mu.RUnlock()

	val, ok := store[key]

	if !ok {
		return "", false
	}

	return val, true
}

func setStore(key string, val string) {
	mu.Lock()
	defer mu.Unlock()
	store[key] = val
}
