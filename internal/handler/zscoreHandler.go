package handler

import (
	"errors"
	"fmt"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func zscoreHandler(req *request.Request, args []string) error {
	if len(args) != 2 {
		return errors.New("ZSCORE requires exactly 2 arguments")
	}

	key, member := args[0], args[1]

	score, ok := req.State.GetStore().QuerySortedSetScore(key, member)

	var res resp.RESPValue
	if !ok {
		res = resp.RESPNilBulkString
	} else {
		s := fmt.Sprintf("%.17g", score)
		res = resp.NewBulkString(&s)
	}

	writeResponse(req, res)
	return nil
}
