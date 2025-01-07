package structtool

import (
	"fmt"
	"testing"
)

// 定义结构体
type Person struct {
	Name    string
	Age     int
	Address []Address // 子结构体
	CityMap map[string]string
}

type Address struct {
	City  string
	State string
}

func TestCloneStruct(t *testing.T) {
	person := &Person{Name: "John", Age: 30}
	newPerson := NewEmptyInstance(person)
	fmt.Printf("Original Person: %+v\n", person)
	fmt.Printf("New Person: %+v\n", newPerson)
}

func TestGetStructFields(t *testing.T) {
	person := Person{Name: "John", Age: 30}
	name := GetStructName(person)
	t.Logf("name: %s", name)
}

func TestPrintStruct(t *testing.T) {
	// 创建一个 Person 实例，包含 Address 子结构体
	p := Person{
		Name: "Alice",
		Age:  30,
		Address: []Address{
			{City: "New York", State: "NY"},
			{City: "San Francisco", State: "CA"},
		},
		CityMap: map[string]string{
			"home": "Beijing",
			"work": "Shanghai",
		},
	}
	PrintStructAll(p)
}
