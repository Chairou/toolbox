package cal

import (
	"github.com/Chairou/toolbox/util/conv"
	"math"
	"strconv"
)

var (
	CUT_OFF_CEIL  = 2
	CUT_OFF_ROUND = 3
	CUT_OFF_FLOOR = 4
)

// CeilFloat 向上取整，保留多少位小数
func CeilFloat(f float64, places int) float64 {
	if places < 0 {
		panic("places must be a non-negative integer")
	}
	scale := math.Pow10(places)
	return math.Ceil(f*scale) / scale
}

// RoundFloat 四舍五入， 保留places位小数
func RoundFloat(num float64, places int) float64 {
	if places < 0 {
		panic("places must be a non-negative integer")
	}
	shift := math.Pow(10, float64(places))
	rounded := math.Round(num*shift) / shift
	return rounded
}

// FloorFloat 向下取整，保留places位小数
func FloorFloat(num float64, places int) float64 {
	if places < 0 {
		panic("places must be a non-negative integer")
	}
	multiplier := math.Pow(10, float64(places))
	return math.Floor(num*multiplier) / multiplier
}

// PercentVal 输出百分比。 保留places位小数， 采用cutoff的取整方法
func PercentVal(num float64, places int, cutoff int) string {
	if places < 0 {
		panic("places must be a non-negative integer")
	}
	precision := places + 2
	var tmp float64
	switch cutoff {
	case CUT_OFF_CEIL:
		tmp = CeilFloat(num, precision)
	case CUT_OFF_ROUND:
		tmp = RoundFloat(num, precision)
	case CUT_OFF_FLOOR:
		tmp = FloorFloat(num, precision)
	default:
		panic("invaild cutoff")
	}
	//return fmt.Sprintf("%.2f%%", tmp*100)
	return strconv.FormatFloat(tmp*100, 'f', places, 64)
}

func Percent(num float64, places int, cutoff int) string {
	return PercentVal(num, places, cutoff) + "%"
}

// AnyPercent 任意类型转换为百分比，不过实话意义不是特别大，就是方便一点
func AnyPercent(num interface{}, places int, cutoff int) string {
	numFloat, ok := conv.Float64(num)
	if ok == false {
		panic("AnyPercent err: num cann't turn into float64")
	}
	return Percent(numFloat, places, cutoff)
}
