package conf

import (
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"os"
	"sync"
)

var ConfMap sync.Map

type Config struct {
	Env     string      `yaml:"env" json:"env"`
	Version string      `yaml:"version" json:"version"`
	Redis   []RedisStru `yaml:"redis" json:"redis"`
	MySQL   []MySQLStru `yaml:"mysql" json:"mysql"`
}

type RedisStru struct {
	Name  string `yaml:"name" json:"name"`
	Host  string `yaml:"host" json:"host"`
	Auth  string `yaml:"auth" json:"auth"`
	Owner string `yaml:"owner" json:"owner"`
}

type MySQLStru struct {
	Name     string `yaml:"name" json:"name"`
	Host     string `yaml:"host" json:"host"`
	Username string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	Charset  string `yaml:"charset" json:"charset"`
	Owner    string `yaml:"owner" json:"owner"`
}

func LoadConfig() *Config {
	env := GetEnvironment("env")
	if env == "" {
		// 默认读取dev
		env = "dev"
	}

	inst, ok := ConfMap.Load(env)
	if ok {
		return inst.(*Config)
	}

	var config Config
	if env != "dev" && env != "release" {
		log.Fatalln("env only can be set to dev or release")
	}
	fileName := env + ".yaml"
	fd, err := os.OpenFile(fileName, os.O_RDONLY, 666)
	if err != nil {
		log.Fatalln("OpenFile error: ", err)
	}
	dataBytes, err := io.ReadAll(fd)
	if err != nil {
		log.Fatalln("ReadAll error: ", err)
	}

	err = yaml.Unmarshal(dataBytes, &config)
	if err != nil {
		log.Fatalln("yaml.Unmarshal error: ", err)
	}

	ConfMap.Store(env, &config)
	return &config
}
