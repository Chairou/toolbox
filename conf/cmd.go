package conf

import (
	"flag"
	"os"
	"reflect"
)

func loadConfFromCmd() (*Config, error) {
	config := &Config{}
	if len(os.Args) < 2 {
		return nil, nil
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
		}
	}

	// 解析命令行参数
	flag.Parse()

	// 输出结果
	return config, nil
}
