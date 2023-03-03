package conv

import (
	"fmt"
	"testing"
	"time"
)

// go test -v conv_test.go conv.go
func TestString(t *testing.T) {
	ans := String(1234)
	if ans != "1234" {
		t.Error("must be String 1234, not int", ans)
	}
}

func TestInt64(t *testing.T) {
	ans, _ := Int64("9999")
	if ans != 9999 {
		t.Error("musst be Int64 9999, not string", ans)
	}
}

func TestUint64(t *testing.T) {
	ans, _ := Uint64("9999")
	if ans != 9999 {
		t.Error("must be uint 9999, not string", ans)
	}
}
func TestUint(t *testing.T) {
	ans, _ := Int64("9999")
	if ans != 9999 {
		t.Error("must be uint 9999, not string", ans)
	}
}

func TestBool(t *testing.T) {
	ans, _ := Bool(1)
	if ans != true {
		t.Error("must be true, not false", true)
	}
}

func TestInt(t *testing.T) {
	ans, _ := Int64("-9999")
	if ans != -9999 {
		t.Error("must be int -9999, not string", ans)
	}
}

func TestUInt(t *testing.T) {
	ans, _ := Uint64("9999")
	if ans != 9999 {
		t.Error("must be Uint 9999, not string", ans)
	}
}

func TestFloat64(t *testing.T) {
	ans, _ := Float64("11.11")
	if ans != 11.11 {
		t.Error("must be float64 11.11", ans)
	}
}

func TestIsNil(t *testing.T) {
	ans := IsNil(nil)
	if ans != true {
		t.Error("must be true ", ans)
	}
}

func TestTime(t *testing.T) {
	ans, _ := Time("2020-01-12 23:23:59")
	if ans.Year() != 2020 {
		t.Error("must be time.Time 2020", ans)
	}
}

func TestTimePtr(t *testing.T) {
	var ans *time.Time
	ans = TimePtr("2020-01-12 23:23:59")
	if ans.Year() != 2020 {
		t.Error("must be time.Time 2020", ans)
	}
}

func TestStringToArray(t *testing.T) {
	// 此函数要求返回长度为偶数的数组, 是因为需要被StringToMap调用, 所以必须为偶数
	arr, _ := StringToArray("a1 b2  c3 hello")
	if len(arr) == 0 {
		t.Error("must be len 3", arr)
	}
}

func TestStructToMap(t *testing.T) {
	type person struct {
		Name string `json:"Name"`
		Age  int    `json:"Age"`
	}
	p := person{Name: "Alice", Age: 30}
	retData := make(map[string]interface{}, 0)
	retData, _ = StructToMap(p, "json")
	if retData["Name"] != "Alice" {
		t.Error("must be Alice but ", retData["Name"])
	}
}

func BenchmarkStructToMap(b *testing.B) {
	type person struct {
		Name string `json:"Name"`
		Age  int    `json:"Age"`
	}
	p := person{Name: "Alice", Age: 30}
	retData := make(map[string]interface{}, 0)
	fmt.Println(retData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		retData, _ = StructToMap(p, "json")
	}
}