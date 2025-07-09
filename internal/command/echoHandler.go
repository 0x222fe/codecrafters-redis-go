package command

import (
	"errors"

	"github.com/codecrafters-io/redis-starter-go/internal/state"
)

func echoHandler(state *state.AppState, args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("ECHO requires at least one argument")
	}

	response := "+" + args[0] + "\r\n"
	return []byte(response), nil
}
