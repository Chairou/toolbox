package conf

import (
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"os"
)

func loadConfFromFile(fileName string) (*Config, error) {
	config := &Config{}
	fd, err := os.OpenFile(fileName, os.O_RDONLY, 666)
	if err != nil {
		log.Fatalln("OpenFile error: ", err)
	}
	dataBytes, err := io.ReadAll(fd)
	if err != nil {
		log.Fatalln("ReadAll error: ", err)
	}
	err = yaml.Unmarshal(dataBytes, config)
	if err != nil {
		log.Fatalln("yaml.Unmarshal error: ", err)
	}

	return config, nil
}
