package request

type TransactionCommandHandler interface {
	Handle(req *Request, cmd Command) error
}

type TxnCommand struct {
	Command Command
	Handler TransactionCommandHandler
}

type Transaction struct {
	Commands []TxnCommand
}
