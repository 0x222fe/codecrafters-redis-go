package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
	"github.com/google/uuid"
)

func waitHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) < 2 {
		return errors.New("WAIT requires at least two arguments")
	}
	repCount, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid replication count: %w", err)
	}
	if repCount < 0 {
		return errors.New("replication count cannot be negative")
	}

	timeoutMillis, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timeout: %w", err)
	}
	if timeoutMillis < 0 {
		return errors.New("timeout cannot be negative")
	}

	if repCount == 0 {
		return writeResponse(c, resp.NewInt(0))
	}

	ctx, cancel := context.WithTimeout(c.Ctx, time.Duration(timeoutMillis)*time.Millisecond)
	defer cancel()

	command := resputil.BulkStringsToRESPArray([]string{"REPLCONF", "GETACK", "*"})

	replicas := s.GetReplicas()
	acked, jobs := make(map[uuid.UUID]struct{}, len(replicas)), make(map[uuid.UUID]struct{}, len(replicas))

	syncedChan, doneChan := make(chan uuid.UUID, len(replicas)), make(chan uuid.UUID, len(replicas))
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	masterOffset := 0
	s.ReadState(func(st state.ReplicaState) {
		masterOffset = st.ReplicationOffset
	})
	for _, rep := range replicas {
		if rep.Offset >= masterOffset {
			acked[rep.Conn.ID] = struct{}{}
		}
	}

outter:
	for {
		if int64(len(acked)) >= repCount {
			break
		}

		select {
		case <-ctx.Done():
			break outter
		case id := <-syncedChan:
			acked[id] = struct{}{}
			if int64(len(acked)) >= repCount {
				break outter
			}
		case id := <-doneChan:
			delete(jobs, id)
		case <-ticker.C:
			replicas := s.GetReplicas()
			for _, r := range replicas {
				if _, ok := acked[r.Conn.ID]; ok {
					continue
				}

				if _, ok := jobs[r.Conn.ID]; ok {
					continue
				}

				_, err := r.Conn.Write(command.Bytes())
				if err != nil {
					fmt.Printf("Error writing to replica %s: %v\n", r.Conn.ID, err)
					continue
				}

				go getRepOffsetUpdate(ctx, c, s, r, syncedChan, doneChan)

				jobs[r.Conn.ID] = struct{}{}
			}
		}
	}

	ackCount := int64(len(acked))
	return writeResponse(c, resp.NewInt(int64(ackCount)))
}

func getRepOffsetUpdate(ctx context.Context, c *client.Client, s *state.AppState, rep *state.Replica, syncedChan chan uuid.UUID, doneChan chan uuid.UUID) {
	defer func() { doneChan <- rep.Conn.ID }()

	select {
	case count := <-rep.OffsetChan:
		masterOffset := 0
		s.ReadState(func(st state.ReplicaState) {
			masterOffset = st.ReplicationOffset
		})

		if count >= masterOffset {
			syncedChan <- rep.Conn.ID
		}
	case <-ctx.Done():
	case <-rep.Ctx.Done():
	}
}
