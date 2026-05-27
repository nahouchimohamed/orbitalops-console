package domain

import "math"

func ClampLatDeg(lat float64) float64 {
	if lat > 90 {
		return 90
	}
	if lat < -90 {
		return -90
	}
	return lat
}

func WrapLonDeg(lon float64) float64 {
	x := math.Mod(lon+180, 360)
	if x < 0 {
		x += 360
	}
	return x - 180
}
