package state

import (
	"context"
	"fmt"
	"sync"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/config"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
	"github.com/google/uuid"
)

const (
	subChanBufSize = 16
)

type Replica struct {
	Client     *client.Client
	Offset     int
	OffsetChan chan int
	Ctx        context.Context
	Cancel     context.CancelFunc
}

type Subscriber struct {
	Client   *client.Client
	Channels map[string]struct{}
	MsgChan  chan PubSubMsg
	Ctx      context.Context
	Cancel   context.CancelFunc
}

type PubSubMsg struct {
	Channel string
	Payload []byte
}

type AppState struct {
	mu          sync.RWMutex
	cfg         *config.Config
	store       *store.Store
	replicas    map[uuid.UUID]*Replica
	subscribers map[uuid.UUID]*Subscriber
	channelSubs map[string]map[uuid.UUID]*Subscriber
	state       *State
}

func NewAppState(s *State, cfg *config.Config, store *store.Store) *AppState {
	appState := &AppState{
		cfg:         cfg,
		store:       store,
		state:       s,
		replicas:    make(map[uuid.UUID]*Replica),
		subscribers: make(map[uuid.UUID]*Subscriber),
		channelSubs: make(map[string]map[uuid.UUID]*Subscriber),
	}

	return appState
}

type State struct {
	IsReplica           bool
	MasterReplicationID string
	ReplicationID       string
	ReplicationOffset   int
	User                string
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

// SetStore should only be called during replication/master handshake
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

func (s *AppState) AddSubscriber(c *client.Client, channel string) *Subscriber {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub, ok := s.subscribers[c.ID]
	if !ok {
		ctx, cancel := context.WithCancel(context.Background())
		sub = &Subscriber{
			Client:   c,
			Ctx:      ctx,
			Channels: make(map[string]struct{}),
			Cancel:   cancel,
			MsgChan:  make(chan PubSubMsg, subChanBufSize),
		}
		s.subscribers[c.ID] = sub
		go func() {
			for {
				select {
				case msg := <-sub.MsgChan:
					c.WriteResp(resputil.BulkStringsToRESPArray([]string{
						"message",
						msg.Channel,
						string(msg.Payload),
					}))
				case <-sub.Ctx.Done():
					return
				}
			}
		}()
	}
	sub.Channels[channel] = struct{}{}

	chanMap, ok := s.channelSubs[channel]
	if !ok {
		chanMap = make(map[uuid.UUID]*Subscriber)
		s.channelSubs[channel] = chanMap
	}

	chanMap[c.ID] = sub

	fmt.Printf("Subscriber connected: %s\n", c.RemoteAddr().String())

	return sub
}

func (s *AppState) UnsubChannel(id uuid.UUID, channel string) *Subscriber {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub, ok := s.subscribers[id]
	if !ok {
		return nil
	}

	chanMap, ok := s.channelSubs[channel]
	if !ok {
		return sub
	}

	_, ok = chanMap[id]
	if !ok {
		return sub
	}

	delete(chanMap, id)
	delete(sub.Channels, channel)

	return sub
}

func (s *AppState) RemoveSubscriber(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sub, exists := s.subscribers[id]; exists {
		sub.Cancel()

		for channel := range sub.Channels {
			delete(s.channelSubs[channel], id)
		}

		delete(s.subscribers, id)
		fmt.Printf("Subscriber disconnected: %s\n", sub.Client.RemoteAddr().String())
	}
}

func (s *AppState) Publish(channel string, payload []byte) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sent := 0
	chanMap, ok := s.channelSubs[channel]
	if !ok {
		return sent
	}

	for _, sub := range chanMap {
		select {
		case sub.MsgChan <- PubSubMsg{Channel: channel, Payload: payload}:
			sent++
		default:
		}
	}

	return sent
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