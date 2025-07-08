package command

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/config"
)

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
		return config.Cfg.Dir, nil
	case "dbfilename":
		return config.Cfg.Dbfilename, nil
	default:
		return "", fmt.Errorf("unknown configuration parameter: %s", cfgName)
	}
}
