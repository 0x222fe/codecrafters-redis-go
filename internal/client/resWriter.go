package client

import "github.com/0x222fe/codecrafters-redis-go/internal/resp"

type RespWriter interface {
	WriteResp(r resp.RESPValue) error
}
