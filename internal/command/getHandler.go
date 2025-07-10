package command

import (
	"errors"
	"fmt"
	"io"

	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func getHandler(state *state.AppState, args []string, writer io.Writer) error {
	if len(args) != 1 {
		return errors.New("Usage: GET <key>")
	}

	value, exists := state.GetStore().Get(args[0])
	if !exists {
		return writeResponse(writer, resp.RESPNIL)
	}

	result, err := resp.RESPEncode(value)
	if err != nil {
		return fmt.Errorf("failed to encode value into RESP format: %w", err)
	}

	return writeResponse(writer, result)
}
