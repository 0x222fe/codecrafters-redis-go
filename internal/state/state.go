package state

import (
	"sync"

	"github.com/0x222fe/codecrafters-redis-go/internal/config"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

type AppState struct {
	mu sync.RWMutex
	// Cfg               *config.Config
	// Store             *store.Store
	// IsReplica         bool
	// ReplicationID     string
	// ReplicationOffset int
	state *State
}

func NewAppState(s *State, cfg *config.Config, store *store.Store) *AppState {
	s.cfg = cfg
	s.store = store

	return &AppState{
		state: s,
	}
}

type State struct {
	cfg               *config.Config
	store             *store.Store
	IsReplica         bool
	ReplicationID     string
	ReplicationOffset int
}

func (s *AppState) ReadState(f func(s State)) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f(*s.state)
}

func (s *AppState) WriteState(f func(s *State)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f(s.state)
}

func (s *AppState) GetStore() *store.Store {
	return s.state.store
}

func (s *AppState) ReadCfg() config.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.state.cfg
}
