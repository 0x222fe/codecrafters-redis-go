package command

import (
	"fmt"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

func xaddHandler(req *request.Request, args []string) error {
	if len(args) < 4 || len(args)%2 != 0 {
		return fmt.Errorf("XADD requires at least 4 arguments and an even number of additional arguments")
	}

	key, idStr := args[0], args[1]
	if len(key) == 0 || len(idStr) == 0 {
		return fmt.Errorf("XADD requires a key and an ID")
	}

	stream, ok := req.State.GetStore().GetStream(key)

	if !ok {
		stream = store.NewStream(key)
		req.State.GetStore().Set(key, stream, store.Stream, nil)
		err := req.State.GetStore().AddStream(key, stream)
		if err != nil {
			return err
		}
	}

	fields := make(map[string]string)
	for i := 2; i < len(args); i += 2 {
		fields[args[i]] = args[i+1]
	}
	id, err := stream.AddEntry(idStr, fields)
	if err != nil {
		return err
	}

	s := id.String()
	encoded := resp.NewRESPBulkString(&s).Encode()
	writeResponse(req.Client, encoded)
	return nil
}
