package state

import (
	"context"
	"fmt"

	"github.com/0x222fe/codecrafters-redis-go/internal/connection"
	"github.com/google/uuid"
)

type Replica struct {
	Conn       *connection.Connection
	Offset     int
	OffsetChan chan int
	Ctx        context.Context
	Cancel     context.CancelFunc
}

type ReplicaState struct {
	IsReplica           bool
	MasterReplicationID string
	ReplicationID       string
	ReplicationOffset   int
}

func (s *AppState) AddReplica(conn *connection.Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	s.replicas[conn.ID] = &Replica{
		Conn:       conn,
		Offset:     0,
		OffsetChan: make(chan int, 1),
		Ctx:        ctx,
		Cancel:     cancel,
	}
	fmt.Printf("Replica connected: %s\n", conn.RemoteAddr().String())
}

func (s *AppState) RemoveReplica(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if r, exists := s.replicas[id]; exists {
		r.Cancel()

		delete(s.replicas, id)
		fmt.Printf("Replica disconnected: %s\n", r.Conn.RemoteAddr().String())
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
