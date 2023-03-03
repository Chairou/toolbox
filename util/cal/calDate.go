package cal

import (
	"math"
	"time"
)

// GetDiffDays 获取两个时间相差的天数
func GetDiffDays(t1, t2 time.Time) int {
	t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, time.Local)
	t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, time.Local)
	return int(math.Abs(t1.Sub(t2).Hours() / 24))
}

// GetFirstAndLastDateOfWeek 获取当天所在周的周一和周日时间
func GetFirstAndLastDateOfWeek(date time.Time) (weekMonday, weekSunday string) {
	now := date

	sOffset := int(time.Monday - now.Weekday())
	if sOffset > 0 {
		sOffset = -6
	}
	weekStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, sOffset)
	weekMonday = weekStartDate.Format("2006-01-02")

	eOffset := int(time.Saturday - now.Weekday())
	if eOffset > 5 {
		eOffset = -1
	}
	weekEndDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, eOffset+1)
	weekSunday = weekEndDate.Format("2006-01-02")
	return
}
