package redis

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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

	expMillis := int64(-1)
	var err error

	if len(args) > 2 {
		switch strings.ToUpper(args[2]) {
		case "PX":
			expMillis, err = strconv.ParseInt(args[3], 10, 64)
			if err != nil || expMillis < 0 {
				return nil, fmt.Errorf("invalid expiration time: %w", err)
			}
		}
	}

	setStore(args[0], args[1], expMillis)
	return []byte("+OK\r\n"), nil
}

func getHandler(args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("GET requires exactly one argument")
	}

	value, exists := getStore(args[0])
	if !exists {
		return []byte("$-1\r\n"), nil
	}

	res := fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)

	return []byte(res), nil
}
