package handler

import (
	"errors"
	"strconv"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
)

func lpopHandler(req *request.Request, args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return errors.New("LPOP takes 1 or 2 arguments")
	}

	key, count := args[0], 1
	if len(args) == 2 {
		c, err := strconv.Atoi(args[1])
		if err != nil {
			return errors.New("invalid count")
		}
		if c <= 0 {
			return errors.New("value is out of range, must be positive")
		}
		count = c
	}

	v, _, ok := req.State.GetStore().Get(key)
	if !ok {
		writeResponse(req, resp.RESPNilBulkString)
		return nil
	}
	list, ok := v.(*store.RedisList)
	if !ok {
		return store.ERRWrongType
	}

	vals, ok := list.LPop(count)
	if !ok {
		writeResponse(req, resp.RESPNilBulkString)
		return nil
	}

	if count == 1 {
		writeResponse(req, resp.NewBulkString(&vals[0]))
		return nil
	}

	writeResponse(req, resputil.BulkStringsToRESPArray(vals))
	return nil
}
