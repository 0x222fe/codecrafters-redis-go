package command

import (
	"errors"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/state"
)

func getHandler(state *state.AppState, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Usage: GET <key>")
	}

	value, exists := state.Store.Get(args[0])
	if !exists {
		return resp.RESPNIL, nil
	}

	result, err := resp.RESPEncode(value)
	if err != nil {
		return nil, fmt.Errorf("failed to encode value into RESP format: %w", err)
	}

	return result, nil
}
