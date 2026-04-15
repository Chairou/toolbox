package structtool

import (
	"reflect"
	"testing"
)

// person 测试用结构体
type person struct {
	Name    string
	Age     int
	Address []address
	CityMap map[string]string
}

// address 测试用子结构体
type address struct {
	City  string
	State string
}

// mixedStruct 包含未导出字段的结构体
type mixedStruct struct {
	Public  string
	private string
}

// nestedStruct 嵌套结构体
type nestedStruct struct {
	Inner person
	Tag   string
}

// emptyStruct 空结构体
type emptyStruct struct{}

// ==================== NewEmptyInstance 测试 ====================

func TestNewEmptyInstance_Basic(t *testing.T) {
	p := &person{Name: "John", Age: 30}
	newP := NewEmptyInstance(p)
	if newP == nil {
		t.Fatal("NewEmptyInstance returned nil")
	}
	// 新实例应该是零值
	if newP.Name != "" || newP.Age != 0 {
		t.Errorf("expected zero value, got Name=%q, Age=%d", newP.Name, newP.Age)
	}
	// 新实例和原实例应该是不同的指针
	if newP == p {
		t.Error("NewEmptyInstance returned the same pointer")
	}
}

func TestNewEmptyInstance_WithSliceAndMap(t *testing.T) {
	p := &person{
		Name:    "Alice",
		Age:     25,
		Address: []address{{City: "Beijing", State: "BJ"}},
		CityMap: map[string]string{"home": "Beijing"},
	}
	newP := NewEmptyInstance(p)
	if newP == nil {
		t.Fatal("NewEmptyInstance returned nil")
	}
	if newP.Address != nil {
		t.Errorf("expected nil Address, got %v", newP.Address)
	}
	if newP.CityMap != nil {
		t.Errorf("expected nil CityMap, got %v", newP.CityMap)
	}
}

func TestNewEmptyInstance_EmptyStruct(t *testing.T) {
	e := &emptyStruct{}
	newE := NewEmptyInstance(e)
	if newE == nil {
		t.Fatal("NewEmptyInstance returned nil for empty struct")
	}
}

func TestNewEmptyInstance_IntType(t *testing.T) {
	val := 42
	newVal := NewEmptyInstance(&val)
	if newVal == nil {
		t.Fatal("NewEmptyInstance returned nil")
	}
	if *newVal != 0 {
		t.Errorf("expected 0, got %d", *newVal)
	}
}

func TestNewEmptyInstance_StringType(t *testing.T) {
	val := "hello"
	newVal := NewEmptyInstance(&val)
	if newVal == nil {
		t.Fatal("NewEmptyInstance returned nil")
	}
	if *newVal != "" {
		t.Errorf("expected empty string, got %q", *newVal)
	}
}

// ==================== GetStructName 测试 ====================

func TestGetStructName_Basic(t *testing.T) {
	p := person{Name: "John", Age: 30}
	name := GetStructName(p)
	if name != "person" {
		t.Errorf("expected 'person', got %q", name)
	}
}

func TestGetStructName_Pointer(t *testing.T) {
	p := &person{Name: "John", Age: 30}
	name := GetStructName(p)
	if name != "person" {
		t.Errorf("expected 'person', got %q", name)
	}
}

func TestGetStructName_DoublePointer(t *testing.T) {
	p := &person{Name: "John"}
	pp := &p
	name := GetStructName(pp)
	if name != "person" {
		t.Errorf("expected 'person', got %q", name)
	}
}

func TestGetStructName_Nil(t *testing.T) {
	name := GetStructName(nil)
	if name != "" {
		t.Errorf("expected empty string for nil, got %q", name)
	}
}

func TestGetStructName_EmptyStruct(t *testing.T) {
	e := emptyStruct{}
	name := GetStructName(e)
	if name != "emptyStruct" {
		t.Errorf("expected 'emptyStruct', got %q", name)
	}
}

