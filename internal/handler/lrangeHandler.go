package handler

import (
	"errors"
	"strconv"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
)

func lrangeHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) != 3 {
		return errors.New("LRANGE requires exactly 3 arguments")
	}

	key, startS, endS := args[0], args[1], args[2]

	start, err := strconv.Atoi(startS)
	if err != nil {
		return errors.New("invalid start")
	}

	end, err := strconv.Atoi(endS)
	if err != nil {
		return errors.New("invalid end")
	}

	v, _, ok := s.GetStore().Get(key)
	list, parseOk := v.(*store.RedisList)
	if !ok || !parseOk {
		writeResponse(c, resp.RESPEmptyArray)
		return nil
	}

	values := list.GetRange(start, end)
	res := resputil.BulkStringsToRESPArray(values)
	writeResponse(c, res)
	return nil
}
