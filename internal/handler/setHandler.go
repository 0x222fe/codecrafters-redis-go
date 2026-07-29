package handler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

func setHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) < 2 {
		return errors.New("SET requires at least two arguments")
	}

	var expMillis int64
	var err error

	if len(args) > 2 {
		switch strings.ToUpper(args[2]) {
		case "PX":
			expMillis, err = strconv.ParseInt(args[3], 10, 64)
			if err != nil || expMillis < 0 {
				return fmt.Errorf("invalid expiration time: %w", err)
			}
		case "EX":
			expSeconds, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil || expSeconds < 0 {
				return fmt.Errorf("invalid expiration time: %w", err)
			}
			expMillis = expSeconds * 1000
		}
	}

	var expireAt *int64
	if expMillis > 0 {
		expireAt = new(int64)
		*expireAt = time.Now().Add(time.Duration(expMillis) * time.Millisecond).UnixMilli()
	}

	s.GetStore().Set(args[0], args[1], store.String, expireAt)

	if c.Propagated {
		return nil
	}
	return writeResponse(c, resp.NewString("OK"))
}
