package handler

import (
	"errors"
	"strconv"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

func incrHandler(req *request.Request, args []string) error {
	if len(args) < 1 {
		return errors.New("INCR requires at least 1 argument")
	}

	key := args[0]

	val, ok := req.State.GetStore().Get(key, store.String)
	if !ok {
		req.State.GetStore().Set(key, "1", store.String, nil)
		res := resp.NewRESPInt(1)
		writeResponse(req, res)
		return nil
	}

	strVal, ok := val.(string)
	if !ok {
		return errors.New("value is not a string")
	}

	n, err := strconv.ParseInt(strVal, 10, 64)
	if err != nil {
		return errors.New("value is not an integer or out of range")
	}

	n++
	req.State.GetStore().Set(key, strconv.FormatInt(n, 10), store.String, nil)
	res := resp.NewRESPInt(n)
	writeResponse(req, res)
	return nil
}
