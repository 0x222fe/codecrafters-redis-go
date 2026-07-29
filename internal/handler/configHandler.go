package handler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
)

func configHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) < 2 {
		return errors.New("CONFIG requires at least two arguments")
	}
	cfgName := strings.ToLower(args[1])

	switch strings.ToUpper(args[0]) {
	case "GET":
		val, err := getConfig(s, cfgName)
		if err != nil {
			return err
		}

		res := resputil.BulkStringsToRESPArray([]string{cfgName, val})
		return writeResponse(c, res)

	default:
		return writeResponse(c, resp.RESPNilBulkString)
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
