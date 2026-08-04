package domain

func PassQualityTier(maxElDeg float64) string {
	switch {
	case maxElDeg >= 40:
		return "excellent"
	case maxElDeg >= 20:
		return "good"
	case maxElDeg >= 10:
		return "fair"
	default:
		return "marginal"
	}
}
