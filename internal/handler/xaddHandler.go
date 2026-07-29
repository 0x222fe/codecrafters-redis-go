package handler

import (
	"fmt"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

func xaddHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) < 4 || len(args)%2 != 0 {
		return fmt.Errorf("XADD requires at least 4 arguments and an even number of additional arguments")
	}

	key, idStr := args[0], args[1]
	if len(key) == 0 || len(idStr) == 0 {
		return fmt.Errorf("XADD requires a key and an ID")
	}

	var stream *store.RedisStream
	v, t, ok := s.GetStore().Get(key)
	if !ok {
		stream = store.NewStream(key)
		s.GetStore().Set(key, stream, store.Stream, nil)
	} else {
		var parseOk bool
		stream, parseOk = v.(*store.RedisStream)
		if t != store.Stream || !parseOk {
			return fmt.Errorf("key is not a stream")
		}
	}

	fields := make(map[string]string)
	for i := 2; i < len(args); i += 2 {
		fields[args[i]] = args[i+1]
	}
	entry, err := stream.AddEntry(idStr, fields)
	if err != nil {
		return err
	}

	entryID := entry.ID.String()
	res := resp.NewBulkString(&entryID)
	writeResponse(c, res)

	go func() {
		store := s.GetStore()
		store.IterateStreamInsertHandlers(key, entry)
	}()

	return nil
}
