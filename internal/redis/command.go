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

func configHandler(args []string) ([]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("CONFIG requires at least two arguments")
	}
	cfgName := strings.ToLower(args[1])

	switch strings.ToUpper(args[0]) {
	case "GET":
		val, err := getConfig(cfgName)
		if err != nil {
			return nil, err
		}
		return fmt.Appendf(nil,
				"*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
				len(cfgName), cfgName, len(val), val),
			nil
	default:
		return []byte("$-1\r\n"), nil
	}
}

func getConfig(cfgName string) (string, error) {
	switch cfgName {
	case "dir":
		return cfg.dir, nil
	case "dbfilename":
		return cfg.dbfilename, nil
	default:
		return "", fmt.Errorf("unknown configuration parameter: %s", cfgName)
	}
}
