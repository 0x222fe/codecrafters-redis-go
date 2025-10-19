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

func (store *Store) AddToSortedSet(key string, members []SortedSetMember) int {
	store.sortedSetMu.Lock()
	entry, ok := store.sortedSetEntries[key]
	if !ok {
		entry = &sortedSetEntry{
			set: sortedset.New(),
		}
		store.sortedSetEntries[key] = entry
	}

	entry.mu.Lock()
	store.sortedSetMu.Unlock()
	defer entry.mu.Unlock()

	count := 0
	for _, m := range members {
		count += entry.set.Set(m.Member, m.Score)
	}
	return count
}

func (store *Store) QuerySortedSetRank(key string, member string) (int, bool) {
	store.sortedSetMu.RLock()
	entry, ok := store.sortedSetEntries[key]
	if !ok {
		store.sortedSetMu.RUnlock()
		return -1, false
	}

	entry.mu.RLock()
	store.sortedSetMu.RUnlock()
	defer entry.mu.RUnlock()

	rank, ok := entry.set.Rank(member)
	return rank, ok
}

func (store *Store) ListSortedSetMembersByRank(key string, start, end int) []string {
	store.sortedSetMu.RLock()
	entry, ok := store.sortedSetEntries[key]
	if !ok {
		store.sortedSetMu.RUnlock()
		return []string{}
	}

	entry.mu.RLock()
	store.sortedSetMu.RUnlock()
	defer entry.mu.RUnlock()

	return entry.set.RangeByRank(start, end)
}

func (store *Store) CountSortedSetMembers(key string) int {
	store.sortedSetMu.RLock()
	entry, ok := store.sortedSetEntries[key]
	if !ok {
		store.sortedSetMu.RUnlock()
		return 0
	}

	entry.mu.RLock()
	store.sortedSetMu.RUnlock()
	defer entry.mu.RUnlock()

	return entry.set.Len()
}

func (store *Store) QuerySortedSetScore(key, member string) (float64, bool) {
	store.sortedSetMu.RLock()
	entry, ok := store.sortedSetEntries[key]
	if !ok {
		store.sortedSetMu.RUnlock()
		return 0, false
	}

	entry.mu.RLock()
	store.sortedSetMu.RUnlock()
	defer entry.mu.RUnlock()

	return entry.set.Get(member)
}

func (store *Store) RemoveSortedSetMember(key, member string) bool {

	store.sortedSetMu.RLock()
	entry, ok := store.sortedSetEntries[key]
	if !ok {
		store.sortedSetMu.RUnlock()
		return false
	}

	entry.mu.Lock()
	store.sortedSetMu.RUnlock()
	defer entry.mu.Unlock()

	return entry.set.Remove(member)
}