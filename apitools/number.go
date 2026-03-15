package apitools

import "math"

func RoundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func FloatEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}
