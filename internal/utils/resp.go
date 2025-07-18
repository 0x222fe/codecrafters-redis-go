package utils

import (
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

func StringsToRESPBulkStr(slice []string) resp.RESPValue {
	arr := make([]resp.RESPValue, len(slice))

	for i, v := range slice {
		arr[i] = resp.NewRESPBulkString(&v)
	}

	return resp.NewRESPArray(arr)
}

func StreamEntriesToRESPArray(entries []*store.StreamEntry) resp.RESPValue {
	entryArr := make([]resp.RESPValue, 0)

	for _, entry := range entries {
		idStr := entry.ID.String()
		fieldArr := make([]resp.RESPValue, 0, 2*len(entry.Fields))
		for k, v := range entry.Fields {
			fieldArr = append(fieldArr, resp.NewRESPBulkString(&k))
			fieldArr = append(fieldArr, resp.NewRESPBulkString(&v))
		}

		inner := []resp.RESPValue{
			resp.NewRESPBulkString(&idStr),
			resp.NewRESPArray(fieldArr),
		}
		entryArr = append(entryArr, resp.NewRESPArray(inner))
	}

	return resp.NewRESPArray(entryArr)
}
