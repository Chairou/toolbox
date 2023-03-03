package test

import (
	"fmt"
	"github.com/dop251/goja"
	"testing"
)

type person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func abc(a int, b int) []int {
	return []int{a, b}
}

func str() map[string]interface{} {
	return map[string]interface{}{"a": "a", "b": "b"}
}

func personOne() []person {
	p := person{
		"HaiCoder",
		110,
	}
	list := make([]person, 0)
	list = append(list, p)
	return list
}

func TestGoja(t *testing.T) {
	vm := goja.New()
	vm.Set("str", personOne)
	//v, _ := vm.RunString(`aaa=abc(2,3);`)
	//v, _ := vm.RunString(`aaa=str();`)
	v, _ := vm.RunString(`aaa=str();
	bbb=9;
	aaa[0];
`)
	//fmt.Println(v.Export().(map[string]interface{}))
	fmt.Println(v.Export().(person).Name)
	fmt.Println(vm.Get("bbb"))
}
