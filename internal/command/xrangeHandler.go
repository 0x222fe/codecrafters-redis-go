package command

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

func xrangeHandler(req *request.Request, args []string) error {
	if len(args) != 3 {
		return errors.New("XRANGE requires 3 arguments")
	}
	key, startS, endS := args[0], args[1], args[2]

	start, err := store.ParseStreamEntryID(startS)
	if err != nil {
		return errors.New("invalid start")
	}
	end, err := store.ParseStreamEntryID(endS)
	if err != nil {
		return errors.New("invalid end")
	}

	arr := make([]resp.RESPValue, 0)
	stream, ok := req.State.GetStore().GetStream(key)
	if ok {
		entries := stream.Range(start.RadixKey(), end.RadixKey())

		for _, entry := range entries {
			idStr := entry.ID.String()
			entryArr := make([]resp.RESPValue, 0, 2*len(entry.Fields))
			for k, v := range entry.Fields {
				entryArr = append(entryArr, resp.NewRESPBulkString(&k))
				entryArr = append(entryArr, resp.NewRESPBulkString(&v))
			}

			inner := []resp.RESPValue{
				resp.NewRESPBulkString(&idStr),
				resp.NewRESPArray(entryArr),
			}
			arr = append(arr, resp.NewRESPArray(inner))
		}
	}

	encoded := resp.NewRESPArray(arr).Encode()
	writeResponse(req.Client, encoded)
	return nil
}
