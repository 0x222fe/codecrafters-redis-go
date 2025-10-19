package handler

import (
	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils"
)

func pingHandler(req *request.Request, args []string) error {
	if req.Propagated {
		return nil
	}
	if req.SubMode {
		return writeResponse(req, utils.BulkStringsToRESPArray([]string{"pong", ""}))
	}
	return writeResponse(req, resp.NewString("PONG"))
}
