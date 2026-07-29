package handler

import (
	"errors"
	"strconv"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

func incrHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) < 1 {
		return errors.New("INCR requires at least 1 argument")
	}

	key := args[0]

	val, ok := s.GetStore().GetExact(key, store.String)
	if !ok {
		s.GetStore().Set(key, "1", store.String, nil)
		res := resp.NewInt(1)
		writeResponse(c, res)
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
	s.GetStore().Set(key, strconv.FormatInt(n, 10), store.String, nil)
	res := resp.NewInt(n)
	writeResponse(c, res)
	return nil
}
