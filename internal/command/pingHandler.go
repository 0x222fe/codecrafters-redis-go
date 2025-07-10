package command

import "github.com/0x222fe/codecrafters-redis-go/internal/state"

func pingHandler(*state.AppState, []string) ([]byte, error) {
	return []byte("+PONG\r\n"), nil
}
