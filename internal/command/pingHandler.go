package command

import (
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func pingHandler(state *state.AppState, args []string) ([]byte, error) {
	return []byte("+PONG\r\n"), nil
}
