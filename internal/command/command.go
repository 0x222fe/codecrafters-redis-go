package command

import (
	"errors"
)

type command string
type commandHandler func(args []string) ([]byte, error)

const (
	PING   command = "PING"
	ECHO   command = "ECHO"
	SET    command = "SET"
	GET    command = "GET"
	CONFIG command = "CONFIG"
)

var (
	commands = map[command]commandHandler{
		PING:   pingHandler,
		ECHO:   echoHandler,
		SET:    setHandler,
		GET:    getHandler,
		CONFIG: configHandler,
	}
)

func RunCommand(cmd string, args []string) ([]byte, error) {
	handler, exists := commands[command(cmd)]
	if !exists {
		return nil, errors.New("unknown command: " + cmd)
	}

	return handler(args)
}

