package precision

import "math"

func ToFix(amount float64, precision int64) float64 {
	output := math.Pow10(int(precision))
	return float64(int(amount*output)) / output
}
