package conf

import (
	"fmt"
	"github.com/Chairou/toolbox/util/structtool"
	"github.com/jinzhu/copier"
	"log"
)

func mergeConfig[T any](conf, cmd, file, env *T) {
	_ = copier.CopyWithOption(conf, file, copier.Option{IgnoreEmpty: true})
	_ = copier.CopyWithOption(conf, env, copier.Option{IgnoreEmpty: true})
	_ = copier.CopyWithOption(conf, cmd, copier.Option{IgnoreEmpty: true})
	fmt.Printf("FINIAL: %#v\n", conf)
}

// LoadAllConf 优先级，命令行，配置文件，环境变量
func LoadAllConf[T any](conf *T) {
	cmdConfig := structtool.NewEmptyInstance(conf)
	fileConfig := structtool.NewEmptyInstance(conf)
	envConfig := structtool.NewEmptyInstance(conf)
	once.Do(func() {
		//读取命令行参数
		err := LoadConfFromCmd(cmdConfig)
		if err != nil {
			log.Println("cmdConfig err: ", err)
		}
		fmt.Printf("cmdConfig: %+v\n", cmdConfig)
		//读取配置文件
		err = loadConfFromFile("dev.yaml", fileConfig)
		if err != nil {
			log.Println("confFile err: ", err)
		}
		fmt.Printf("fileConfig: %+v\n", fileConfig)

		// 读取环境变量
		err = loadConfFromEnv(envConfig)
		if err != nil {
			log.Println("envConfig err: ", err)
		}
		fmt.Printf("env: %+v\n", envConfig)
		//合并配置
		mergeConfig(conf, cmdConfig, fileConfig, envConfig)

	})
}

func createStruct[T any]() *T {
	return new(T)
}

func LoadConf[T any](config *T) error {
	// 先处理YAML文件， 保底处理方式
	// 然后处理e

	return nil
}
