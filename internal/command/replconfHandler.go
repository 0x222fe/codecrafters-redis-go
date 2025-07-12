package command

import (
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func replconfHandler(state *state.AppState, args []string) ([]byte, error) {
	//INFO: ignore args for now

	return resp.NewRESPString("OK").Encode(), nil
}
