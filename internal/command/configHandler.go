package command

import (
	"errors"
	"fmt"
	"strings"

	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils"
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

		encoded := utils.EncodeStringSliceToRESP([]string{cfgName, val})
		return encoded, nil

	default:
		return resp.RESPNilBulkString.Encode(), nil
	}
}

func getConfig(appState *state.AppState, cfgName string) (string, error) {
	cfg := appState.ReadCfg()
	switch cfgName {
	case "dir":
		return cfg.Dir, nil
	case "dbfilename":
		return cfg.Dbfilename, nil
	default:
		return "", fmt.Errorf("unknown configuration parameter: %s", cfgName)
	}
}
