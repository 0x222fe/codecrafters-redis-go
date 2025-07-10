package command

import (
	"errors"

	"github.com/codecrafters-io/redis-starter-go/internal/state"
)

func psyncHandler(state *state.AppState, args []string) ([]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("PSYNC requires at least 2 arguments")
	}

	if args[0] != "?" || args[1] != "-1" {
		return nil, errors.New("PSYNC only supports ? -1 for now")
	}

	message := "+FULLRESYNC " + state.ReplicantionID + " " + "0" + "\r\n"

	return []byte(message), nil
}
