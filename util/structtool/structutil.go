package structtool

import (
	"fmt"
	"reflect"
)

// NewInstance 定义一个泛型函数，接受一个类型为T的值，返回一个新的该类型的实例
func NewInstance[T any](item *T) *T {
	itemType := reflect.TypeOf(item).Elem()
	itemValue := reflect.New(itemType).Interface()
	return itemValue.(*T)
}

// GetStructName 获取结构体名字
func GetStructName(item any) string {
	return reflect.TypeOf(item).Name()
}

// PrintStructAll 打印结构体所有字段
func PrintStructAll(item any) {
	v := reflect.ValueOf(item)
	PrintStruct2(v, "")
}

// PrintStruct 打印结构体，递归用
func PrintStruct(v reflect.Value, indent string) {
	// 确保我们有一个结构体
	if v.Kind() == reflect.Struct {
		// 遍历结构体的字段
		for i := 0; i < v.NumField(); i++ {
			// 获取字段的反射值对象
			fieldValue := v.Field(i)
			// 获取字段的类型对象
			fieldType := v.Type().Field(i)

			// 如果字段是结构体类型，则递归调用
			if fieldValue.Kind() == reflect.Struct {
				fmt.Printf("%sName: %s, Type: %s",
					indent, fieldType.Name, fieldType.Type)
				PrintStruct(fieldValue, indent+"---") // 递归，增加缩进
			} else {
				// 打印字段名、类型和值
				fmt.Printf("\n%sName: %s, Type: %s, Value: %v",
					indent, fieldType.Name, fieldType.Type, fieldValue.Interface())
			}
		}
	}
}

// PrintStruct2 打印结构体，递归用
func PrintStruct2(v reflect.Value, indent string) {
	// 确保我们有一个结构体或slice/map
	switch v.Kind() {
	case reflect.Struct:
		// 遍历结构体的字段
		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			fieldType := v.Type().Field(i)

			if fieldValue.Kind() == reflect.Struct || fieldValue.Kind() == reflect.Slice || fieldValue.Kind() == reflect.Map {
				fmt.Printf("\n%sName: %s, Type: %s",
					indent, fieldType.Name, fieldType.Type)
				PrintStruct2(fieldValue, indent+"---") // 递归，增加缩进
			} else {
				fmt.Printf("\n%sName: %s, Type: %s, Value: %v",
					indent, fieldType.Name, fieldType.Type, fieldValue.Interface())
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			fmt.Printf("\n%sIndex: %d, Type: %s, Value: %v",
				indent, i, v.Type().Elem(), v.Index(i).Interface())
			if v.Index(i).Kind() == reflect.Struct || v.Index(i).Kind() == reflect.Slice || v.Index(i).Kind() == reflect.Map {
				PrintStruct2(v.Index(i), indent+"---") // 递归，增加缩进
			}
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			value := v.MapIndex(key)
			fmt.Printf("\n%sKey: %v, Type: %s, Value: %v",
				indent, key.Interface(), v.Type().Elem(), value.Interface())
			if value.Kind() == reflect.Struct || value.Kind() == reflect.Slice || value.Kind() == reflect.Map {
				PrintStruct2(value, indent+"---") // 递归，增加缩进
			}
		}
	}

}
