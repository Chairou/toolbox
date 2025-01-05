package test

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
)

// LoadConfFromCmd 使用反射从命令行参数加载配置到结构体
func LoadConfFromCmd[T any](config *T) error {
	if len(os.Args) < 2 {
		return errors.New("no config file specified")
	}

	val := reflect.ValueOf(config).Elem()
	t := val.Type()

	// 创建一个临时的map来存储flag
	flagMap := make(map[string]interface{})

	for i := 0; i < val.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		fieldType := field.Type.Kind()

		var defaultValue interface{}
		var usage string = "usage"

		switch fieldType {
		case reflect.String:
			defaultValue = ""
		case reflect.Int:
			defaultValue = 0
		// 可以根据需要添加更多类型
		default:
			continue // 忽略不支持的类型
		}

		flagVar := reflect.New(field.Type).Interface()
		flagName := fmt.Sprintf("--%s", fieldName)
		switch fieldType {
		case reflect.String:
			flag.StringVar(flagVar.(*string), flagName, defaultValue.(string), usage)
		case reflect.Int:
			flag.IntVar(flagVar.(*int), flagName, defaultValue.(int), usage)
			// 添加更多类型的flag设置
		}

		flagMap[flagName] = flagVar
	}

	// 解析命令行参数
	flag.Parse()

	// 将flag的值设置回结构体
	for i := 0; i < val.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		flagName := fmt.Sprintf("--%s", fieldName)
		if valField := flagMap[flagName]; valField != nil {
			val.Field(i).Set(reflect.ValueOf(valField).Elem())
		}
	}

	return nil
}
