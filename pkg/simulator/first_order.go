package simulator

func firstOrderStep(processValue, target, seconds, tauSeconds float64) float64 {
	return processValue + seconds/tauSeconds*(target-processValue)
}
