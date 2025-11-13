package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

func getHandler(req *request.Request, args []string) error {
	if len(args) != 1 {
		return errors.New("Usage: GET <key>")
	}

	var res resp.RESPValue
	value, ok := req.State.GetStore().GetExact(args[0], store.String)
	str, parseOk := value.(string)
	if !ok || !parseOk {
		res = resp.RESPNilBulkString
	} else {
		res = resp.NewBulkString(&str)
	}

	return writeResponse(req, res)
}
