package command

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils"
)

func xreadHandler(req *request.Request, args []string) error {
	if len(args) < 3 {
		return errors.New("XREAD requires at least 3 arguments")
	}

	count := len(args) - 1 // XREAD streams
	if count%2 != 0 {
		return errors.New("XREAD must have even numer of arguments after 'streams'")
	}
	count /= 2
	keys, idStrs := args[1:1+count], args[1+count:]
	ids := make([]store.StreamEntryID, 0, count)
	for i := 0; i < count; i++ {
		id, err := store.ParseStreamEntryID(idStrs[i])
		if err != nil {
			return err
		}
		id.Seq++
		ids = append(ids, id)
	}

	arr := make([]resp.RESPValue, 0)
	for i, key := range keys {
		stream, ok := req.State.GetStore().GetStream(key)
		if ok {
			entries := stream.Range(ids[i].RadixKey(), nil)
			streamArr := make([]resp.RESPValue, 0, 2*len(entries))

			streamArr = append(streamArr, resp.NewRESPBulkString(&key))
			streamEntryRESP := utils.StreamEntriesToRESPArray(entries)
			streamArr = append(streamArr, streamEntryRESP)
			arr = append(arr, resp.NewRESPArray(streamArr))
		}
	}
	encoded := resp.NewRESPArray(arr).Encode()
	writeResponse(req.Client, encoded)
	return nil
}
