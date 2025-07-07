package redis

import (
	"errors"
)

type command string
type commandHandler func(args []string) ([]byte, error)

const (
	PING command = "PING"
	ECHO command = "ECHO"
	SET  command = "SET"
	GET  command = "GET"
)

var (
	commands = map[command]commandHandler{
		PING: pingHandler,
		ECHO: echoHandler,
		SET:  setHandler,
		GET:  getHandler,
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

func setHandler(args []string) ([]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("SET requires at least two arguments")
	}

	store[args[0]] = args[1]
	return []byte("+OK\r\n"), nil
}

func getHandler(args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("GET requires exactly one argument")
	}

	value, exists := store[args[0]]
	if !exists {
		return []byte("$-1\r\n"), nil
	}

	return []byte(value), nil
}
