package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func subscribeHandler(req *request.Request, args []string) error {
	if len(args) < 1 {
		return errors.New("SUBSCRIBE requires at least one argument")
	}

	subMsg := "subscribe"
	for _, channel := range args {
		sub := req.State.AddSubscriber(req.Client, channel)
		req.Client.WriteResp(
			resp.NewRESPArray(
				[]resp.RESPValue{
					resp.NewRESPBulkString(&subMsg),
					resp.NewRESPBulkString(&channel),
					resp.NewRESPInt(int64(len(sub.Channels))),
				},
			),
		)
	}

	req.SubMode = true

	return nil
}