package logger

import (
	"testing"

	"github.com/Chairou/toolbox/util/conv"
	"github.com/jinzhu/copier"
)

var MainConf Config

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
	PrivateKey   string `yaml:"private_key" json:"private_key" env:"private_key"`
	KeyPasswd    string `yaml:"key_password" json:"key_password" env:"key_password"`
	Uri          string `yaml:"uri" json:"uri" env:"uri"`
	PicPath      string `yaml:"pic_path" json:"pic_path" env:"pic_path"`
	FileName     string `config:"fileName"`     // 日志文件名或路径
	Level        int    `config:"level"`        // 日志级别，可选 DEBUG_LEVEL、INFO_LEVEL、ERROR_LEVEL
	MaxSizeMB    int    `config:"MaxSizeMB"`    // 单个日志文件最大大小（MB）
	MaxBackups   int    `config:"MaxBackups"`   // 最大保留的旧日志文件数量
	MaxAgeDay    int    `config:"MaxAgeDay"`    // 旧日志文件最大保留天数
	Compress     bool   `config:"Compress"`     // 是否压缩旧日志文件，默认不压缩
	PrintConsole bool   `config:"PrintConsole"` // 是否同时输出到控制台

}

// 测试自动匹配配置到logOpt中
func TestLogFromConfig(t *testing.T) {
	logOpt := LogOpt{}
	MainConf.MaxSizeMB = 100
	MainConf.MaxAgeDay = 31
	MainConf.FileName = "test.log"
	mainConfMap, err := conv.StructToMap(MainConf, "config")
	if err != nil {
		t.Error(err)
	}
	err = trans(&logOpt, MainConf)
	if err != nil {
		t.Error(err)
	}
	t.Log(mainConfMap)
	t.Log(logOpt)
}

func trans(pool *LogOpt, config any) error {
	err := copier.Copy(pool, config)
	if err != nil {
		return err
	}
	return nil
}

func TestGetLogV2(t *testing.T) {
	opt := LogOpt{}
	opt.MaxSizeMB = 100
	opt.MaxBackups = 3
	opt.MaxAgeDay = 7
	opt.Compress = true
	opt.PrintConsole = false
	opt.FileName = "./log/test.log"
	logInst, err := NewLogOpt("test", &opt)
	if err != nil {
		t.Error(err)
	}
	testInst := GetLogV2()
	if testInst == nil {
		t.Error("GetLogV2 return nil")
	}
	logInst.Info("test info message")
	t.Log("GetLogV2 success")
}

func TestInitLog2(t *testing.T) {
	opt := LogOpt{}
	opt.MaxSizeMB = 100
	opt.MaxBackups = 3
	opt.MaxAgeDay = 7
	opt.Compress = true
	opt.PrintConsole = false
	opt.FileName = "./log/test.log"
	logInst, err := NewLogOpt("test", &opt)
	if err != nil {
		t.Log("InitLog| init logger failed", err)
		return
	}
	t.Log("InitLog| logger initialized successfully")
	logInst.Info("test info message")
}
