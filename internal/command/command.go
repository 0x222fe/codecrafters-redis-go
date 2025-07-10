package command

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

type command string
type commandHandler func(state *state.AppState, args []string) ([]byte, error)

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

func RunCommand(state *state.AppState, cmd string, args []string) ([]byte, error) {
	handler, exists := commands[command(cmd)]
	if !exists {
		return nil, errors.New("unknown command: " + cmd)
	}

	return handler(state, args)
}
