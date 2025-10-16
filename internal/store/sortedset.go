package store

import (
	"sync"

	"github.com/0x222fe/codecrafters-redis-go/internal/types/sortedset"
)

type SortedSetMember struct {
	Score  float64
	Member string
}

type sortedSetEntry struct {
	mu  sync.RWMutex
	set *sortedset.SortedSet
}

func (store *Store) AddToSortedSet(name string, members []SortedSetMember) {
	store.sortedSetMu.Lock()
	entry, ok := store.sortedSetEntries[name]
	if !ok {
		entry = &sortedSetEntry{
			set: sortedset.New(),
		}
		store.sortedSetEntries[name] = entry
	}

	entry.mu.Lock()
	store.sortedSetMu.Unlock()
	defer entry.mu.Unlock()

	for _, m := range members {
		entry.set.Set(m.Member, m.Score)
	}
}
