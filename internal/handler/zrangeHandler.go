package handler

import (
	"errors"
	"strconv"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
)

func zrangeHandler(req *request.Request, args []string) error {
	if len(args) != 3 {
		return errors.New("ZRANGE requires exactly 3 arguments")
	}

	key := args[0]

	start, err := strconv.Atoi(args[1])
	if err != nil {
		return errors.New("ZRANGE start argument must be an integer")
	}
	end, err := strconv.Atoi(args[2])
	if err != nil {
		return errors.New("ZRANGE end argument must be an integer")
	}

	members := req.State.GetStore().ListSortedSetMembersByRank(key, start, end)

	var res = resputil.BulkStringsToRESPArray(members)

	writeResponse(req, res)
	return nil
}
