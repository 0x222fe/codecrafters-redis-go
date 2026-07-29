package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func zcardHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) != 1 {
		return errors.New("ZRANGE requires exactly 1 argument")
	}

	key := args[0]

	count := s.GetStore().CountSortedSetMembers(key)

	var res = resp.NewInt(int64(count))

	writeResponse(c, res)

	return nil
}