func TestGetStructName_BuiltinType(t *testing.T) {
	name := GetStructName(42)
	if name != "int" {
		t.Errorf("expected 'int', got %q", name)
	}
}

// ==================== PrintStructAll 测试 ====================

func TestPrintStructAll_Basic(t *testing.T) {
	p := person{
		Name: "Alice",
		Age:  30,
		Address: []address{
			{City: "New York", State: "NY"},
			{City: "San Francisco", State: "CA"},
		},
		CityMap: map[string]string{
			"home": "Beijing",
			"work": "Shanghai",
		},
	}
	// 不应 panic
	PrintStructAll(p)
}

func TestPrintStructAll_Pointer(t *testing.T) {
	p := &person{
		Name: "Bob",
		Age:  25,
	}
	// 传入指针不应 panic
	PrintStructAll(p)
}

func TestPrintStructAll_Nil(t *testing.T) {
	// 传入 nil 不应 panic
	PrintStructAll(nil)
}

func TestPrintStructAll_NilPointer(t *testing.T) {
	var p *person
	// 传入 nil 指针不应 panic
	PrintStructAll(p)
}

func TestPrintStructAll_EmptyStruct(t *testing.T) {
	e := emptyStruct{}
	// 空结构体不应 panic
	PrintStructAll(e)
}

func TestPrintStructAll_WithUnexportedFields(t *testing.T) {
	m := mixedStruct{Public: "visible", private: "hidden"}
	// 包含未导出字段不应 panic
	PrintStructAll(m)
}

func TestPrintStructAll_NestedStruct(t *testing.T) {
	n := nestedStruct{
		Inner: person{Name: "Inner", Age: 10},
		Tag:   "test",
	}
	// 嵌套结构体不应 panic
	PrintStructAll(n)
}

// ==================== PrintStruct 测试 ====================

func TestPrintStruct_Basic(t *testing.T) {
	p := person{Name: "Alice", Age: 30}
	v := reflect.ValueOf(p)
	// 不应 panic
	PrintStruct(v, "")
}

func TestPrintStruct_WithUnexportedFields(t *testing.T) {
	m := mixedStruct{Public: "visible", private: "hidden"}
	v := reflect.ValueOf(m)
	// 包含未导出字段不应 panic
	PrintStruct(v, "")
}

func TestPrintStruct_NonStruct(t *testing.T) {
	v := reflect.ValueOf(42)
	// 非结构体类型不应 panic
	PrintStruct(v, "")
}

func TestPrintStruct_NestedStruct(t *testing.T) {
	n := nestedStruct{
		Inner: person{Name: "Inner", Age: 10},
		Tag:   "test",
	}
	v := reflect.ValueOf(n)
	// 嵌套结构体不应 panic
	PrintStruct(v, "")
}

// ==================== PrintStruct2 测试 ====================

func TestPrintStruct2_Struct(t *testing.T) {
	p := person{
		Name: "Alice",
		Age:  30,
		Address: []address{
			{City: "New York", State: "NY"},
		},
		CityMap: map[string]string{"home": "Beijing"},
	}
	v := reflect.ValueOf(p)
	// 不应 panic
	PrintStruct2(v, "")
}

func TestPrintStruct2_Slice(t *testing.T) {
	s := []address{
		{City: "New York", State: "NY"},
		{City: "San Francisco", State: "CA"},
	}
	v := reflect.ValueOf(s)
	// 直接传入 slice 不应 panic
	PrintStruct2(v, "")
}

func TestPrintStruct2_Map(t *testing.T) {
	m := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	v := reflect.ValueOf(m)
	// 直接传入 map 不应 panic
	PrintStruct2(v, "")
}

func TestPrintStruct2_EmptySlice(t *testing.T) {
	s := []address{}
	v := reflect.ValueOf(s)
	// 空 slice 不应 panic
	PrintStruct2(v, "")
}

