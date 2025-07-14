package request

import (
	"context"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

type Request struct {
	Ctx    context.Context
	Client *client.Client
	State  *state.AppState
	// Wether this request is propagated from master
	Propagated bool
}

func NewRequest(ctx context.Context, client *client.Client, state *state.AppState) *Request {
	return &Request{
		Ctx:        ctx,
		Client:     client,
		State:      state,
		Propagated: false,
	}
}
