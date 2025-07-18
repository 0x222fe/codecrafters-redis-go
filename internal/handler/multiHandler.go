package handler

import (
	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func multiHandler(req *request.Request, args []string) error {
	req.InTxn = true
	encoded := resp.NewRESPString("OK").Encode()
	writeResponse(req.Client, encoded)
	return nil
}
