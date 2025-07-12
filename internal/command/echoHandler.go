package command

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func echoHandler(state *state.AppState, args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("ECHO requires at least one argument")
	}

	encoded := resp.NewRESPString(args[0]).Encode()

	return encoded, nil
}
