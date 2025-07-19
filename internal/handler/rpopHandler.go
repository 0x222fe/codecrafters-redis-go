package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

func rpopHandler(req *request.Request, args []string) error {
	if len(args) != 1 {
		return errors.New("RPOP requires exactly 1 argument")
	}
	key := args[0]

	v, _, ok := req.State.GetStore().Get(key)
	if !ok {
		writeResponse(req, resp.RESPNilBulkString)
		return nil
	}
	list, ok := v.(*store.RedisList)
	if !ok {
		return store.ERRWrongType
	}

	val, ok := list.RPop()
	if !ok {
		writeResponse(req, resp.RESPNilBulkString)
		return nil
	}

	writeResponse(req, resp.NewRESPBulkString(&val))
	return nil
}
