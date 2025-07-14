package command

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

func setHandler(req *request.Request, args []string) error {
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

	req.State.GetStore().Set(args[0], args[1], store.String, expireAt)

	if req.Propagated {
		return nil
	}
	return writeResponse(req.Client, resp.NewRESPString("OK").Encode())
}
