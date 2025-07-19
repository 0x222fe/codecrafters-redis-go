package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

func lpushHandler(req *request.Request, args []string) error {
	if len(args) < 2 {
		return errors.New("LPUSH requires at least 2 arguments")
	}

	key, items := args[0], args[1:]

	var list *store.RedisList
	v, t, ok := req.State.GetStore().Get(key)
	if !ok {
		list = store.NewList()
		req.State.GetStore().Set(key, list, store.List, nil)
	} else {
		l, parseOk := v.(*store.RedisList)
		if t != store.List || !parseOk {
			return errors.New("key is not a list")
		}
		list = l
	}

	count := list.LPush(items...)

	s := req.State.GetStore()
	for _, item := range items {
		s.NotifyListPush(key, item)
	}

	writeResponse(req, resp.NewRESPInt(int64(count)))
	return nil
}
