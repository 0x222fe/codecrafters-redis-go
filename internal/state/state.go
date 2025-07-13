package state

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/0x222fe/codecrafters-redis-go/internal/config"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

type replica struct {
	conn *net.TCPConn
	mu   *sync.Mutex
}

type AppState struct {
	mu       sync.RWMutex
	cfg      *config.Config
	store    *store.Store
	replicas map[*net.TCPConn]replica
	state    *State
}

func NewAppState(s *State, cfg *config.Config, store *store.Store) *AppState {
	return &AppState{
		cfg:      cfg,
		store:    store,
		state:    s,
		replicas: make(map[*net.TCPConn]replica),
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

func (s *AppState) AddReplica(conn *net.TCPConn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.replicas[conn] = replica{
		conn: conn,
		mu:   &sync.Mutex{},
	}
	fmt.Printf("Replica connected: %s\n", conn.RemoteAddr().String())
}

func (s *AppState) RemoveReplica(conn *net.TCPConn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.replicas[conn]; exists {
		delete(s.replicas, conn)
		fmt.Printf("Replica disconnected: %s\n", conn.RemoteAddr().String())
	}
}

func (s *AppState) IterateReplicas(f func(conn io.Writer)) {
	s.mu.RLock()
	reps := make([]replica, 0, len(s.replicas))
	for _, rep := range s.replicas {
		reps = append(reps, rep)
	}
	s.mu.RUnlock()

	for _, rep := range reps {
		rep.mu.Lock()
		f(rep.conn)
		rep.mu.Unlock()
	}
}
