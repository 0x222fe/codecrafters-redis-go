package command

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils"
)

func keysHandler(state *state.AppState, args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("keys requires at least one argument")
	}

	if args[0] != "*" {
		return nil, errors.New("only wildcard '*' is supported")
	}

	keys := state.GetStore().Keys()

	result := utils.EncodeStringSliceToRESP(keys)

	return result, nil
}
