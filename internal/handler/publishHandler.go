package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func publishHandler(req *request.Request, args []string) error {
	if len(args) != 2 {
		return errors.New("PUBLISH requires exactly 2 arguments")
	}

	channel, message := args[0], args[1]

	sent := req.State.Publish(channel, []byte(message))

	writeResponse(req, resp.NewRESPInt(int64(sent)))

	return nil
}