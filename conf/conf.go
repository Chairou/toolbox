package conf

import (
	"fmt"
	"github.com/Chairou/toolbox/util/structtool"
	"github.com/jinzhu/copier"
)

//func initLock() {
//	confLock = new(sync.RWMutex)
//}

//func init() {
//	initLock()
//	LoadAllConf()
//}

func mergeConfig[T any](cmd, file, env *T) {
	_ = copier.CopyWithOption(Conf, env, copier.Option{IgnoreEmpty: true})
	_ = copier.CopyWithOption(Conf, file, copier.Option{IgnoreEmpty: true})
	_ = copier.CopyWithOption(Conf, cmd, copier.Option{IgnoreEmpty: true})
	fmt.Printf("FINIAL: %#v\n", Conf)
}

// LoadAllConf 优先级，命令行，配置文件，环境变量
func LoadAllConf[T any](conf *T) {
	cmdConfig := structtool.NewInstance(conf)
	fileConfig := structtool.NewInstance(conf)
	envConfig := structtool.NewInstance(conf)
	once.Do(func() {
		//读取命令行参数
		err := LoadConfFromCmd(cmdConfig)
		if err != nil {
			fmt.Println("cmdConfig err: ", err)
		}
		fmt.Printf("cmdConfig: %+v\n", cmdConfig)
		//读取配置文件
		err = loadConfFromFile("dev.yaml", fileConfig)
		if err != nil {
			fmt.Println("confFile err: ", err)
		}
		fmt.Printf("fileConfig: %+v\n", fileConfig)

		// 读取环境变量
		err = loadConfFromEnv(envConfig)
		if err != nil {
			fmt.Println("envConfig err: ", err)
		}
		fmt.Printf("env: %+v\n", envConfig)
		//合并配置
		mergeConfig(cmdConfig, fileConfig, envConfig)

	})
}

func GetConf() *Config {
	confLock.RLock()
	defer confLock.RUnlock()
	return Conf
}

func createStruct[T any]() *T {
	return new(T)
}

func LoadConf[T any](config *T) error {
	// 先处理YAML文件， 保底处理方式
	// 然后处理e

	return nil
}
