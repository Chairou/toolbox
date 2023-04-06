package cal

import (
	"errors"
	"math"
	"reflect"
	"time"
)

const (
	TIME_DATE      string = "2006-01-02"
	TIME_COMMON    string = "2006-01-02 15:04:05"
	TIME_SHORTDATE string = "20060102"

	TIME_TO_DAYS    int = 1
	TIME_TO_HOURS   int = 2
	TIME_TO_MINUTES int = 3
	TIME_TO_SECONDS int = 4
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
	case TIME_TO_DAYS:
		later = time.Date(later.Year(), later.Month(), later.Day(), 0, 0, 0, 0, time.Local)
		previous = time.Date(previous.Year(), previous.Month(), previous.Day(), 0,
			0, 0, 0, time.Local)
		ret = int64(math.Abs(later.Sub(previous).Hours() / 24))
	case TIME_TO_HOURS:
		ret = int64(math.Abs(later.Sub(previous).Hours()))
	case TIME_TO_MINUTES:
		ret = int64(math.Abs(later.Sub(previous).Minutes()))
	case TIME_TO_SECONDS:
		ret = int64(math.Abs(later.Sub(previous).Seconds()))
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

func getDiffStr(previous, later string, flag int) (int64, error) {
	var previousTime, laterTime time.Time
	var err1, err2 error
	len1 := len(previous)
	len2 := len(later)

	switch len1 {
	case 8:
		previousTime, err1 = time.ParseInLocation(TIME_SHORTDATE, previous, time.Local)
	case 10:
		previousTime, err1 = time.ParseInLocation(TIME_DATE, previous, time.Local)
	case 19:
		previousTime, err1 = time.ParseInLocation(TIME_COMMON, previous, time.Local)
	default:
		err1 = errors.New("previous format not supported")
	}

	switch len2 {
	case 8:
		laterTime, err2 = time.ParseInLocation(TIME_SHORTDATE, later, time.Local)
	case 10:
		laterTime, err2 = time.ParseInLocation(TIME_DATE, later, time.Local)
	case 19:
		laterTime, err2 = time.ParseInLocation(TIME_COMMON, later, time.Local)
	default:
		err2 = errors.New("later format not supported")
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

func UnixTimeStamp2TimeStr(sec int64) string {
	return time.Unix(sec, 0).Format(TIME_COMMON)
}

func DayListBetweenStartEnd(start, end string) ([]string, error) {
	dayList := make([]string, 0)
	days, err := getDiffStr(start, end, TIME_TO_DAYS)
	if err != nil {
		return nil, err
	}
	t1, err := time.ParseInLocation(TIME_DATE, start, time.Local)
	if err != nil {
		return nil, err
	}

	for i := 0; i < int(days)+1; i++ {
		tmpDate := t1.AddDate(0, 0, i)
		dayList = append(dayList, tmpDate.Format(TIME_DATE))
	}
	return dayList, nil
}

func Yesterday(today string) (string, error) {
	var nTime = time.Time{}
	var err error
	if today == "" {
		nTime = time.Now()
	} else {
		nTime, err = time.ParseInLocation(TIME_DATE, today, time.Local)
		if err != nil {
			return "", err
		}
	}
	yesterdayTime := nTime.AddDate(0, 0, -1)
	logDay := yesterdayTime.Format(TIME_DATE)
	return logDay, nil
}

func Tomorrow(today string) (string, error) {
	var nTime = time.Time{}
	var err error
	if today == "" {
		nTime = time.Now()
	} else {
		nTime, err = time.ParseInLocation(TIME_DATE, today, time.Local)
		if err != nil {
			return "", err
		}
	}
	tomorrowTime := nTime.AddDate(0, 0, 1)
	logDay := tomorrowTime.Format(TIME_DATE)
	return logDay, nil
}

func GetCurrentAndNextHour(timeStr string) (string, string, error) {
	now, err := time.ParseInLocation(TIME_COMMON, timeStr, time.Local)
	if err != nil {
		return "", "", err
	}
	currentHour := now.Truncate(time.Hour).Format(TIME_COMMON)
	nextHour := now.Truncate(time.Hour).Add(time.Hour).Format(TIME_COMMON)
	return currentHour, nextHour, nil
}
