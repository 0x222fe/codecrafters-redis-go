package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func zremHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) != 2 {
		return errors.New("ZSCORE requires exactly 2 arguments")
	}

	key, member := args[0], args[1]

	ok := s.GetStore().RemoveSortedSetMember(key, member)

	var res resp.RESPValue
	if !ok {
		res = resp.NewInt(0)
	} else {
		res = resp.NewInt(1)
	}

	writeResponse(c, res)
	return nil
}
