package command

import (
	"io"

	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func pingHandler(state *state.AppState, args []string, writer io.Writer) error {
	return writeResponse(writer, []byte("+PONG\r\n"))
}