func TestPrintStruct2_EmptyMap(t *testing.T) {
	m := map[string]string{}
	v := reflect.ValueOf(m)
	// 空 map 不应 panic
	PrintStruct2(v, "")
}

func TestPrintStruct2_WithUnexportedFields(t *testing.T) {
	m := mixedStruct{Public: "visible", private: "hidden"}
	v := reflect.ValueOf(m)
	// 包含未导出字段不应 panic
	PrintStruct2(v, "")
}

func TestPrintStruct2_NonStructSliceMap(t *testing.T) {
	v := reflect.ValueOf(42)
	// 非 struct/slice/map 类型不应 panic
	PrintStruct2(v, "")
}

func TestPrintStruct2_MapWithStructValue(t *testing.T) {
	m := map[string]address{
		"home": {City: "Beijing", State: "BJ"},
		"work": {City: "Shanghai", State: "SH"},
	}
	v := reflect.ValueOf(m)
	// map 值为结构体时不应 panic
	PrintStruct2(v, "")
}

func TestPrintStruct2_SliceOfSlice(t *testing.T) {
	s := [][]int{{1, 2}, {3, 4}}
	v := reflect.ValueOf(s)
	// 嵌套 slice 不应 panic
	PrintStruct2(v, "")
}

// ==================== StructToMap 测试 ====================

func TestStructToMap_Basic(t *testing.T) {
	p := person{Name: "Alice", Age: 30}
	result, err := StructToMap(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["Name"] != "Alice" {
		t.Errorf("expected Name='Alice', got %v", result["Name"])
	}
	if result["Age"] != 30 {
		t.Errorf("expected Age=30, got %v", result["Age"])
	}
}

func TestStructToMap_Pointer(t *testing.T) {
	p := &person{Name: "Bob", Age: 25}
	result, err := StructToMap(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["Name"] != "Bob" {
		t.Errorf("expected Name='Bob', got %v", result["Name"])
	}
}

func TestStructToMap_WithSliceAndMap(t *testing.T) {
	p := person{
		Name:    "Charlie",
		Age:     35,
		Address: []address{{City: "Beijing", State: "BJ"}},
		CityMap: map[string]string{"home": "Beijing"},
	}
	result, err := StructToMap(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 4 {
		t.Errorf("expected 4 fields, got %d", len(result))
	}
	addrs, ok := result["Address"].([]address)
	if !ok {
		t.Fatal("Address field type assertion failed")
	}
	if len(addrs) != 1 || addrs[0].City != "Beijing" {
		t.Errorf("unexpected Address value: %v", addrs)
	}
}

func TestStructToMap_UnexportedFields(t *testing.T) {
	m := mixedStruct{Public: "visible", private: "hidden"}
	result, err := StructToMap(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 只应包含导出字段
	if len(result) != 1 {
		t.Errorf("expected 1 field, got %d", len(result))
	}
	if result["Public"] != "visible" {
		t.Errorf("expected Public='visible', got %v", result["Public"])
	}
	if _, exists := result["private"]; exists {
		t.Error("unexported field 'private' should not be in result")
	}
}

func TestStructToMap_EmptyStruct(t *testing.T) {
	e := emptyStruct{}
	result, err := StructToMap(e)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d fields", len(result))
	}
}

func TestStructToMap_Nil(t *testing.T) {
	_, err := StructToMap(nil)
	if err == nil {
		t.Error("expected error for nil input")
	}
}

func TestStructToMap_NilPointer(t *testing.T) {
	var p *person
	_, err := StructToMap(p)
	if err == nil {
		t.Error("expected error for nil pointer input")
	}
}

func TestStructToMap_NonStruct(t *testing.T) {
	_, err := StructToMap(42)
	if err == nil {
		t.Error("expected error for non-struct input")
	}
}

func TestStructToMap_StringInput(t *testing.T) {
	_, err := StructToMap("hello")
	if err == nil {
		t.Error("expected error for string input")
	}
}