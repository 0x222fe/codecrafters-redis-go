package command

import (
	"errors"
	"fmt"
	"strings"

	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
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

		result, err := resp.RESPEncode([]string{cfgName, val})
		if err != nil {
			return nil, fmt.Errorf("failed to encode into RESP format: %w", err)
		}

		return result, nil
	default:
		return resp.RESPNIL, nil
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
