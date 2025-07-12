package utils

import "github.com/0x222fe/codecrafters-redis-go/internal/resp"

func EncodeStringSliceToRESP(slice []string) []byte {
	arr := make([]resp.RESPValue, len(slice))

	for i, v := range slice {
		arr[i] = resp.NewRESPBulkString(&v)
	}

	return resp.NewRESPArray(arr).Encode()
}
