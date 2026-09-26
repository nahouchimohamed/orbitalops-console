package domain

func ClampPriority(p int) int {
	if p <= 0 {
		return 3
	}
	if p > 5 {
		return 5
	}
	return p
}

func PriorityBand(p int) string {
	switch ClampPriority(p) {
	case 1:
		return "critical"
	case 2:
		return "high"
	case 3:
		return "normal"
	default:
		return "background"
	}
}
