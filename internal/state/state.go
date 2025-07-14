package state

import (
	"context"
	"fmt"
	"sync"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/config"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
	"github.com/google/uuid"
)

type Replica struct {
	Client     *client.Client
	Offset     int
	OffsetChan chan int
	Ctx        context.Context
	Cancel     context.CancelFunc
}

type AppState struct {
	mu       sync.RWMutex
	cfg      *config.Config
	store    *store.Store
	replicas map[uuid.UUID]*Replica
	state    *State
}

func NewAppState(s *State, cfg *config.Config, store *store.Store) *AppState {
	return &AppState{
		cfg:      cfg,
		store:    store,
		state:    s,
		replicas: make(map[uuid.UUID]*Replica),
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

func (s *AppState) AddReplica(c *client.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	s.replicas[c.ID] = &Replica{
		Client:     c,
		Offset:     0,
		OffsetChan: make(chan int, 1),
		Ctx:        ctx,
		Cancel:     cancel,
	}
	fmt.Printf("Replica connected: %s\n", c.RemoteAddr().String())
}

func (s *AppState) RemoveReplica(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if r, exists := s.replicas[id]; exists {
		r.Cancel()

		delete(s.replicas, id)
		fmt.Printf("Replica disconnected: %s\n", r.Client.RemoteAddr().String())
	}
}

func (s *AppState) GetReplica(id uuid.UUID) (*Replica, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	replica, exists := s.replicas[id]
	return replica, exists
}

func (s *AppState) GetReplicas() []*Replica {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reps := make([]*Replica, 0, len(s.replicas))
	for _, r := range s.replicas {
		reps = append(reps, r)
	}
	return reps
}
