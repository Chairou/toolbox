package cal

import (
	"testing"
	"time"
)

// go test -v calDate_test.go calDate.go calIP_test.go calIP.go

/*
*******************************
按照国际标准ISO 8601 的说法，星期一是一周的开始，
而星期日是一周的结束。 虽然已经有了国际标准，
但是很多国家，比如「美国」、「加拿大」和
「澳大利亚」等国家，依然以星期日作为一周的开始。
同时参考国家标准GB/T 7408-2005，
其中也明确表述了星期一为一周的开始。
*******************************
*/
func TestGetFirstAndLastDateOfWeek(t *testing.T) {
	t1 := time.Date(2022, 1, 1, 1, 1, 1, 1, time.Local)
	monday, sunday := GetFirstAndLastDateOfWeek(t1)
	t.Log(monday, sunday)
	if monday != "2023-02-13" || sunday != "2023-02-19" {
		t.Error("GetFirstAndLastDateOfWeek err:", monday, sunday)
	}
}

func TestGetDiffDays(t *testing.T) {
	t1 := "2023-02-14 01:01:01"
	t2 := "2023-02-19 01:01:01"
	days, err := GetDiffTime(t1, t2, TIME_TO_DAYS)
	if err != nil {
		t.Error(err)
	}
	if days != 5 {
		t.Error("GetDiffDays err")
	}
	t.Log(days)

	t3 := time.Date(2023, 2, 14, 1, 1, 1, 1, time.Local)
	t4 := time.Date(2023, 2, 19, 1, 1, 1, 1, time.Local)
	days, err = GetDiffTime(t3, t4, TIME_TO_DAYS)
	if err != nil {
		t.Error(err)
	}
	if days != 5 {
		t.Error("GetDiffDays err")
	}
	t.Log(days)

}

func TestGetDiffHours(t *testing.T) {
	t1 := "2023-01-14 00:11:10"
	t2 := "2023-02-13 00:10:10"

	hours, err := GetDiffTime(t1, t2, TIME_TO_HOURS)
	if err != nil {
		t.Error(err)
	}
	t.Log(hours)
	if hours != 719 {
		t.Error("GetDiffHours err")
	}
	t3 := time.Date(2023, 01, 14, 00, 11, 10, 0, time.Local)
	t4 := time.Date(2023, 02, 13, 00, 10, 10, 0, time.Local)
	hours, err = GetDiffTime(t3, t4, TIME_TO_HOURS)
	if err != nil {
		t.Error(err)
	}
	t.Log(hours)
	if hours != 719 {
		t.Error("GetDiffHours err")
	}
}

func TestGetDiffMinutes(t *testing.T) {
	t1 := "2023-01-13 00:10:00"
	t2 := "2023-02-14 00:11:00"

	minutes, err := GetDiffTime(t1, t2, TIME_TO_MINUTES)
	if err != nil {
		t.Error(err)
	}
	t.Log(minutes)
	if minutes != 46081 {
		t.Error("GetDiffMinutes err")
	}

	t3 := time.Date(2023, 01, 13, 00, 10, 00, 0, time.Local)
	t4 := time.Date(2023, 02, 14, 00, 11, 00, 0, time.Local)
	minutes, err = GetDiffTime(t3, t4, TIME_TO_MINUTES)

	if err != nil {
		t.Error(err)
	}
	t.Log(minutes)
	if minutes != 46081 {
		t.Error("GetDiffMinutes err")
	}
}

func TestGetDiffSeconds(t *testing.T) {
	t1 := "2023-01-13 00:10:00"
	t2 := "2023-02-14 00:11:00"

	minutes, err := GetDiffTime(t1, t2, TIME_TO_SECONDS)
	if err != nil {
		t.Error(err)
	}
	t.Log(minutes)
	if minutes != 2764860 {
		t.Error("GetDiffMinutes err")
	}

	t3 := time.Date(2023, 01, 13, 00, 10, 00, 0, time.Local)
	t4 := time.Date(2023, 02, 14, 00, 11, 00, 0, time.Local)
	minutes, err = GetDiffTime(t3, t4, TIME_TO_SECONDS)

	if err != nil {
		t.Error(err)
	}
	t.Log(minutes)
	if minutes != 2764860 {
		t.Error("GetDiffMinutes err")
	}
}

func TestLocation(t *testing.T) {
	loc, _ := time.LoadLocation("Cuba")
	now := time.Now()
	t1, err := time.ParseInLocation(TIME_COMMON, now.Format(TIME_COMMON), loc)
	if err != nil {
		return
	}
	t.Log(t1)
}

func TestDayListBetweenStartEnd(t *testing.T) {
	startDate := "2023-02-25"
	endDate := "2023-03-10"
	list, err := DayListBetweenStartEnd(startDate, endDate)
	if err != nil {
		t.Error(err)
	}
	t.Log(list)
}

func TestYesterday(t *testing.T) {
	today := "2023-04-01"
	yesterday, err := Yesterday(today)
	if err != nil {
		t.Error(err)
	}
	t.Log(yesterday)
}

func TestTomorrow(t *testing.T) {
	today := "2023-04-01"
	tomorrow, err := Tomorrow(today)
	if err != nil {
		t.Error(err)
	}
	t.Log(tomorrow)
}

func TestCalBefore(t *testing.T) {
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -7)
	t.Log(startTime)
}

func TestWeekNum(t *testing.T) {
	t.Log(GetWeekNumByDate("2023-08-01"))
}
