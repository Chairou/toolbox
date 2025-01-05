package conf

import (
	"sync"
)

var once sync.Once

type Config struct {
	Env          string `yaml:"env" json:"env" env:"env"`
	Version      int    `yaml:"version" json:"version" env:"version"`
	RedisName    string `yaml:"redis_name" json:"redis_name" env:"redis_name"`
	RedisHost    string `yaml:"redis_host" json:"redis_host" env:"redis_host"`
	RedisAuth    string `yaml:"redis_auth" json:"redis_auth" env:"redis_auth"`
	MysqlName    string `yaml:"mysql_name" json:"mysql_name" env:"mysql_name"`
	MysqlHost    string `yaml:"mysql_host" json:"mysql_host" env:"mysql_host"`
	MysqlUser    string `yaml:"mysql_user" json:"mysql_user" env:"mysql_user"`
	MysqlPass    string `yaml:"mysql_pass" json:"mysql_pass" env:"mysql_pass"`
	MysqlDb      string `yaml:"mysql_db" json:"mysql_db" env:"mysql_db"`
	MysqlCharSet string `yaml:"mysql_charset" json:"mysql_charset" env:"mysql_charset"`
	LogFileName  string `yaml:"log_file_name" json:"log_file_name" env:"log_file_name"`
}
