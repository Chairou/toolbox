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
	ans, _ := Int("-9999")
	if ans != -9999 {
		t.Error("must be int -9999, not string", ans)
	}
}

func TestUInt(t *testing.T) {
	ans, _ := Uint("9999")
	if ans != uint(9999) {
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

func TestJsonToArray(t *testing.T) {
	jsonStr := `{"-a":"123","-b":"hello"}`
	array, err := JsonToArray(jsonStr)
	if err != nil {
		t.Error("JsonToArray err:", err)
	}
	t.Log(array)
	if !(array[0] == "-a" && array[1] == "123" && array[2] == "-b" && array[3] == "hello") {
		t.Error("JsonToArray expected err")
	}
}

func TestStringToMap(t *testing.T) {
	str := "-a 123 -b hello"
	toMap, err := StringToMap(str)
	if err != nil {
		t.Error("StringToMap err:", err)
	}
	t.Log(toMap)
	if !(toMap["-a"] == "123" && toMap["-b"] == "hello") {
		t.Error("StringToMap expected err")
	}
}

func TestJsonToString(t *testing.T) {
	jsonStr := `{"-a":"123","-b":"hello"}`
	str, err := JsonToString(jsonStr)
	if err != nil {
		t.Error("JsonToArray err:", err)
	}
	t.Log(str)
	if !(str == "-a 123 -b hello" || str == "-b hello -a 123") {
		t.Error("JsonToArray expected err")
	}
}

func TestMapToStruct(t *testing.T) {
	type ExampleStruct struct {
		Name   string
		Age    int
		Gender string
	}

	m := map[string]interface{}{
		"Name":   "John",
		"Age":    30,
		"Gender": "Male",
	}

	var s ExampleStruct
	err := MapToStruct(m, &s)
	if err != nil {
		t.Error("Error:", err)
		return
	}
	t.Logf("Result: %#v\n", s)
}

func TestIntToByte(t *testing.T) {
	var aa int = 16
	byteList := IntToByte(aa, 4)
	t.Log(byteList)
	if byteList[0] != 16 {
		t.Error("CheckIntToByte err:")
	}
}

func TestByteToString(t *testing.T) {
	var data interface{} = []byte("Hello, World!")
	str := String(data)
	if str != "Hello, World!" {
		t.Error("CheckByteToString err:")
	}
	t.Log(str)
}

// ==================== 以下为覆盖修改的测试用例 ====================

// ---------- IsNil 修复测试 ----------

// 测试 IsNil 对非指针类型不会 panic（修复前会 panic）
func TestIsNil_NonPointerType(t *testing.T) {
	// 传入普通 struct，不应 panic，应返回 false
	type Foo struct{ X int }
	result := IsNil(Foo{X: 1})
	if result != false {
		t.Error("IsNil(Foo{X:1}) should be false, got true")
	}

	// 传入 int
	result = IsNil(42)
	if result != false {
		t.Error("IsNil(42) should be false, got true")
	}

	// 传入 string
	result = IsNil("hello")
	if result != false {
		t.Error("IsNil(\"hello\") should be false, got true")
	}
}

// 测试 IsNil 对指向 struct 的指针不会 panic（修复前会 panic）
func TestIsNil_PointerToStruct(t *testing.T) {
	type Foo struct{ X int }
	p := &Foo{X: 1}
	result := IsNil(p)
	if result != false {
		t.Error("IsNil(&Foo{X:1}) should be false, got true")
	}
}

// 测试 IsNil 对 nil 指针返回 true
func TestIsNil_NilPointer(t *testing.T) {
	var p *int
	result := IsNil(p)
	if result != true {
		t.Error("IsNil(nil *int) should be true, got false")
	}
}

// 测试 IsNil 对多层 nil 指针返回 true
func TestIsNil_MultiLevelNilPointer(t *testing.T) {
	var p *int
	pp := &p // **int，其中 *p == nil
	result := IsNil(pp)
	if result != true {
		t.Error("IsNil(**int with nil inner) should be true, got false")
	}
}

// 测试 IsNil 对 nil map 返回 true（修复前返回 false）
func TestIsNil_NilMap(t *testing.T) {
	var m map[string]int
	result := IsNil(m)
	if result != true {
		t.Error("IsNil(nil map) should be true, got false")
	}
}

// 测试 IsNil 对非 nil map 返回 false
func TestIsNil_NonNilMap(t *testing.T) {
	m := make(map[string]int)
	result := IsNil(m)
	if result != false {
		t.Error("IsNil(non-nil map) should be false, got true")
	}
}

// 测试 IsNil 对 nil slice 返回 true（修复前返回 false）
func TestIsNil_NilSlice(t *testing.T) {
	var s []int
	result := IsNil(s)
	if result != true {
		t.Error("IsNil(nil slice) should be true, got false")
	}
}

// 测试 IsNil 对非 nil slice 返回 false
func TestIsNil_NonNilSlice(t *testing.T) {
	s := make([]int, 0)
	result := IsNil(s)
	if result != false {
		t.Error("IsNil(non-nil slice) should be false, got true")
	}
}

// 测试 IsNil 对 nil chan 返回 true（修复前返回 false）
func TestIsNil_NilChan(t *testing.T) {
	var ch chan int
	result := IsNil(ch)
	if result != true {
		t.Error("IsNil(nil chan) should be true, got false")
	}
}

// 测试 IsNil 对非 nil chan 返回 false
func TestIsNil_NonNilChan(t *testing.T) {
	ch := make(chan int)
	result := IsNil(ch)
	if result != false {
		t.Error("IsNil(non-nil chan) should be false, got true")
	}
}

// 测试 IsNil 对 nil func 返回 true（修复前返回 false）
func TestIsNil_NilFunc(t *testing.T) {
	var fn func()
	result := IsNil(fn)
	if result != true {
		t.Error("IsNil(nil func) should be true, got false")
	}
}

// 测试 IsNil 对非 nil func 返回 false
func TestIsNil_NonNilFunc(t *testing.T) {
	fn := func() {}
	result := IsNil(fn)
	if result != false {
		t.Error("IsNil(non-nil func) should be false, got true")
	}
}

// ---------- IntToByte 修复测试 ----------

// 测试 IntToByte 运算符优先级修复后的正确性
func TestIntToByte_MultiByteValue(t *testing.T) {
	// 0x0102 = 258，小端序应为 [0x02, 0x01, 0x00, 0x00]
	byteList := IntToByte(0x0102, 4)
	if byteList[0] != 0x02 {
		t.Errorf("IntToByte(0x0102, 4)[0] expected 0x02, got 0x%02x", byteList[0])
	}
	if byteList[1] != 0x01 {
		t.Errorf("IntToByte(0x0102, 4)[1] expected 0x01, got 0x%02x", byteList[1])
	}
	if byteList[2] != 0x00 {
		t.Errorf("IntToByte(0x0102, 4)[2] expected 0x00, got 0x%02x", byteList[2])
	}
	if byteList[3] != 0x00 {
		t.Errorf("IntToByte(0x0102, 4)[3] expected 0x00, got 0x%02x", byteList[3])
	}
}

// 测试 IntToByte 对 0xFF 的正确性
func TestIntToByte_SingleByte(t *testing.T) {
	byteList := IntToByte(0xFF, 1)
	if byteList[0] != 0xFF {
		t.Errorf("IntToByte(0xFF, 1)[0] expected 0xFF, got 0x%02x", byteList[0])
	}
}

// 测试 IntToByte 对较大值的正确性（修复前运算符优先级错误会导致结果不正确）
func TestIntToByte_LargeValue(t *testing.T) {
	// 0xAABBCCDD 小端序应为 [0xDD, 0xCC, 0xBB, 0xAA]
	val := 0xAABBCCDD
	byteList := IntToByte(val, 4)
	expected := []byte{0xDD, 0xCC, 0xBB, 0xAA}
	for i, exp := range expected {
		if byteList[i] != exp {
			t.Errorf("IntToByte(0xAABBCCDD, 4)[%d] expected 0x%02x, got 0x%02x", i, exp, byteList[i])
		}
	}
}

// 测试 IntToByte 对 0 的正确性
func TestIntToByte_Zero(t *testing.T) {
	byteList := IntToByte(0, 4)
	for i, b := range byteList {
		if b != 0 {
			t.Errorf("IntToByte(0, 4)[%d] expected 0x00, got 0x%02x", i, b)
		}
	}
}

// ---------- StringToArray 引号修复测试 ----------

// 测试 StringToArray 双引号包裹含空格的值（修复前 sOff/dOff 混淆导致错误）
func TestStringToArray_DoubleQuote(t *testing.T) {
	arr, err := StringToArray(`-a "hello world" -b test`)
	if err != nil {
		t.Error("StringToArray err:", err)
	}
	if len(arr) != 4 {
		t.Errorf("expected 4 elements, got %d: %v", len(arr), arr)
		return
	}
	if arr[0] != "-a" || arr[1] != "hello world" || arr[2] != "-b" || arr[3] != "test" {
		t.Errorf("unexpected result: %v", arr)
	}
}

// 测试 StringToArray 单引号包裹含空格的值（修复前 sOff/dOff 混淆导致错误）
func TestStringToArray_SingleQuote(t *testing.T) {
	arr, err := StringToArray(`-a 'hello world' -b test`)
	if err != nil {
		t.Error("StringToArray err:", err)
	}
	if len(arr) != 4 {
		t.Errorf("expected 4 elements, got %d: %v", len(arr), arr)
		return
	}
	if arr[0] != "-a" || arr[1] != "hello world" || arr[2] != "-b" || arr[3] != "test" {
		t.Errorf("unexpected result: %v", arr)
	}
}

// 测试 StringToArray 双引号内嵌套单引号
func TestStringToArray_DoubleQuoteWithSingleInside(t *testing.T) {
	arr, err := StringToArray(`-a "it's fine" -b ok`)
	if err != nil {
		t.Error("StringToArray err:", err)
	}
	if len(arr) != 4 {
		t.Errorf("expected 4 elements, got %d: %v", len(arr), arr)
		return
	}
	if arr[1] != "it's fine" {
		t.Errorf("expected \"it's fine\", got %q", arr[1])
	}
}

// 测试 StringToArray 单引号内嵌套双引号
func TestStringToArray_SingleQuoteWithDoubleInside(t *testing.T) {
	arr, err := StringToArray(`-a 'say "hi"' -b ok`)
	if err != nil {
		t.Error("StringToArray err:", err)
	}
	if len(arr) != 4 {
		t.Errorf("expected 4 elements, got %d: %v", len(arr), arr)
		return
	}
	if arr[1] != `say "hi"` {
		t.Errorf("expected 'say \"hi\"', got %q", arr[1])
	}
}

// 测试 StringToArray 奇数长度不再报错（偶数检查已移至 StringToMap）
func TestStringToArray_OddLength(t *testing.T) {
	arr, err := StringToArray("a1 b2 c3")
	if err != nil {
		t.Error("StringToArray should not return error for odd length, got:", err)
	}
	if len(arr) != 3 {
		t.Errorf("expected 3 elements, got %d: %v", len(arr), arr)
	}
}

// ---------- StringToMap 偶数检查测试 ----------

// 测试 StringToMap 对奇数长度数组返回错误（偶数检查从 StringToArray 移至此处）
func TestStringToMap_OddLengthError(t *testing.T) {
	_, err := StringToMap("a1 b2 c3")
	if err == nil {
		t.Error("StringToMap should return error for odd length array")
	}
}

// 测试 StringToMap 正常偶数长度
func TestStringToMap_EvenLength(t *testing.T) {
	m, err := StringToMap("-a 123 -b hello")
	if err != nil {
		t.Error("StringToMap err:", err)
	}
	if m["-a"] != "123" || m["-b"] != "hello" {
		t.Errorf("unexpected result: %v", m)
	}
}

// 测试 StringToMap 空字符串
func TestStringToMap_Empty(t *testing.T) {
	m, err := StringToMap("")
	if err != nil {
		t.Error("StringToMap err:", err)
	}
	if m != nil {
		t.Errorf("expected nil map for empty string, got: %v", m)
	}
}

// ---------- JsonToArray 修复测试 ----------

// 测试 JsonToArray 值以 "-" 开头不再被丢弃（修复前会被 continue 跳过）
func TestJsonToArray_ValueWithDash(t *testing.T) {
	jsonStr := `{"-a":"-123","-b":"hello"}`
	array, err := JsonToArray(jsonStr)
	if err != nil {
		t.Error("JsonToArray err:", err)
	}
	// 应该有 4 个元素，-a 和 -123 不应被丢弃
	if len(array) != 4 {
		t.Errorf("expected 4 elements, got %d: %v", len(array), array)
		return
	}
	// 由于 map 遍历顺序不确定，用 map 验证
	m := make(map[string]string)
	for i := 0; i < len(array); i += 2 {
		m[array[i]] = array[i+1]
	}
	if m["-a"] != "-123" {
		t.Errorf("expected -a=-123, got -a=%s", m["-a"])
	}
	if m["-b"] != "hello" {
		t.Errorf("expected -b=hello, got -b=%s", m["-b"])
	}
}

// 测试 JsonToArray 所有值都以 "-" 开头（修复前全部被丢弃，返回空数组）
func TestJsonToArray_AllValuesWithDash(t *testing.T) {
	jsonStr := `{"-a":"-100","-b":"-200"}`
	array, err := JsonToArray(jsonStr)
	if err != nil {
		t.Error("JsonToArray err:", err)
	}
	if len(array) != 4 {
		t.Errorf("expected 4 elements, got %d: %v", len(array), array)
	}
}

// 测试 JsonToArray key 不带 "-" 前缀时自动补充
func TestJsonToArray_KeyWithoutDash(t *testing.T) {
	jsonStr := `{"a":"123"}`
	array, err := JsonToArray(jsonStr)
	if err != nil {
		t.Error("JsonToArray err:", err)
	}
	if len(array) != 2 {
		t.Errorf("expected 2 elements, got %d: %v", len(array), array)
		return
	}
	if array[0] != "-a" || array[1] != "123" {
		t.Errorf("expected [-a 123], got %v", array)
	}
}

// 测试 JsonToArray 空字符串
func TestJsonToArray_Empty(t *testing.T) {
	array, err := JsonToArray("")
	if err != nil {
		t.Error("JsonToArray err:", err)
	}
	if len(array) != 0 {
		t.Errorf("expected empty array, got %v", array)
	}
}

// ---------- JsonToString 修复测试 ----------

// 测试 JsonToString 值以 "-" 开头不再被丢弃（修复前会被 continue 跳过）
func TestJsonToString_ValueWithDash(t *testing.T) {
	jsonStr := `{"-a":"-123"}`
	str, err := JsonToString(jsonStr)
	if err != nil {
		t.Error("JsonToString err:", err)
	}
	if str != "-a -123" {
		t.Errorf("expected '-a -123', got '%s'", str)
	}
}

// 测试 JsonToString 所有值都以 "-" 开头（修复前返回空字符串）
func TestJsonToString_AllValuesWithDash(t *testing.T) {
	jsonStr := `{"-x":"-10"}`
	str, err := JsonToString(jsonStr)
	if err != nil {
		t.Error("JsonToString err:", err)
	}
	if str != "-x -10" {
		t.Errorf("expected '-x -10', got '%s'", str)
	}
}

// 测试 JsonToString key 不带 "-" 前缀时自动补充
func TestJsonToString_KeyWithoutDash(t *testing.T) {
	jsonStr := `{"a":"123"}`
	str, err := JsonToString(jsonStr)
	if err != nil {
		t.Error("JsonToString err:", err)
	}
	if str != "-a 123" {
		t.Errorf("expected '-a 123', got '%s'", str)
	}
}

// 测试 JsonToString 空字符串
func TestJsonToString_Empty(t *testing.T) {
	str, err := JsonToString("")
	if err != nil {
		t.Error("JsonToString err:", err)
	}
	if str != "" {
		t.Errorf("expected empty string, got '%s'", str)
	}
}
