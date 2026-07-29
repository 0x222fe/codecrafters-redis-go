package handler

import (
	"errors"
	"fmt"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func zscoreHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) != 2 {
		return errors.New("ZSCORE requires exactly 2 arguments")
	}

	key, member := args[0], args[1]

	score, ok := s.GetStore().QuerySortedSetScore(key, member)

	var res resp.RESPValue
	if !ok {
		res = resp.RESPNilBulkString
	} else {
		s := fmt.Sprintf("%.17g", score)
		res = resp.NewBulkString(&s)
	}

	writeResponse(c, res)
	return nil
}
