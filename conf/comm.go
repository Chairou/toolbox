package conf

import (
	"sync"
)

var once sync.Once

type Config struct {
	Env          string `yaml:"env" json:"env" env:"env" usage:"运行环境，如 dev、test"`
	Version      int    `yaml:"version" json:"version" env:"version" usage:"配置版本号"`
	RedisName    string `yaml:"redis_name" json:"redis_name" env:"redis_name" usage:"Redis 实例名称"`
	RedisHost    string `yaml:"redis_host" json:"redis_host" env:"redis_host" usage:"Redis 地址，如 127.0.0.1:6379"`
	RedisAuth    string `yaml:"redis_auth" json:"redis_auth" env:"redis_auth" usage:"Redis 密码"`
	MysqlName    string `yaml:"mysql_name" json:"mysql_name" env:"mysql_name" usage:"MySQL 实例名称"`
	MysqlHost    string `yaml:"mysql_host" json:"mysql_host" env:"mysql_host" usage:"MySQL 地址，如 127.0.0.1:3306"`
	MysqlUser    string `yaml:"mysql_user" json:"mysql_user" env:"mysql_user" usage:"MySQL 用户名"`
	MysqlPass    string `yaml:"mysql_pass" json:"mysql_pass" env:"mysql_pass" usage:"MySQL 密码"`
	MysqlDb      string `yaml:"mysql_db" json:"mysql_db" env:"mysql_db" usage:"MySQL 数据库名"`
	MysqlCharSet string `yaml:"mysql_charset" json:"mysql_charset" env:"mysql_charset" usage:"MySQL 字符集，如 utf8mb4"`
	LogFileName  string `yaml:"log_file_name" json:"log_file_name" env:"log_file_name" usage:"日志文件名"`
}
