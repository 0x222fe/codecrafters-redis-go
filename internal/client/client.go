package client

import (
	"context"

	"github.com/0x222fe/codecrafters-redis-go/internal/connection"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

type Client struct {
	Ctx         context.Context
	Conn        *connection.Connection
	Transaction *Transaction
	Propagated  bool
	SubMode     bool
}

func NewClient(ctx context.Context, conn *connection.Connection) *Client {
	return &Client{
		Ctx:         ctx,
		Conn:        conn,
		Transaction: nil,
		Propagated:  false,
	}
}

func (c *Client) StartTransaction() {
	c.Transaction = NewTransaction()
}

func (c *Client) ExecTransaction(s *state.AppState) ([]resp.RESPValue, bool, error) {
	defer func() {
		c.Transaction = nil
	}()

	c.Transaction.Executing = true

	if len(c.Transaction.Commands) == 0 {
		return []resp.RESPValue{}, false, nil
	}

	if !s.GetStore().WatchesValid(c.Conn.ID) {
		return nil, false, nil
	}

	for _, cmd := range c.Transaction.Commands {
		err := cmd.Handler.Handle(c, s, cmd.Command)
		if err != nil {
			c.Transaction.WriteResp(resp.NewError(err))
		}
	}
	res := c.Transaction.Responses
	return res, true, nil
}

func (c *Client) DiscardTransaction() {
	c.Transaction = nil
}

func (c *Client) IsInTxn() bool {
	return c.Transaction != nil
}

func (c *Client) GetWriter() connection.RespWriter {
	if c.Transaction != nil && c.Transaction.Executing {
		return c.Transaction
	}

	return c.Conn
}
