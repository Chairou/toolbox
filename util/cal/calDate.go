// Package cal 提供日期、IP、数值等常用计算工具函数
package cal

import (
	"errors"
	"math"
	"reflect"
	"time"
)

const (
	TimeDate      string = "2006-01-02"
	TimeCommon    string = "2006-01-02 15:04:05"
	TimeShortDate string = "20060102"

	TimeToDays    int = 1
	TimeToHours   int = 2
	TimeToMinutes int = 3
	TimeToSeconds int = 4
)

// 保留旧常量名以保持向后兼容
const (
	TIME_DATE      = TimeDate
	TIME_COMMON    = TimeCommon
	TIME_SHORTDATE = TimeShortDate

	TIME_TO_DAYS    = TimeToDays
	TIME_TO_HOURS   = TimeToHours
	TIME_TO_MINUTES = TimeToMinutes
	TIME_TO_SECONDS = TimeToSeconds
)

// GetDiffTime get the diff time from previous to later time
// GetDiffTime 获取时间之间的天数,小时数, 分钟数和秒数
// parameter previousTime, should be string or time.Time
// parameter laterTime, should be string or time.Time
// parameter flag, value is TIME_TO_DAYS,TIME_TO_HOURS, TIME_TO_MINUTES, TIME_TO_SECONDS
func GetDiffTime(previousTime, laterTime interface{}, flag int) (int64, error) {
	timeType1 := reflect.TypeOf(laterTime)
	timeType2 := reflect.TypeOf(previousTime)
	if timeType1.String() == "string" && timeType2.String() == "string" {
		diffDays, err := getDiffStr(laterTime.(string), previousTime.(string), flag)
		if err != nil {
			return -1, err
		}
		return diffDays, nil

	}
	if timeType1.String() == "time.Time" && timeType2.String() == "time.Time" {
		diffDays := getDiff(laterTime.(time.Time), previousTime.(time.Time), flag)
		return diffDays, nil
	}
	return -1, errors.New("format error, arg must be string or time.Time")
}

