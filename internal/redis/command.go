package redis

import (
	"errors"
)

type command string
type commandHandler func(args []string) ([]byte, error)

const (
	PING command = "PING"
	ECHO command = "ECHO"
)

var (
	commands = map[command]commandHandler{
		PING: pingHandler,
		ECHO: echoHandler,
	}
)

func RunCommand(cmd string, args []string) ([]byte, error) {
	handler, exists := commands[command(cmd)]
	if !exists {
		return nil, errors.New("unknown command: " + cmd)
	}

	return handler(args)
}

func pingHandler(_ []string) ([]byte, error) {
	return []byte("+PONG\r\n"), nil
}

func echoHandler(args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("ECHO requires at least one argument")
	}

	response := "+" + args[0] + "\r\n"
	return []byte(response), nil
}
