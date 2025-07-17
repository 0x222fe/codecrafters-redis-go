package command

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils"
)

func xreadHandler(req *request.Request, args []string) error {
	if len(args) == 0 {
		return errors.New("XREAD requires at least one argument")
	}
	key, idStr := args[1], args[2]

	id, err := store.ParseStreamEntryID(idStr)
	if err != nil {
		return err
	}
	id.Seq++

	stream, ok := req.State.GetStore().GetStream(key)
	arr := make([]resp.RESPValue, 0)
	if ok {
		entries := stream.Range(id.RadixKey(), nil)
		streamArr := make([]resp.RESPValue, 0, 2*len(entries))

		streamArr = append(streamArr, resp.NewRESPBulkString(&key))
		streamEntryRESP := utils.StreamEntriesToRESPArray(entries)
		streamArr = append(streamArr, streamEntryRESP)
		arr = append(arr, resp.NewRESPArray(streamArr))
	}
	encoded := resp.NewRESPArray(arr).Encode()
	writeResponse(req.Client, encoded)
	return nil
}
