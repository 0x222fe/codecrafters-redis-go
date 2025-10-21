package handler

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/geoutil"
)

func geoaddHandler(req *request.Request, args []string) error {
	if len(args) < 4 {
		return errors.New("GEOADD requires at least 4 arguments")
	}

	if (len(args)-1)%3 != 0 {
		return errors.New("GEOADD requires arguments in groups of longitude, latitude, and member after the key")
	}

	key := args[0]
	locationLen := (len(args) - 1) / 3

	locations := make([]store.SortedSetMember, 0, locationLen)

	for i := range locationLen {
		lo, la, m := args[i*3+1], args[i*3+2], args[i*3+3]

		longitude, err := strconv.ParseFloat(lo, 64)
		if err != nil {
			return fmt.Errorf("GEOADD: invalid longitude: %s", err)
		}

		latitude, err := strconv.ParseFloat(la, 64)
		if err != nil {
			return fmt.Errorf("GEOADD: invalid latitude: %s", err)
		}

		score := geoutil.GenerateScore(longitude, latitude)
		locations = append(locations, store.SortedSetMember{Score: score, Member: m})
	}

	count := req.State.GetStore().AddToSortedSet(key, locations)

	res := resp.NewInt(int64(count))

	writeResponse(req, res)
	return nil
}
