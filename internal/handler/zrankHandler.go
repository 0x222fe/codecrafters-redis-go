package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func zrankHandler(req *request.Request, args []string) error {
	if len(args) != 2 {
		return errors.New("ZRANK requires exactly 2 arguments")
	}

	key, member := args[0], args[1]

	rank, ok := req.State.GetStore().QuerySortedSetRank(key, member)

	var res resp.RESPValue
	if !ok {
		res = resp.RESPNilBulkString
	} else {
		res = resp.NewRESPInt(int64(rank))
	}

	writeResponse(req, res)
	return nil
}