package handler

import (
	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
)

func pingHandler(c *client.Client, s *state.AppState, args []string) error {
	if c.Propagated {
		return nil
	}
	if c.SubMode {
		return writeResponse(c, resputil.BulkStringsToRESPArray([]string{"pong", ""}))
	}
	return writeResponse(c, resp.NewString("PONG"))
}