// getDiff 获取两个时间相差的单位数
func getDiff(previous, later time.Time, flag int) int64 {
	var ret int64
	switch flag {
	case TimeToDays:
		later = time.Date(later.Year(), later.Month(), later.Day(), 0, 0, 0, 0, time.Local)
		previous = time.Date(previous.Year(), previous.Month(), previous.Day(), 0,
			0, 0, 0, time.Local)
		ret = int64(math.Abs(later.Sub(previous).Hours() / 24))
	case TimeToHours:
		ret = int64(math.Abs(later.Sub(previous).Hours()))
	case TimeToMinutes:
		ret = int64(math.Abs(later.Sub(previous).Minutes()))
	case TimeToSeconds:
		ret = int64(math.Abs(later.Sub(previous).Seconds()))
	default:
		ret = 0
	}
	return ret
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

func getDiffStr(laterStr, previousStr string, flag int) (int64, error) {
	var laterTime, previousTime time.Time
	var err1, err2 error
	len1 := len(laterStr)
	len2 := len(previousStr)

	switch len1 {
	case 8:
		laterTime, err1 = time.ParseInLocation(TimeShortDate, laterStr, time.Local)
	case 10:
		laterTime, err1 = time.ParseInLocation(TimeDate, laterStr, time.Local)
	case 19:
		laterTime, err1 = time.ParseInLocation(TimeCommon, laterStr, time.Local)
	default:
		err1 = errors.New("later format not supported")
	}

	switch len2 {
	case 8:
		previousTime, err2 = time.ParseInLocation(TimeShortDate, previousStr, time.Local)
	case 10:
		previousTime, err2 = time.ParseInLocation(TimeDate, previousStr, time.Local)
	case 19:
		previousTime, err2 = time.ParseInLocation(TimeCommon, previousStr, time.Local)
	default:
		err2 = errors.New("previous format not supported")
	}

	if err1 != nil {
		return -1, err1
	}
	if err2 != nil {
		return -1, err2
	}

	diffDays := getDiff(previousTime, laterTime, flag)
	return diffDays, nil
}

// UnixTimeStamp2TimeStr timestamp转换为标准日期时间
func UnixTimeStamp2TimeStr(sec int64) string {
	return time.Unix(sec, 0).Format(TimeCommon)
}

// DayListBetweenStartEnd 获取两个日期之间的所有日期
func DayListBetweenStartEnd(start, end string) ([]string, error) {
	dayList := make([]string, 0)
	days, err := getDiffStr(start, end, TimeToDays)
	if err != nil {
		return nil, err
	}
	t1, err := time.ParseInLocation(TimeDate, start, time.Local)
	if err != nil {
		return nil, err
	}

	for i := 0; i < int(days)+1; i++ {
		tmpDate := t1.AddDate(0, 0, i)
		dayList = append(dayList, tmpDate.Format(TimeDate))
	}
	return dayList, nil
}

// Yesterday 获取昨天的日期
func Yesterday(today string) (string, error) {
	var nTime = time.Time{}
	var err error
	if today == "" {
		nTime = time.Now()
	} else {
		nTime, err = time.ParseInLocation(TimeDate, today, time.Local)
		if err != nil {
			return "", err
		}
	}
	yesterdayTime := nTime.AddDate(0, 0, -1)
	logDay := yesterdayTime.Format(TimeDate)
	return logDay, nil
}

// Tomorrow 获取明天的日期
func Tomorrow(today string) (string, error) {
	var nTime = time.Time{}
	var err error
	if today == "" {
		nTime = time.Now()
	} else {
		nTime, err = time.ParseInLocation(TimeDate, today, time.Local)
		if err != nil {
			return "", err
		}
	}
	tomorrowTime := nTime.AddDate(0, 0, 1)
	logDay := tomorrowTime.Format(TimeDate)
	return logDay, nil
}

// GetCurrentAndNextHour 获取当前时间和后一小时
func GetCurrentAndNextHour(timeStr string) (string, string, error) {
	now, err := time.ParseInLocation(TimeCommon, timeStr, time.Local)
	if err != nil {
		return "", "", err
	}
	currentHour := now.Truncate(time.Hour).Format(TimeCommon)
	nextHour := now.Truncate(time.Hour).Add(time.Hour).Format(TimeCommon)
	return currentHour, nextHour, nil
}

// GetWeekendDates 获取指定年份的第几周的周日和周六
func GetWeekendDates(year, week int) (time.Time, time.Time) {
	// 获取指定年份的第一天
	firstDay := time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)

	// 计算第一天是周几
	weekday := int(firstDay.Weekday())

	// 计算第一周的第一个周日的日期
	firstSunday := firstDay.AddDate(0, 0, 7-weekday)

	// 计算指定周的周日和周六的日期
	sunday := firstSunday.AddDate(0, 0, (week-1)*7)
	saturday := sunday.AddDate(0, 0, 6)

	return sunday, saturday
}

// GetLastDayOfMonth 获取指定年月的最后一天
func GetLastDayOfMonth(year, month int) time.Time {
	// 获取下个月的第一天
	nextMonth := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.Local)

	// 减去一天得到当前月份的最后一天
	lastDay := nextMonth.AddDate(0, 0, -1)

	return lastDay
}

// GetWeekNumByDate 获取第几周
func GetWeekNumByDate(date string) int {
	d, _ := time.Parse(time.DateOnly, date)
	// 计算第几周
	_, weekNum := d.ISOWeek()
	return weekNum
}

// GetCalDataStr 计算日期加减后的string
func GetCalDataStr(date string, delta int) string {
	d, err := time.Parse(TimeDate, date)
	if err != nil {
		return ""
	}
	return d.AddDate(0, 0, delta).Format(TimeDate)
}

// GetTimestampGap 获取时间间隔的起始时间戳，用来做基于时间的幂等函数
func GetTimestampGap(intervalMinutes int) int64 {
	now := time.Now()
	// 计算当前时间属于哪个时间间隔
	intervalStart := now.Truncate(time.Duration(intervalMinutes) * time.Minute)
	return intervalStart.Unix()
}

// 获取当前时间的时间戳，基于指定的秒数
func GetCurrentIntervalTimestamp(intervalSeconds int) int64 {
	now := time.Now()
	// 计算当前时间属于哪个时间间隔
	intervalStart := now.Truncate(time.Duration(intervalSeconds) * time.Second)
	return intervalStart.Unix()
}