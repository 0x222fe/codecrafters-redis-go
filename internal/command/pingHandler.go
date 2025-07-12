package command

import (
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func pingHandler(state *state.AppState, args []string) ([]byte, error) {
	return resp.NewRESPString("PONG").Encode(), nil
}
