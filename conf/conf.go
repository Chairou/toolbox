package conf

import (
	"fmt"
	"log"
	"os"

	"github.com/Chairou/toolbox/util/structtool"
	"github.com/jinzhu/copier"
)

func mergeConfig[T any](conf, cmd, env, file *T) {
	_ = copier.CopyWithOption(conf, file, copier.Option{IgnoreEmpty: true})
	_ = copier.CopyWithOption(conf, env, copier.Option{IgnoreEmpty: true})
	_ = copier.CopyWithOption(conf, cmd, copier.Option{IgnoreEmpty: true})
	fmt.Printf("FINIAL: %#v\n", conf)
}

// LoadAllConf 优先级，命令行，环境变量，配置文件
func LoadAllConf[T any](conf *T) {
	once.Do(func() {
		cmdConfig := structtool.NewEmptyInstance(conf)
		fileConfig := structtool.NewEmptyInstance(conf)
		envConfig := structtool.NewEmptyInstance(conf)
		//读取命令行参数
		err := LoadConfFromCmd(cmdConfig)
		if err != nil {
			log.Println("cmdConfig err: ", err)
		}
		fmt.Printf("cmdConfig: %+v\n", cmdConfig)
		//读取配置文件
		envStr := os.Getenv("env")
		if envStr == "" {
			log.Println("env is empty , set to dev.yaml")
			envStr = "dev"
		} else {
			log.Println("set conf to ", envStr+".yaml")
		}
		err = loadConfFromFile(envStr+".yaml", fileConfig)
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
		mergeConfig(conf, cmdConfig, envConfig, fileConfig)
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
