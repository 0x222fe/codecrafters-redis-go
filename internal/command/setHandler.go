package command

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/state"
)

func setHandler(state *state.AppState, args []string) ([]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("SET requires at least two arguments")
	}

	expMillis := new(int64)
	var err error

	if len(args) > 2 {
		switch strings.ToUpper(args[2]) {
		case "PX":
			*expMillis, err = strconv.ParseInt(args[3], 10, 64)
			if err != nil || *expMillis < 0 {
				return nil, fmt.Errorf("invalid expiration time: %w", err)
			}
		case "EX":
			expSeconds, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil || expSeconds < 0 {
				return nil, fmt.Errorf("invalid expiration time: %w", err)
			}
			*expMillis = expSeconds * 1000
		}
	}

	state.Store.Set(args[0], args[1], expMillis)
	return []byte("+OK\r\n"), nil
}
