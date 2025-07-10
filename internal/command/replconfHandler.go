package command

import "github.com/codecrafters-io/redis-starter-go/internal/state"

func replconfHandler(state *state.AppState, args []string) ([]byte, error) {
	//INFO: ignore args for now

	return []byte("+OK\r\n"), nil
}
