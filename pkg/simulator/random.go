package simulator

import "math/rand"

func newRandomSource(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

func sampleNoise(source *rand.Rand, standardDeviation float64) float64 {
	if standardDeviation == 0 {
		return 0
	}
	return source.NormFloat64() * standardDeviation
}
