package command

import (
	"errors"
	"io"

	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func psyncHandler(state *state.AppState, args []string, writer io.Writer) error {
	if len(args) < 2 {
		return errors.New("PSYNC requires at least 2 arguments")
	}

	if args[0] != "?" || args[1] != "-1" {
		return errors.New("PSYNC only supports ? -1 for now")
	}

	message := "+FULLRESYNC " + state.ReplicantionID + " " + "0" + "\r\n"

	return writeResponse(writer, []byte(message))
}
