package conf

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// cmdFlagName 命令行参数名与 yaml 键保持一致，没有 yaml 标签时使用字段名
func cmdFlagName(field reflect.StructField) string {
	if name := strings.Split(field.Tag.Get("yaml"), ",")[0]; name != "" && name != "-" {
		return name
	}
	return field.Name
}

// cmdFlagUsage 命令行参数说明取自 usage 标签，没有时退回为参数名
func cmdFlagUsage(field reflect.StructField, name string) string {
	if usage := field.Tag.Get("usage"); usage != "" {
		return usage
	}
	return name
}

// LoadConfFromCmd 使用反射从命令行参数加载配置到结构体
// 参数名为 yaml 键（如 -redis_host），说明来自 usage 标签
func LoadConfFromCmd[T any](config T) error {
	if len(os.Args) < 2 {
		return nil
	}
	// 使用反射获取结构体的字段
	val := reflect.ValueOf(config)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		flagName := cmdFlagName(field)
		usage := cmdFlagUsage(field, flagName)
		fieldType := field.Type.Kind()
		fieldVal := val.Field(i)

		// 创建命令行标志
		switch fieldType {
		case reflect.String:
			if fieldVal.String() != "" {
				continue
			}
			flag.StringVar(val.Field(i).Addr().Interface().(*string), flagName, "", usage)
		case reflect.Int:
			if fieldVal.Int() != 0 {
				continue
			}
			flag.IntVar(val.Field(i).Addr().Interface().(*int), flagName, 0, usage)
		case reflect.Float64:
			if fieldVal.Float() != 0 {
				continue
			}
			flag.Float64Var(val.Field(i).Addr().Interface().(*float64), flagName, 0, usage)
		default:
			return fmt.Errorf("unsupported field type: %s for field: %s", fieldType, fieldName)
		}
	}

	// 解析命令行参数
	if !flag.Parsed() {
		flag.Parse()
	}

	return nil

}
