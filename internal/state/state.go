package state

import (
	"sync"

	"github.com/0x222fe/codecrafters-redis-go/internal/config"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
	"github.com/0x222fe/codecrafters-redis-go/internal/user"
	"github.com/google/uuid"
)

type AppState struct {
	mu           sync.RWMutex
	cfg          *config.Config
	store        *store.Store
	replicaState *ReplicaState
	replicas     map[uuid.UUID]*Replica
	subscribers  map[uuid.UUID]*Subscriber
	channelSubs  map[string]map[uuid.UUID]*Subscriber
	users        map[string]*user.User
}

func NewAppState(s *ReplicaState, cfg *config.Config, store *store.Store) *AppState {
	defaultUser := user.New(user.DefaultUserName)
	defaultUser.AddFlag(user.FlagNoPass)

	appState := &AppState{
		cfg:          cfg,
		store:        store,
		replicaState: s,
		replicas:     make(map[uuid.UUID]*Replica),
		subscribers:  make(map[uuid.UUID]*Subscriber),
		channelSubs:  make(map[string]map[uuid.UUID]*Subscriber),
		users: map[string]*user.User{
			user.DefaultUserName: defaultUser,
		},
	}

	return appState
}

func (s *AppState) ReadState(f func(s ReplicaState)) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f(*s.replicaState)
}

func (s *AppState) WriteState(f func(s *ReplicaState)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f(s.replicaState)
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
