package cal

import (
	"testing"
	"time"
)

// go test -v calDate_test.go calDate.go calIP_test.go calIP.go

func TestGetDiffDays(t *testing.T) {
	t1 := time.Date(2023, 2, 14, 1, 1, 1, 1, time.Local)
	t2 := time.Date(2023, 2, 19, 1, 1, 1, 1, time.Local)
	days := GetDiffDays(t2, t1)
	if days != 5 {
		t.Error("GetDiffDays err")
	}
}

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
	t1 := time.Date(2023, 2, 14, 1, 1, 1, 1, time.Local)
	monday, sunday := GetFirstAndLastDateOfWeek(t1)
	t.Log(monday, sunday)
	if monday != "2023-02-13" || sunday != "2023-02-19" {
		t.Error("GetFirstAndLastDateOfWeek err:", monday, sunday)
	}
}
