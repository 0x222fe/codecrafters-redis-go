package command

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func getHandler(state *state.AppState, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Usage: GET <key>")
	}

	value, exists := state.GetStore().Get(args[0])
	if !exists {
		return resp.RESPNilArray.Encode(), nil
	}

	encoded := resp.NewRESPString(value).Encode()

	return encoded, nil
}
