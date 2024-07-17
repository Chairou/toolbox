package conf

import (
	"fmt"
	"github.com/jinzhu/copier"
	"sync"
)

func initLock() {
	confLock = new(sync.RWMutex)
}

func init() {
	initLock()
	LoadAllConf()
}

func mergeConfig(cmd, file, env *Config) {
	_ = copier.CopyWithOption(Conf, env, copier.Option{IgnoreEmpty: true})
	_ = copier.CopyWithOption(Conf, file, copier.Option{IgnoreEmpty: true})
	_ = copier.CopyWithOption(Conf, cmd, copier.Option{IgnoreEmpty: true})
	fmt.Printf("FINIAL: %#v\n", Conf)
}

// LoadAllConf 优先级，命令行，配置文件，环境变量
func LoadAllConf() {
	once.Do(func() {
		//读取命令行参数
		cmd, err := loadConfFromCmd()
		if err != nil {
			fmt.Println("cmd err: ", err)
		}
		//fmt.Printf("cmd: %#v\n", cmd)
		//读取配置文件
		file, err := loadConfFromFile("dev.yaml")
		if err != nil {
			fmt.Println("confFile err: ", err)
		}
		//fmt.Printf("file: %#v\n", file)

		//读取环境变量
		env, err := loadConfFromEnv()
		if err != nil {
			fmt.Println("env err: ", err)
		}
		//fmt.Printf("env: %#v\n", env)
		//合并配置
		mergeConfig(cmd, file, env)

	})
}

func GetConf() *Config {
	confLock.RLock()
	defer confLock.RUnlock()
	return Conf
}
