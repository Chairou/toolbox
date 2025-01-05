package conf

import (
	"fmt"
	"github.com/Chairou/toolbox/util/structtool"
	"reflect"
	"testing"
)

type Person struct {
	Name string
	Age  int
}

func TestLoadAllConf(t *testing.T) {
	config := Config{}
	LoadAllConf(&config)
}

func TestReflect(t *testing.T) {
	p := Person{
		Name: "Alice",
		Age:  30,
	}

	// 获取 p 的反射值对象
	value := reflect.ValueOf(p)

	// 获取 Person 结构体的类型对象
	typ := value.Type()

	// 遍历结构体的所有字段
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)        // 获取第 i 个字段的反射类型对象
		fieldValue := value.Field(i) // 获取第 i 个字段的反射值对象

		fmt.Printf("Field Name: %s, Field Value: %v\n", field.Name, fieldValue.Interface())
	}
}

func TestLoadConfFromCmd(t *testing.T) {
	p := Person{
		//Name: "Alice",
		//Age:  12,
	}
	LoadConfFromCmd(&p)
}

// 定义结构体
type Config2 struct {
	UserName string `yaml:"user_name,omitempty"`
	Age      int    `yaml:"age,omitempty"` // 使用指针类型
}

func TestClone(t *testing.T) {
	person := &Person{Name: "John", Age: 30}

	newPerson := structtool.NewInstance(person)

	fmt.Printf("Original Person: %+v\n", person)
	fmt.Printf("New Person: %+v\n", newPerson)

}
