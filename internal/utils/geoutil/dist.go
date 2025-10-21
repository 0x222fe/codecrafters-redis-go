package geoutil

import "math"

const (
	earthRadius = 6372797.560856
)

func Distance(lo1, la1, lo2, la2 float64) float64 {
	lo1 = radians(lo1)
	la1 = radians(la1)
	lo2 = radians(lo2)
	la2 = radians(la2)

	dLa := la2 - la1
	dLo := lo2 - lo1

	a := haversine(dLa) + math.Cos(la1)*math.Cos(la2)*haversine(dLo)
	c := 2 * math.Asin(math.Sqrt(a))

	return earthRadius * c
}

func radians(v float64) float64 {
	return v * math.Pi / 180
}

func haversine(v float64) float64 {
	return 0.5 * (1 - math.Cos(v))
}