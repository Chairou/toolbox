package conf

import (
	"fmt"
	"os"
	"toolbox/logger"
)

// GetEnvironment get the system environment by name.
// GetEnvironment 获取系统环境变量
// parameters envName
// return envValue
func GetEnvironment(envName string) (envValue string) {
	log, err := logger.NewLogPool("internal.log")
	if err != nil {
		fmt.Println("GetEnvironment|NewLogPool err:", err)
	}
	envValue = os.Getenv(envName)
	if envValue == "" {
		log.Errorln("Getenv null:", envName)
		return ""
	}
	return envValue
}

// GetPid get the proccess id
// GetPid 获取进程ID
// return pid
func GetPid() (pid int) {
	getWd := os.Getppid()
	return getWd
}

// Getwd get the current working directory
// Getwd 获取当前工作目录
// return getwd
func Getwd() (getwd string, err error) {
	log, err := logger.NewLogPool("internal.log")
	if err != nil {
		fmt.Println("GetEnvironment|NewLogPool err:", err)
	}
	getwd, err = os.Getwd()
	if err != nil {
		log.Errorln("Getwd|Getwd err:", err)
		return "", nil
	}
	return getwd, nil
}
