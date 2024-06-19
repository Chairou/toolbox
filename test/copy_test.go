package test

import (
	"fmt"
	"github.com/jinzhu/copier"
	"reflect"
	"testing"
)

type point struct {
	Name   string `copier:"notNilCopy"`
	Sex    *string
	Age    *int
	Length *int
	Level  int
}

type inst struct {
	Name   string
	Sex    string
	Age    int
	Length int
}

type G struct {
	Num   int64
	Level string
	F     F
}

type F struct {
	Value string
	C     *C
}

type C struct {
	Name string
	Id   int64
}

func TestCopy(t *testing.T) {
	var a point
	a.Name = ""
	a.Sex = new(string)
	*a.Sex = "man"
	a.Age = new(int)
	*a.Age = 44
	a.Level = 11

	var b inst
	b.Name = "zoe"
	err := copier.Copy(&b, &a)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("%#v", b)

	//getType := reflect.TypeOf(a)
	//getValue := reflect.ValueOf(a)
	//
	//for i:=0; i < getType.NumField(); i++ {
	//	field := getType.Field(i)
	//	value := getValue.Field(i).Interface()
	//	fmt.Printf("%s: %v = %v\n", field.Name, field.Type, value)
	//
	//}
	//
	//
	//
	//
	//c := C{Name: "namec", Id: 100}
	//f := F{Value: "bvalue", C: &c}
	//q := G{Num: 10, Level: "high", F: f}
	//
	//walkField(q)
	//CopyField(&a,&b)
	//fmt.Println("b=", b)

}

func walkField(i interface{}) {
	getType := reflect.TypeOf(i)
	getValue := reflect.ValueOf(i)
	for i := 0; i < getType.NumField(); i++ {
		field := getType.Field(i)
		value := getValue.Field(i)
		if field.Type.Kind() == reflect.Struct {
			fmt.Println("=========struct=========")
			walkField(field)
		}
		fmt.Printf("%s: %v = %v\n", field.Name, field.Type, value)

	}
}

// CopyField copy field from A to B
func CopyField(src, dst interface{}) {
	srcVal := reflect.ValueOf(src).Elem()
	dstVal := reflect.ValueOf(dst).Elem()
	for i := 0; i < srcVal.NumField(); i++ {
		value := srcVal.Field(i)
		name := srcVal.Type().Field(i).Name
		dstValue := dstVal.FieldByName(name)
		if !dstValue.IsValid() {
			continue
		}
		if srcVal.Type().Field(i).Type != dstValue.Type() {
			continue
		}
		if srcVal.Type().Kind() == reflect.Struct {
			//调整src和dst, 进入递归
		}
		dstValue.Set(value)
	}
}

func TestCopyStruct(t *testing.T) {
	type A struct {
		Name string
		Age  int
	}
	type B struct {
		Kill string
		A
	}
	a := A{
		Name: "a",
		Age:  1,
	}
	b := B{}
	copier.Copy(&b, &a)
	fmt.Printf("%+v", b)
}
