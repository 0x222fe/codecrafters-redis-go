package handler

import (
	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func multiHandler(c *client.Client, s *state.AppState, args []string) error {
	c.StartTransaction()
	res := resp.NewString("OK")
	writeResponse(c, res)
	return nil
}
