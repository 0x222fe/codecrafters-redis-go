package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

func llenHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) != 1 {
		return errors.New("LLEN requires exactly 1 argument")
	}
	key := args[0]

	v, _, ok := s.GetStore().Get(key)
	if !ok {
		writeResponse(c, resp.NewInt(0))
		return nil
	}

	list, ok := v.(*store.RedisList)
	if !ok {
		return store.ERRWrongType
	}

	writeResponse(c, resp.NewInt(int64(list.Len())))
	return nil
}
