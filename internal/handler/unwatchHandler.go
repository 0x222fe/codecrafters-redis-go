package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func unwatchHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) != 0 {
		return errors.New("wrong number of arguments for 'unwatch' command")
	}

	s.GetStore().Unwatch(c.Conn.ID)

	return writeResponse(c, resp.NewString("OK"))
}
