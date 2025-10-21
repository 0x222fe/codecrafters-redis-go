package handler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/geoutil"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
)

func geosearchHandler(req *request.Request, args []string) error {
	if len(args) < 6 {
		return errors.New("GEOSEARCH requires at least 6 arguments")
	}

	key := args[0]

	mode := strings.ToUpper(args[1])

	switch mode {
	case "FROMLONLAT":
		return fromLonLat(req, key, args)
	default:
		return fmt.Errorf("GEOSEARCH mode %s not supported", mode)
	}
}

func fromLonLat(req *request.Request, key string, args []string) error {
	if len(args) < 7 {
		return errors.New("GEOSEARCH FROMLONLAT requires at least 7 arguments")
	}

	shape := strings.ToUpper(args[4])
	switch shape {
	case "BYRADIUS":
	default:
		return fmt.Errorf("GEOSEARCH FROMLONLAT Shape %s not supported", shape)
	}

	lo, la, rad, unit := args[2], args[3], args[5], strings.ToLower(args[6])

	longitude, err := strconv.ParseFloat(lo, 64)
	if err != nil {
		return fmt.Errorf("GEOSEARCH FROMLONLAT longitude is not a number: %s", lo)
	}
	latitude, err := strconv.ParseFloat(la, 64)
	if err != nil {
		return fmt.Errorf("GEOSEARCH FROMLONLAT latitude is not a number: %s", la)
	}
	radius, err := strconv.ParseFloat(rad, 64)
	if err != nil {
		return fmt.Errorf("GEOSEARCH FROMLONLAT radius is not a number: %s", rad)
	}

	switch unit {
	case "m":
	case "km":
		radius = radius * 1000
	case "mi":
		radius = radius * 1609.344
	default:
		return fmt.Errorf("GEOSEARCH FROMLONLAT unit %s not supported", unit)
	}

	minScore, maxScore := geoutil.NeighborScoreRange(longitude, latitude, radius)

	locations := req.State.GetStore().QuerySortedSetMemberByScore(key, minScore, maxScore)

	result := make([]string, 0, len(locations))
	for _, location := range locations {
		lo, la := geoutil.DecodeScore(location.Score)
		dist := geoutil.Distance(longitude, latitude, lo, la)
		if dist <= radius {
			result = append(result, location.Member)
		}
	}

	res := resputil.BulkStringsToRESPArray(result)
	writeResponse(req, res)

	return nil
}
