package handler

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

func zaddHandler(req *request.Request, args []string) error {
	if len(args) < 3 {
		return errors.New("ZADD requires at least 3 arguments")
	}

	if len(args)%2 != 1 {
		return errors.New("ZADD requires an odd number of arguments")
	}
	key := args[0]

	members := make([]store.SortedSetMember, 0, len(args)/2)

	for i := 1; i < len(args); i += 2 {
		scoreStr, member := args[i], args[i+1]
		score, err := strconv.ParseFloat(scoreStr, 64)
		if err != nil {
			return fmt.Errorf("ZADD %s %s %s: error parsing score: %s", key, scoreStr, member, err)
		}

		members = append(members, store.SortedSetMember{
			Score:  score,
			Member: member,
		})
	}

	count := req.State.GetStore().AddToSortedSet(key, members)

	res := resp.NewInt(int64(count))

	writeResponse(req, res)
	return nil
}
