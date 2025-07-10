package command

import (
	"errors"
	"fmt"
	"io"

	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func keysHandler(state *state.AppState, args []string, writer io.Writer) error {
	if len(args) == 0 {
		return errors.New("keys requires at least one argument")
	}

	if args[0] != "*" {
		return errors.New("only wildcard '*' is supported")
	}

	keys := state.GetStore().Keys()

	result, err := resp.RESPEncode(keys)
	if err != nil {
		return fmt.Errorf("failed to encode keys into RESP format: %w", err)
	}

	return writeResponse(writer, result)
}
