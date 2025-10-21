package geoutil

import (
	"math"
)

const (
	minLon = -180.0
	maxLon = 180.0
	minLat = -85.05112878
	maxLat = 85.05112878
	scale  = float64((1 << 26) - 1)
)

func normalize(value, min, max float64) uint32 {
	if value < min {
		value = min
	} else if value > max {
		value = max
	}
	return uint32(math.Round((value - min) / (max - min) * scale))
}

func interleaveBits(x, y uint32) uint64 {
	var result uint64
	for i := range 26 {
		result |= ((uint64(y) >> i) & 1) << (2 * i)
		result |= ((uint64(x) >> i) & 1) << (2*i + 1)
	}
	return result
}

// GenerateScore encodes longitude and latitude into a Redis-style geo score.
func GenerateScore(lo, la float64) float64 {
	x := normalize(lo, minLon, maxLon)
	y := normalize(la, minLat, maxLat)
	score := interleaveBits(x, y)
	return float64(score)
}

func deinterleaveBits(z uint64) (uint32, uint32) {
	var x, y uint32
	for i := range 26 {
		y |= uint32((z>>(2*i))&1) << i
		x |= uint32((z>>(2*i+1))&1) << i
	}
	return x, y
}

func denormalize(n uint32, min, max float64) float64 {
	return float64(n)/scale*(max-min) + min
}

// DecodeScore decodes a Redis-style geo score back to longitude and latitude.
func DecodeScore(score float64) (lo, la float64) {
	z := uint64(score)
	x, y := deinterleaveBits(z)
	lo = denormalize(x, minLon, maxLon)
	la = denormalize(y, minLat, maxLat)
	return
}