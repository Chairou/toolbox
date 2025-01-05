package conf

import (
	"gopkg.in/yaml.v2"
	"io"
	"os"
)

func loadConfFromFile[T any](fileName string, config T) error {
	fd, err := os.OpenFile(fileName, os.O_RDONLY, 666)
	if err != nil {
		return err
	}
	dataBytes, err := io.ReadAll(fd)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(dataBytes, config)
	if err != nil {
		return err
	}
	return nil
}
