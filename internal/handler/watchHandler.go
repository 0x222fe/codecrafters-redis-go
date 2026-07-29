package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func watchHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) == 0 {
		return errors.New("wrong number of arguments for 'watch' command")
	}

	keys := args
	s.GetStore().Watch(keys, c.Conn.ID)

	return writeResponse(c, resp.NewString("OK"))
}
