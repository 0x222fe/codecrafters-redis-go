package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

func lpushHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) < 2 {
		return errors.New("LPUSH requires at least 2 arguments")
	}

	key, items := args[0], args[1:]

	var list *store.RedisList
	v, t, ok := s.GetStore().Get(key)
	if !ok {
		list = store.NewList()
		s.GetStore().Set(key, list, store.List, nil)
	} else {
		l, parseOk := v.(*store.RedisList)
		if t != store.List || !parseOk {
			return errors.New("key is not a list")
		}
		list = l
	}

	count := list.LPush(items...)

	store := s.GetStore()
	for _, item := range items {
		store.NotifyListPush(key, item)
	}

	writeResponse(c, resp.NewInt(int64(count)))
	return nil
}
