package conf

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
)

func loadConfFromEnv() (*Config, error) {
	config := &Config{}
	val := reflect.ValueOf(config).Elem()
	t := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := t.Field(i)
		//fmt.Println("##########", field.Name)
		envValue := os.Getenv(field.Tag.Get("env"))
		fmt.Println("##########", field.Tag.Get("env"))
		fmt.Println("##########", envValue)
		if envValue == "" {
			continue
		}

		fieldValue := val.Field(i)
		if !fieldValue.CanSet() {
			return nil, fmt.Errorf("cannot set field %s", field.Name)
		}

		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(envValue)
		case reflect.Int:
			intVal, err := strconv.Atoi(envValue)
			if err != nil {
				return nil, fmt.Errorf("invalid int value for field %s: %s", field.Name, envValue)
			}
			fieldValue.SetInt(int64(intVal))
		default:
			return nil, fmt.Errorf("unsupported kind: %s", fieldValue.Kind())
		}
	}
	return config, nil
}
