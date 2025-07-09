package command

import (
	"errors"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/state"
)

func keysHandler(state *state.AppState, args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("keys requires at least one argument")
	}

	if args[0] != "*" {
		return nil, errors.New("only wildcard '*' is supported")
	}
	keys := state.Store.Keys()

	result, err := resp.RESPEncode(keys)
	if err != nil {
		return nil, fmt.Errorf("failed to encode keys into RESP format: %w", err)
	}

	return result, nil
}
