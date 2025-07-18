package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func echoHandler(req *request.Request, args []string) error {
	if len(args) == 0 {
		return errors.New("ECHO requires at least one argument")
	}

	res := resp.NewRESPString(args[0])

	return writeResponse(req, res)
}
