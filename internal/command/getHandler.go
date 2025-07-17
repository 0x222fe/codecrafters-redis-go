package command

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func getHandler(req *request.Request, args []string) error {
	if len(args) != 1 {
		return errors.New("Usage: GET <key>")
	}

	var res []byte
	value, exists := req.State.GetStore().GetString(args[0])
	if !exists {
		res = resp.RESPNilArray.Encode()
	} else {
		res = resp.NewRESPString(value).Encode()
	}

	return writeResponse(req.Client, res)
}
