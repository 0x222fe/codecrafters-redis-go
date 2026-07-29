package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

func getHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) != 1 {
		return errors.New("Usage: GET <key>")
	}

	var res resp.RESPValue
	value, ok := s.GetStore().GetExact(args[0], store.String)
	str, parseOk := value.(string)
	if !ok || !parseOk {
		res = resp.RESPNilBulkString
	} else {
		res = resp.NewBulkString(&str)
	}

	return writeResponse(c, res)
}
