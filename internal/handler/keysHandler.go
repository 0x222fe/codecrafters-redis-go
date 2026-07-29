package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
)

func keysHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) == 0 {
		return errors.New("keys requires at least one argument")
	}

	if args[0] != "*" {
		return errors.New("only wildcard '*' is supported")
	}

	keys := s.GetStore().Keys()

	res := resputil.BulkStringsToRESPArray(keys)

	return writeResponse(c, res)
}
