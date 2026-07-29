package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func zrankHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) != 2 {
		return errors.New("ZRANK requires exactly 2 arguments")
	}

	key, member := args[0], args[1]

	rank, ok := s.GetStore().QuerySortedSetRank(key, member)

	var res resp.RESPValue
	if !ok {
		res = resp.RESPNilBulkString
	} else {
		res = resp.NewInt(int64(rank))
	}

	writeResponse(c, res)
	return nil
}
