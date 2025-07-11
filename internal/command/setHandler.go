package command

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func setHandler(appState *state.AppState, args []string) ([]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("SET requires at least two arguments")
	}

	var expMillis int64
	var err error

	if len(args) > 2 {
		switch strings.ToUpper(args[2]) {
		case "PX":
			expMillis, err = strconv.ParseInt(args[3], 10, 64)
			if err != nil || expMillis < 0 {
				return nil, fmt.Errorf("invalid expiration time: %w", err)
			}
		case "EX":
			expSeconds, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil || expSeconds < 0 {
				return nil, fmt.Errorf("invalid expiration time: %w", err)
			}
			expMillis = expSeconds * 1000
		}
	}

	var expireAt *int64
	if expMillis > 0 {
		expireAt = new(int64)
		*expireAt = time.Now().Add(time.Duration(expMillis) * time.Millisecond).UnixMilli()
	}

	appState.GetStore().Set(args[0], args[1], expireAt)

	return []byte("+OK\r\n"), nil
}
