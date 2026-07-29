package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

func publishHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) != 2 {
		return errors.New("PUBLISH requires exactly 2 arguments")
	}

	channel, message := args[0], args[1]

	sent := s.Publish(channel, []byte(message))

	writeResponse(c, resp.NewInt(int64(sent)))

	return nil
}
