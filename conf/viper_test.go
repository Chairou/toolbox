package conf

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"testing"
)

func TestViper(t *testing.T) {

	pflag.Int("redis.port", 3302, "redis port")
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return
	}
	pflag.Parse()

	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()

	err = viper.ReadInConfig() //根据上面配置加载文件
	if err != nil {
		fmt.Println(err)
		return
	}
}
