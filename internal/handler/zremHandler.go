package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func zremHandler(req *request.Request, args []string) error {
	if len(args) != 2 {
		return errors.New("ZSCORE requires exactly 2 arguments")
	}

	key, member := args[0], args[1]

	ok := req.State.GetStore().RemoveSortedSetMember(key, member)

	var res resp.RESPValue
	if !ok {
		res = resp.NewInt(0)
	} else {
		res = resp.NewInt(1)
	}

	writeResponse(req, res)
	return nil
}