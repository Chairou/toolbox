package conf

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"toolbox/logger"
)

// GetEnvironment get the system environment by name.
// GetEnvironment 获取系统环境变量
// parameters envName
// return envValue
func GetEnvironment(envName string) (envValue string) {
	envValue = os.Getenv(envName)
	if envValue == "" {
		log.Print("Getenv null:", envName)
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

// GetLocalIPAddr get the local IP address
// GetLocalIPAddr 获取本地IP
// return IP address
func GetLocalIPAddr() string {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addresses {
		ipaddr, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			continue
		}
		if ipaddr.IsLoopback() {
			continue
		}
		if ipaddr.To4() != nil {
			if runtime.GOOS == "darwin" {
				if !strings.HasPrefix(ipaddr.String(), "192") {
					continue
				}
			}
			return ipaddr.String()
		}
	}
	return ""
}

func SetEnv(key string, value string) error {
	err := os.Setenv(key, value)
	if err != nil {
		return err
	}
	return nil
}

func UnSetEnv(key string) error {
	err := os.Unsetenv(key)
	if err != nil {
		return err
	}
	return nil
}
