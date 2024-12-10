package fileopt

import (
	"fmt"
	"io"
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

// ReadFileAllByte 读取文件所有内容，返回byte
func ReadFileAllByte(filename string) (retByte []byte, err error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("文件 %s 不存在\n", filename)
	} else {
		fmt.Printf("文件 %s 存在\n", filename)
	}
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("打开文件 %s 失败: %v\n", filename, err)
		return nil, err
	}
	defer file.Close()
	contentByte, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("读取文件 %s 失败: %v\n", filename, err)
		return nil, err
	}
	return contentByte, nil
}

func ReadFileAllString(filename string) (string, error) {
	retByte, err := ReadFileAllByte(filename)
	return string(retByte), err
}
