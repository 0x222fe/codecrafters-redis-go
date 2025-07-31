package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func subscribeHandler(req *request.Request, args []string) error {
	if len(args) < 1 {
		return errors.New("SUBSCRIBE requires at least one argument")
	}

	subMsg := "subscribe"
	var sub *state.Subscriber
	for _, channel := range args {
		//INFO: will get the same sub for all registrations
		sub = req.State.AddSubscriber(req.Client, channel)
		msgArr := []resp.RESPValue{
			resp.NewRESPBulkString(&subMsg),
			resp.NewRESPBulkString(&channel),
			resp.NewRESPInt(int64(len(sub.Channels))),
		}
		req.Client.WriteResp(resp.NewRESPArray(msgArr))
	}

	return nil
}