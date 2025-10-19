package utils

import (
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

func BulkStringsToRESPArray(slice []string) resp.RESPValue {
	arr := make([]resp.RESPValue, len(slice))

	for i, v := range slice {
		arr[i] = resp.NewBulkString(&v)
	}

	return resp.NewArray(arr)
}

func StreamEntriesToRESPArray(entries []*store.StreamEntry) resp.RESPValue {
	entryArr := make([]resp.RESPValue, 0)

	for _, entry := range entries {
		idStr := entry.ID.String()
		fieldArr := make([]resp.RESPValue, 0, 2*len(entry.Fields))
		for k, v := range entry.Fields {
			fieldArr = append(fieldArr, resp.NewBulkString(&k))
			fieldArr = append(fieldArr, resp.NewBulkString(&v))
		}

		inner := []resp.RESPValue{
			resp.NewBulkString(&idStr),
			resp.NewArray(fieldArr),
		}
		entryArr = append(entryArr, resp.NewArray(inner))
	}

	return resp.NewArray(entryArr)
}
