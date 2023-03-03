package test

import (
	"fmt"
	"github.com/dop251/goja"
	"sync"
	"testing"
)

var VmPool *sync.Pool

func initVmPool() {
	VmPool = &sync.Pool{
		New: func() interface{} {
			return goja.New()
		},
	}
}

func TestGoja2(t *testing.T) {
	initVmPool()
	code := `//asd 
filterObj={"a":"b"}; 
a=Object.keys(filterObj);eval(b+c);`
	preCompile, err := goja.Compile("111", code, false)
	if err != nil {
		t.Error(err)
		return
	}
	vm := VmPool.Get().(*goja.Runtime)
	defer VmPool.Put(vm)
	vm.Set("b", 1)
	vm.Set("c", vm.ToValue(2).ToString())
	//v, err := vm.RunString(code)
	v, err := vm.RunProgram(preCompile)
	if err != nil {
		t.Error(err)
		return
	}
	//retMap := v.Export().(map[string]interface{})
	retMap := v.Export()
	t.Log(retMap)

}

func TestGoja3(t *testing.T) {
	vm := goja.New()
	type S struct {
		Field float32 `json:"field"`
	}
	vm.Set("s", S{Field: 0.0})
	res, _ := vm.RunString(`s.Field`) // without the mapper it would have been s.Field
	fmt.Println(res.Export().(float32))
	// Output: 42
}
