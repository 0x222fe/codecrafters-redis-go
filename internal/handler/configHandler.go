package handler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils"
)

func configHandler(req *request.Request, args []string) error {
	if len(args) < 2 {
		return errors.New("CONFIG requires at least two arguments")
	}
	cfgName := strings.ToLower(args[1])

	switch strings.ToUpper(args[0]) {
	case "GET":
		val, err := getConfig(req.State, cfgName)
		if err != nil {
			return err
		}

		encoded := utils.EncodeBulkStrArrToRESP([]string{cfgName, val})
		return writeResponse(req.Client, encoded)

	default:
		return writeResponse(req.Client, resp.RESPNilBulkString.Encode())
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
