package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func typeHandler(req *request.Request, args []string) error {
	if len(args) != 1 {
		return errors.New("TYPE requires exactly one argument")
	}

	key := args[0]
	valType := req.State.GetStore().Type(key)
	return writeResponse(req.Client, resp.NewRESPString(valType).Encode())
}
