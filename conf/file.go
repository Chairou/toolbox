package conf

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

func loadConfFromFile[T any](fileName string, config T) error {
	fd, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer func(fd *os.File) {
		err := fd.Close()
		if err != nil {
			panic("close file err:" + err.Error())
		}
	}(fd)
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
