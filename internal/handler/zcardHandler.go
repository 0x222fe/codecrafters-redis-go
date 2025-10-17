package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func zcardHandler(req *request.Request, args []string) error {
	if len(args) != 1 {
		return errors.New("ZRANGE requires exactly 1 argument")
	}

	key := args[0]

	count := req.State.GetStore().CountSortedSetMembers(key)

	var res = resp.NewRESPInt(int64(count))

	writeResponse(req, res)

	return nil
}