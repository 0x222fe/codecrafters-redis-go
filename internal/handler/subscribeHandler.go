package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func subscribeHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) < 1 {
		return errors.New("SUBSCRIBE requires at least one argument")
	}

	subMsg := "subscribe"
	for _, channel := range args {
		sub := s.AddSubscriber(c.Conn, channel)
		c.Conn.WriteResp(
			resp.NewArray(
				[]resp.RESPValue{
					resp.NewBulkString(&subMsg),
					resp.NewBulkString(&channel),
					resp.NewInt(int64(len(sub.Channels))),
				},
			),
		)
	}

	c.SubMode = true

	return nil
}
