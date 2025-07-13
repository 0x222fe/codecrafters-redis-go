package state

import (
	"fmt"
	"sync"

	"github.com/0x222fe/codecrafters-redis-go/internal/cnn"
	"github.com/0x222fe/codecrafters-redis-go/internal/config"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

type AppState struct {
	mu       sync.RWMutex
	cfg      *config.Config
	store    *store.Store
	replicas map[*cnn.Connection]struct{}
	state    *State
}

func NewAppState(s *State, cfg *config.Config, store *store.Store) *AppState {
	return &AppState{
		cfg:      cfg,
		store:    store,
		state:    s,
		replicas: make(map[*cnn.Connection]struct{}),
	}
}

type State struct {
	IsReplica           bool
	MasterReplicationID string
	ReplicationID       string
	ReplicationOffset   int
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

func (s *AppState) SetStore(store *store.Store) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store = store
}

func (s *AppState) GetStore() *store.Store {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.store
}

func (s *AppState) ReadCfg() config.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.cfg
}

func (s *AppState) AddReplica(conn *cnn.Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.replicas[conn] = struct{}{}
	fmt.Printf("Replica connected: %s\n", conn.RemoteAddr().String())
}

func (s *AppState) RemoveReplica(conn *cnn.Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.replicas[conn]; exists {
		delete(s.replicas, conn)
		fmt.Printf("Replica disconnected: %s\n", conn.RemoteAddr().String())
	}
}

func (s *AppState) IterateReplicas(f func(conn *cnn.Connection)) {
	s.mu.RLock()
	reps := make([]*cnn.Connection, 0, len(s.replicas))
	for c := range s.replicas {
		reps = append(reps, c)
	}
	s.mu.RUnlock()

	for _, c := range reps {
		f(c)
	}
}
