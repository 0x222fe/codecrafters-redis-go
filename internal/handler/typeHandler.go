package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func typeHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) != 1 {
		return errors.New("TYPE requires exactly one argument")
	}

	key := args[0]
	valType := s.GetStore().Type(key)
	return writeResponse(c, resp.NewString(valType))
}
