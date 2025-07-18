package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

func rpushHandler(req *request.Request, args []string) error {
	if len(args) < 2 {
		return errors.New("RPUSH requires at least 2 arguments")
	}

	key, item := args[0], args[1]

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

	count := list.Push(item)

	writeResponse(req, resp.NewRESPInt(int64(count)))
	return nil
}
