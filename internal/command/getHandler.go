package command

import (
	"errors"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/internal/state"
)

func getHandler(state *state.AppState, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("GET requires exactly one argument")
	}

	value, exists := state.Store.Get(args[0])
	if !exists {
		return []byte("$-1\r\n"), nil
	}

	res := fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)

	return []byte(res), nil
}
