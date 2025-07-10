package command

import (
	"io"

	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func replconfHandler(state *state.AppState, args []string, writer io.Writer) error {
	//INFO: ignore args for now

	return writeResponse(writer, []byte("+OK\r\n"))
}
