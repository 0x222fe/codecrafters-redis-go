package request

import (
	"context"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

type Request struct {
	Ctx         context.Context
	Client      *client.Client
	State       *state.AppState
	Transaction *Transaction
	Propagated  bool
	SubMode     bool
}

func NewRequest(ctx context.Context, client *client.Client, state *state.AppState) *Request {
	return &Request{
		Ctx:         ctx,
		Client:      client,
		State:       state,
		Transaction: nil,
		Propagated:  false,
	}
}

func (r *Request) StartTransaction() {
	r.Transaction = NewTransaction()
}

func (r *Request) ExecTransaction() ([]resp.RESPValue, bool, error) {
	defer func() {
		r.Transaction = nil
	}()

	r.Transaction.Executing = true

	if len(r.Transaction.Commands) == 0 {
		return nil, false, nil
	}

	for _, cmd := range r.Transaction.Commands {
		err := cmd.Handler.Handle(r, cmd.Command)
		if err != nil {
			r.Transaction.WriteResp(resp.NewRESPError(err))
		}
	}
	res := r.Transaction.Responses
	return res, true, nil
}

func (r *Request) DiscardTransaction() {
	r.Transaction = nil
}

func (r *Request) IsInTxn() bool {
	return r.Transaction != nil
}

func (r *Request) GetWriter() client.RespWriter {
	if r.Transaction != nil && r.Transaction.Executing {
		return r.Transaction
	}

	return r.Client
}