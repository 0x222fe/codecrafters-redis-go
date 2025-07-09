package command

import "github.com/codecrafters-io/redis-starter-go/internal/state"

func pingHandler(*state.AppState, []string) ([]byte, error) {
	return []byte("+PONG\r\n"), nil
}
