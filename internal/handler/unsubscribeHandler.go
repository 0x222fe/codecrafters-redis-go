package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func unsubscribeHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) == 0 {
		return errors.New("UNSUBSCRIBE requires at least one argument")
	}

	unsubMsg := "unsubscribe"
	for _, channel := range args {
		sub := s.UnsubChannel(c.Conn.ID, channel)
		chanCount := 0
		if sub != nil {
			chanCount = len(sub.Channels)
		}

		c.Conn.WriteResp(
			resp.NewArray(
				[]resp.RESPValue{
					resp.NewBulkString(&unsubMsg),
					resp.NewBulkString(&channel),
					resp.NewInt(int64(chanCount)),
				},
			),
		)

	}

	return nil
}
