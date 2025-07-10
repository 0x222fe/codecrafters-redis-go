package command

import (
	"errors"
	"io"

	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func echoHandler(state *state.AppState, args []string, writer io.Writer) error {
	if len(args) == 0 {
		return errors.New("ECHO requires at least one argument")
	}

	response := "+" + args[0] + "\r\n"

	return writeResponse(writer, []byte(response))
}
