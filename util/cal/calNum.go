package cal

import (
	"math"
	"strconv"

	"github.com/Chairou/toolbox/util/conv"
)

// CutOffType 取整方式类型
type CutOffType int

const (
	// CutOffCeil 向上取整
	CutOffCeil CutOffType = 2
	// CutOffRound 四舍五入
	CutOffRound CutOffType = 3
	// CutOffFloor 向下取整
	CutOffFloor CutOffType = 4
)

// 保留旧变量名以保持向后兼容
var (
	CUT_OFF_CEIL  = int(CutOffCeil)
	CUT_OFF_ROUND = int(CutOffRound)
	CUT_OFF_FLOOR = int(CutOffFloor)
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
	switch CutOffType(cutoff) {
	case CutOffCeil:
		tmp = CeilFloat(num, precision)
	case CutOffRound:
		tmp = RoundFloat(num, precision)
	case CutOffFloor:
		tmp = FloorFloat(num, precision)
	default:
		panic("invalid cutoff")
	}
	return strconv.FormatFloat(tmp*100, 'f', places, 64)
}

// Percent 输出带百分号的百分比字符串
func Percent(num float64, places int, cutoff int) string {
	return PercentVal(num, places, cutoff) + "%"
}

// AnyPercent 任意类型转换为百分比
func AnyPercent(num interface{}, places int, cutoff int) string {
	numFloat, ok := conv.Float64(num)
	if !ok {
		panic("AnyPercent err: num can't convert to float64")
	}
	return Percent(numFloat, places, cutoff)
}
