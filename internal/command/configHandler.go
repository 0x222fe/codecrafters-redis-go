package command

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func configHandler(state *state.AppState, args []string, writer io.Writer) error {
	if len(args) < 2 {
		return errors.New("CONFIG requires at least two arguments")
	}
	cfgName := strings.ToLower(args[1])

	var bytes []byte

	switch strings.ToUpper(args[0]) {
	case "GET":
		val, err := getConfig(state, cfgName)
		if err != nil {
			return err
		}

		result, err := resp.RESPEncode([]string{cfgName, val})
		if err != nil {
			return fmt.Errorf("failed to encode into RESP format: %w", err)
		}

		bytes = result
	default:
		bytes = resp.RESPNIL
	}

	return writeResponse(writer, bytes)
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
