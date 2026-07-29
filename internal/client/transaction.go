package client

import (
	"github.com/0x222fe/codecrafters-redis-go/internal/command"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

type TransactionCommandHandler interface {
	Handle(c *Client, s *state.AppState, cmd command.Command) error
}

type TxnCommand struct {
	Command command.Command
	Handler TransactionCommandHandler
}

type Transaction struct {
	Executing bool
	Commands  []TxnCommand
	Responses []resp.RESPValue
}

func NewTransaction() *Transaction {
	return &Transaction{
		Executing: false,
		Commands:  make([]TxnCommand, 0),
		Responses: make([]resp.RESPValue, 0),
	}
}

func (t *Transaction) WriteResp(r resp.RESPValue) error {
	t.Responses = append(t.Responses, r)
	return nil
}
