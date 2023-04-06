package cal

import (
	"fmt"
	"math"
)

var (
	CUT_OFF_CEIL  = 2
	CUT_OFF_ROUND = 3
	CUT_OFF_FLOOR = 4
)

func CeilFloat(f float64, places int) float64 {
	if places < 0 {
		panic("places must be a non-negative integer")
	}
	scale := math.Pow10(places)
	return math.Ceil(f*scale) / scale
}

func RoundFloat(num float64, places int) float64 {
	if places < 0 {
		panic("places must be a non-negative integer")
	}
	shift := math.Pow(10, float64(places))
	rounded := math.Round(num*shift) / shift
	return rounded
}

func FloorFloat(num float64, places int) float64 {
	if places < 0 {
		panic("places must be a non-negative integer")
	}
	multiplier := math.Pow(10, float64(places))
	return math.Floor(num*multiplier) / multiplier
}

func Percent(num float64, places int, cutoff int) string {
	if places < 0 {
		panic("places must be a non-negative integer")
	}
	places += 2
	var tmp float64
	switch cutoff {
	case CUT_OFF_CEIL:
		tmp = CeilFloat(num, places)
	case CUT_OFF_ROUND:
		tmp = RoundFloat(num, places)
	case CUT_OFF_FLOOR:
		tmp = FloorFloat(num, places)
	default:
		panic("invaild cutoff")
	}
	return fmt.Sprintf("%.2f%%", tmp*100)
}
