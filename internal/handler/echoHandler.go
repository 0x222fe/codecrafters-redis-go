package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func echoHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) == 0 {
		return errors.New("ECHO requires at least one argument")
	}

	res := resp.NewBulkString(&args[0])

	return writeResponse(c, res)
}
