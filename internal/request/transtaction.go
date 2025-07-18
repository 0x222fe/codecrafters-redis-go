package request

import (
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

type TransactionCommandHandler interface {
	Handle(req *Request, cmd Command) error
}

type TxnCommand struct {
	Command Command
	Handler TransactionCommandHandler
}

type Transaction struct {
	Commands  []TxnCommand
	Responses []resp.RESPValue
}

func (t *Transaction) WriteResp(r resp.RESPValue) error {
	t.Responses = append(t.Responses, r)
	return nil
}
