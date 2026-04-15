// Package fileopt 提供文件和目录操作的工具函数
package fileopt

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

// hasValidSuffix 检查文件名是否具有有效的后缀，支持通配符"*"匹配所有文件
func hasValidSuffix(filename string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if suffix == "*" {
			return true
		}
		// 处理带通配符前缀的后缀，如 "*.txt"
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

// GetAllFilesFromDirectory 递归遍历指定目录，返回匹配后缀过滤条件的所有文件路径
// filterSuffixes 支持通配符"*"匹配所有文件，也支持 "*.txt" 或 ".txt" 格式
func GetAllFilesFromDirectory(rootDir string, filterSuffixes []string) ([]string, error) {
	if _, err := os.Stat(rootDir); err != nil {
		return nil, err
	}

	files := make([]string, 0, 1024)
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && hasValidSuffix(d.Name(), filterSuffixes) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// ReadFileAllByte 读取文件所有内容，返回字节切片
func ReadFileAllByte(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	contentByte, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return contentByte, nil
}

// ReadFileAllString 读取文件所有内容，返回字符串
func ReadFileAllString(filename string) (string, error) {
	retByte, err := ReadFileAllByte(filename)
	return string(retByte), err
}
