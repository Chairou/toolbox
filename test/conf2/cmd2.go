package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"
)

// LoadConfFromCmd 使用反射从命令行参数加载配置到结构体
func LoadConfFromCmd[T any](config *T) error {

	if len(os.Args) < 2 {
		return nil
	}
	// 使用反射获取结构体的字段
	val := reflect.ValueOf(config).Elem()
	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		fieldType := field.Type.Kind()

		// 创建命令行标志
		switch fieldType {
		case reflect.String:
			flag.StringVar(val.Field(i).Addr().Interface().(*string), fieldName, "", "usage")
		case reflect.Int:
			flag.IntVar(val.Field(i).Addr().Interface().(*int), fieldName, 0, "usage")
		case reflect.Float64:
			flag.Float64Var(val.Field(i).Addr().Interface().(*float64), fieldName, 0, "usage")
		default:
			panic("not string or int ")
		}
	}

	// 解析命令行参数
	flag.Parse()

	// 输出结果
	return nil

}

func main() {
	type Config struct {
		Name  string  `json:"Name"`
		Age   int     `json:"Age"`
		Money float64 `json:"Money"`
	}

	config := &Config{}
	if err := LoadConfFromCmd(config); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Config:", config)
}
