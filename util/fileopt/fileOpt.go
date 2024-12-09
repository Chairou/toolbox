package fileopt

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// 检查文件名是否具有有效的后缀，支持通配符
func hasValidSuffix(filename string, suffixes []string) bool {
	for _, suffix := range suffixes {
		// 处理通配符
		if strings.HasPrefix(suffix, "*") {
			if strings.HasSuffix(filename, suffix[1:]) {
				return true
			}
		} else {
			if strings.HasSuffix(filename, suffix) {
				return true
			}
		}
	}
	return false
}

func GetAllFilesFromDirectory(rootDir string, filterSuffixes []string) (files []string, err error) {
	// 检查目录是否存在
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		fmt.Printf("目录 %s 不存在\n", rootDir)
	} else {
		fmt.Printf("目录 %s 存在\n", rootDir)
	}
	files = make([]string, 0)
	// 使用 filepath.WalkDir 遍历目录
	err = filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// 检查是否是文件并打印文件名
		if !d.IsDir() && hasValidSuffix(d.Name(), filterSuffixes) {
			files = append(files, path)
			fmt.Println(path) // 打印完整路径
		}
		return nil
	})

	if err != nil {
		log.Fatalf("无法读取目录: %v", err)
		return nil, err
	}
	return files, nil
}
