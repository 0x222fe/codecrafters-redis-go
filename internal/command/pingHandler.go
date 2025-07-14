package command

import (
	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func pingHandler(req *request.Request, args []string) error {
	if req.Propagated {
		return nil
	}
	return writeResponse(req.Client, resp.NewRESPString("PONG").Encode())
}
