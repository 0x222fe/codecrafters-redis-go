package command

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/state"
)

func configHandler(state *state.AppState, args []string) ([]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("CONFIG requires at least two arguments")
	}
	cfgName := strings.ToLower(args[1])

	switch strings.ToUpper(args[0]) {
	case "GET":
		val, err := getConfig(state, cfgName)
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

func getConfig(state *state.AppState, cfgName string) (string, error) {
	switch cfgName {
	case "dir":
		return state.Cfg.Dir, nil
	case "dbfilename":
		return state.Cfg.Dbfilename, nil
	default:
		return "", fmt.Errorf("unknown configuration parameter: %s", cfgName)
	}
}
