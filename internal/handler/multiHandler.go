package handler

import (
	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func multiHandler(req *request.Request, args []string) error {
	req.StartTransaction()
	res := resp.NewString("OK")
	writeResponse(req, res)
	return nil
}
