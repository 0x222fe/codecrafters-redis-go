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

func NeighborScoreRange(lo, la, radius float64) (float64, float64) {
	degPerMeter := 180.0 / (math.Pi * earthRadius)
	dLat := radius * degPerMeter
	dLon := radius * degPerMeter / math.Cos(radians(la))

	directions := [][2]float64{
		{0, dLat},      // N
		{dLon, dLat},   // NE
		{dLon, 0},      // E
		{dLon, -dLat},  // SE
		{0, -dLat},     // S
		{-dLon, -dLat}, // SW
		{-dLon, 0},     // W
		{-dLon, dLat},  // NW
	}

	scores := make([]float64, 1, 9)
	scores[0] = EncodeScore(lo, la)

	for _, dir := range directions {
		nLo := lo + dir[0]
		nLa := la + dir[1]

		if nLo < MinLongitude {
			nLo = MinLongitude
		}
		if nLo > MaxLongitude {
			nLo = MaxLongitude
		}
		if nLa < MinLatitude {
			nLa = MinLatitude
		}
		if nLa > MaxLatitude {
			nLa = MaxLatitude
		}

		scores = append(scores, EncodeScore(nLo, nLa))
	}

	minScore := scores[0]
	maxScore := scores[0]
	for _, s := range scores[1:] {
		if s < minScore {
			minScore = s
		}
		if s > maxScore {
			maxScore = s
		}
	}

	return minScore, maxScore
}

func radians(v float64) float64 {
	return v * math.Pi / 180
}

func haversine(v float64) float64 {
	return 0.5 * (1 - math.Cos(v))
}
