package geoutil

import "math"

const (
	MinLongitude = -180.0
	MaxLongitude = 180.0
	MinLatitude  = -85.05112878
	MaxLatitude  = 85.05112878
)

var (
	scale = math.Pow(2, 26)
)

// GenerateScore encodes longitude and latitude into a Redis-style geo score.
func GenerateScore(lo, la float64) float64 {
	x := normalize(lo, MinLongitude, MaxLongitude)
	y := normalize(la, MinLatitude, MaxLatitude)
	score := (interleave(x) << 1) | interleave(y)

	return float64(score)
}

// DecodeScore decodes a Redis-style geo score back to longitude and latitude.
func DecodeScore(score float64) (lo, la float64) {
	s := uint64(score)
	x := denormalize(deinterleave(s>>1), MinLongitude, MaxLongitude)
	y := denormalize(deinterleave(s), MinLatitude, MaxLatitude)
	return float64(x), float64(y)
}

func normalize(val, min, max float64) uint32 {
	return uint32((val - min) / (max - min) * scale)
}

func denormalize(val uint32, min, max float64) float64 {
	minVal := min + (max-min)*(float64(val)/scale)
	maxVal := min + (max-min)*(float64(val+1)/scale)
	return (maxVal + minVal) / 2
}

func interleave(val uint32) uint64 {
	v := uint64(val)
	v = (v | (v << 16)) & 0x0000FFFF0000FFFF
	v = (v | (v << 8)) & 0x00FF00FF00FF00FF
	v = (v | (v << 4)) & 0x0F0F0F0F0F0F0F0F
	v = (v | (v << 2)) & 0x3333333333333333
	v = (v | (v << 1)) & 0x5555555555555555
	return v
}

func deinterleave(v uint64) uint32 {
	v = v & 0x5555555555555555
	v = (v | (v >> 1)) & 0x3333333333333333
	v = (v | (v >> 2)) & 0x0F0F0F0F0F0F0F0F
	v = (v | (v >> 4)) & 0x00FF00FF00FF00FF
	v = (v | (v >> 8)) & 0x0000FFFF0000FFFF
	v = (v | (v >> 16)) & 0x00000000FFFFFFFF
	return uint32(v)
}