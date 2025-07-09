package command

import (
	"errors"

	"github.com/codecrafters-io/redis-starter-go/internal/state"
)

type command string
type commandHandler func(state *state.AppState, args []string) ([]byte, error)

const (
	PING   command = "PING"
	ECHO   command = "ECHO"
	SET    command = "SET"
	GET    command = "GET"
	CONFIG command = "CONFIG"
	KEYS   command = "KEYS"
	INFO   command = "INFO"
)

var (
	commands = map[command]commandHandler{
		PING:   pingHandler,
		ECHO:   echoHandler,
		SET:    setHandler,
		GET:    getHandler,
		CONFIG: configHandler,
		KEYS:   keysHandler,
		INFO:   infoHandler,
	}
)

func RunCommand(state *state.AppState, cmd string, args []string) ([]byte, error) {
	handler, exists := commands[command(cmd)]
	if !exists {
		return nil, errors.New("unknown command: " + cmd)
	}

	return handler(state, args)
}
