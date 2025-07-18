package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils"
)

func keysHandler(req *request.Request, args []string) error {
	if len(args) == 0 {
		return errors.New("keys requires at least one argument")
	}

	if args[0] != "*" {
		return errors.New("only wildcard '*' is supported")
	}

	keys := req.State.GetStore().Keys()

	res := utils.StringsToRESPBulkStr(keys)

	return writeResponse(req, res)
}
