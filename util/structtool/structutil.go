// Package structtool 提供结构体相关的工具函数，包括创建空实例、获取结构体名称、打印结构体字段等
package structtool

import (
	"fmt"
	"reflect"
)

// NewEmptyInstance 定义一个泛型函数，接受一个类型为T的指针，返回一个新的该类型的零值实例指针
func NewEmptyInstance[T any](item *T) *T {
	itemType := reflect.TypeOf(item).Elem()
	itemValue := reflect.New(itemType).Interface()
	return itemValue.(*T)
}

// GetStructName 获取结构体名字，支持传入结构体值或指针
func GetStructName(item any) string {
	if item == nil {
		return ""
	}
	t := reflect.TypeOf(item)
	// 如果是指针类型，解引用获取实际类型
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// PrintStructAll 打印结构体所有字段，支持传入结构体值或指针
func PrintStructAll(item any) {
	if item == nil {
		return
	}
	v := reflect.ValueOf(item)
	// 如果是指针类型，解引用获取实际值
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	PrintStruct2(v, "")
	fmt.Println()
}

// PrintStruct 递归打印结构体字段（仅处理 struct 类型）
//
// Deprecated: 推荐使用 PrintStruct2，支持 struct/slice/map 的递归打印
func PrintStruct(v reflect.Value, indent string) {
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			fieldType := v.Type().Field(i)

			// 跳过未导出字段，避免调用 Interface() 时 panic
			if !fieldType.IsExported() {
				continue
			}

			if fieldValue.Kind() == reflect.Struct {
				fmt.Printf("%sName: %s, Type: %s",
					indent, fieldType.Name, fieldType.Type)
				PrintStruct(fieldValue, indent+"---")
			} else {
				fmt.Printf("\n%sName: %s, Type: %s, Value: %v",
					indent, fieldType.Name, fieldType.Type, fieldValue.Interface())
			}
		}
	}
}

// PrintStruct2 递归打印结构体字段，支持 struct、slice、map 类型的嵌套打印
func PrintStruct2(v reflect.Value, indent string) {
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			fieldType := v.Type().Field(i)

			// 跳过未导出字段，避免调用 Interface() 时 panic
			if !fieldType.IsExported() {
				continue
			}

			if fieldValue.Kind() == reflect.Struct || fieldValue.Kind() == reflect.Slice || fieldValue.Kind() == reflect.Map {
				fmt.Printf("\n%sName: %s, Type: %s",
					indent, fieldType.Name, fieldType.Type)
				PrintStruct2(fieldValue, indent+"---")
			} else {
				fmt.Printf("\n%sName: %s, Type: %s, Value: %v",
					indent, fieldType.Name, fieldType.Type, fieldValue.Interface())
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			fmt.Printf("\n%sIndex: %d, Type: %s, Value: %v",
				indent, i, v.Type().Elem(), elem.Interface())
			if elem.Kind() == reflect.Struct || elem.Kind() == reflect.Slice || elem.Kind() == reflect.Map {
				PrintStruct2(elem, indent+"---")
			}
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			value := v.MapIndex(key)
			fmt.Printf("\n%sKey: %v, Type: %s, Value: %v",
				indent, key.Interface(), v.Type().Elem(), value.Interface())
			if value.Kind() == reflect.Struct || value.Kind() == reflect.Slice || value.Kind() == reflect.Map {
				PrintStruct2(value, indent+"---")
			}
		}
	default:
		// 非 struct/slice/map 类型不做处理
	}
}

// StructToMap 将结构体转换为 map[string]any，仅处理导出字段
func StructToMap(item any) (map[string]any, error) {
	if item == nil {
		return nil, fmt.Errorf("input is nil")
	}
	v := reflect.ValueOf(item)
	// 如果是指针类型，解引用获取实际值
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, fmt.Errorf("input is nil pointer")
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input is not a struct, got %s", v.Kind())
	}

	result := make(map[string]any)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		result[field.Name] = v.Field(i).Interface()
	}
	return result, nil
}