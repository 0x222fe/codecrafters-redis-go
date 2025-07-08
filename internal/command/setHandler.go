package command

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

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

	store.Set(args[0], args[1], expMillis)
	return []byte("+OK\r\n"), nil
}
