package command

import (
	"errors"
	"fmt"
	"io"

	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

type command string
type commandHandler func(state *state.AppState, args []string, writer io.Writer) error

const (
	PING     command = "PING"
	ECHO     command = "ECHO"
	SET      command = "SET"
	GET      command = "GET"
	CONFIG   command = "CONFIG"
	KEYS     command = "KEYS"
	INFO     command = "INFO"
	REPLCONF command = "REPLCONF"
	PSYNC    command = "PSYNC"
)

var (
	commands = map[command]commandHandler{
		PING:     pingHandler,
		ECHO:     echoHandler,
		SET:      setHandler,
		GET:      getHandler,
		CONFIG:   configHandler,
		KEYS:     keysHandler,
		INFO:     infoHandler,
		REPLCONF: replconfHandler,
		PSYNC:    psyncHandler,
	}
)

func RunCommand(state *state.AppState, cmd string, args []string, writer io.Writer) error {
	handler, exists := commands[command(cmd)]
	if !exists {
		return errors.New("unknown command: " + cmd)
	}

	return handler(state, args, writer)
}

func writeResponse(writer io.Writer, response []byte) error {
	_, err := writer.Write(response)
	if err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	return nil
}
