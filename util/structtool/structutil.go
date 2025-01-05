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

func PrintStructAll(item any) {
	v := reflect.ValueOf(item)
	PrintStruct(v, "")
}
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
