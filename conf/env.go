package conf

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
)

func loadConfFromEnv[T any](config T) error {
	fmt.Printf("config type: %+v\n", config)
	val := reflect.ValueOf(config).Elem()
	t := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := t.Field(i)
		envValue := os.Getenv(field.Tag.Get("env"))
		//fmt.Println("##########", field.Tag.Get("env"))
		//fmt.Println("##########", envValue)
		if envValue == "" {
			continue
		}

		fieldValue := val.Field(i)
		if !fieldValue.CanSet() {
			return fmt.Errorf("cannot set field %s", field.Name)
		}

		switch fieldValue.Kind() {
		case reflect.String:
			if fieldValue.String() != "" {
				continue
			}
			fieldValue.SetString(envValue)
		case reflect.Int:
			if fieldValue.Int() != 0 {
				continue
			}
			intVal, err := strconv.Atoi(envValue)
			if err != nil {
				return fmt.Errorf("invalid int value for field %s: %s", field.Name, envValue)
			}
			fieldValue.SetInt(int64(intVal))
		default:
			return fmt.Errorf("unsupported kind: %s", fieldValue.Kind())
		}
	}
	return nil
}
