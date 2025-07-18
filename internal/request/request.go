package request

import (
	"context"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

type Request struct {
	Ctx         context.Context
	Client      *client.Client
	State       *state.AppState
	InTxn       bool
	Transaction *Transaction
	TxnCommands []Command
	// Wether this request is propagated from master
	Propagated bool
}

func NewRequest(ctx context.Context, client *client.Client, state *state.AppState) *Request {
	return &Request{
		Ctx:         ctx,
		Client:      client,
		State:       state,
		Transaction: &Transaction{Commands: make([]TxnCommand, 0)},
		TxnCommands: make([]Command, 0),
		Propagated:  false,
	}
}

func (r *Request) GetWriter() client.RespWriter {
	if r.InTxn {
		return r.Transaction
	}

	return r.Client
}
