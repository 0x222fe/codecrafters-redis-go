package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func unsubscribeHandler(req *request.Request, args []string) error {
	if len(args) == 0 {
		return errors.New("UNSUBSCRIBE requires at least one argument")
	}

	unsubMsg := "unsubscribe"
	for _, channel := range args {
		sub := req.State.UnsubChannel(req.Client.ID, channel)
		chanCount := 0
		if sub != nil {
			chanCount = len(sub.Channels)
		}

		req.Client.WriteResp(
			resp.NewRESPArray(
				[]resp.RESPValue{
					resp.NewRESPBulkString(&unsubMsg),
					resp.NewRESPBulkString(&channel),
					resp.NewRESPInt(int64(chanCount)),
				},
			),
		)

	}

	return nil
}