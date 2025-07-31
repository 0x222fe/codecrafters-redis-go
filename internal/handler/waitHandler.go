package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils"
	"github.com/google/uuid"
)

func waitHandler(req *request.Request, args []string) error {
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
		return writeResponse(req, resp.NewRESPInt(0))
	}

	ctx, cancel := context.WithTimeout(req.Ctx, time.Duration(timeoutMillis)*time.Millisecond)
	defer cancel()

	command := utils.BulkStringsToRESPArray([]string{"REPLCONF", "GETACK", "*"})

	replicas := req.State.GetReplicas()
	acked, jobs := make(map[uuid.UUID]struct{}, len(replicas)), make(map[uuid.UUID]struct{}, len(replicas))

	syncedChan, doneChan := make(chan uuid.UUID, len(replicas)), make(chan uuid.UUID, len(replicas))
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	masterOffset := 0
	req.State.ReadState(func(s state.State) {
		masterOffset = s.ReplicationOffset
	})
	for _, rep := range replicas {
		if rep.Offset >= masterOffset {
			acked[rep.Client.ID] = struct{}{}
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
			replicas := req.State.GetReplicas()
			for _, r := range replicas {
				if _, ok := acked[r.Client.ID]; ok {
					continue
				}

				if _, ok := jobs[r.Client.ID]; ok {
					continue
				}

				_, err := r.Client.Write(command.Encode())
				if err != nil {
					fmt.Printf("Error writing to replica %s: %v\n", r.Client.ID, err)
					continue
				}

				go getRepOffsetUpdate(ctx, req, r, syncedChan, doneChan)

				jobs[r.Client.ID] = struct{}{}
			}
		}
	}

	ackCount := int64(len(acked))
	return writeResponse(req, resp.NewRESPInt(int64(ackCount)))
}

func getRepOffsetUpdate(ctx context.Context, req *request.Request, rep *state.Replica, syncedChan chan uuid.UUID, doneChan chan uuid.UUID) {
	defer func() { doneChan <- rep.Client.ID }()

	select {
	case count := <-rep.OffsetChan:
		masterOffset := 0
		req.State.ReadState(func(s state.State) {
			masterOffset = s.ReplicationOffset
		})

		if count >= masterOffset {
			syncedChan <- rep.Client.ID
		}
	case <-ctx.Done():
	case <-rep.Ctx.Done():
	}
}