package handler

import (
	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func multiHandler(req *request.Request, args []string) error {
	res := resp.NewRESPString("OK")
	writeResponse(req, res)
	req.InTxn = true
	return nil
}
