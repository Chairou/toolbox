package fileopt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetAllFilesFromDirectory_Wildcard(t *testing.T) {
	files, err := GetAllFilesFromDirectory("./testtmp", []string{"*"})
	if err != nil {
		t.Fatal("GetAllFilesFromDirectory err:", err)
	}
	if len(files) == 0 {
		t.Error("expected at least one file, got 0")
	}
	// testtmp 目录下应包含 a.txt
	found := false
	for _, f := range files {
		if filepath.Base(f) == "a.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find a.txt in results, got %v", files)
	}
}

func TestGetAllFilesFromDirectory_SingleSuffix(t *testing.T) {
	files, err := GetAllFilesFromDirectory("./", []string{".txt"})
	if err != nil {
		t.Fatal("GetAllFilesFromDirectory err:", err)
	}
	// 应至少包含 testtmp/a.txt
	found := false
	for _, f := range files {
		if filepath.Base(f) == "a.txt" {
			found = true
		}
		// 所有文件都应以 .txt 结尾
		if !hasSuffix(f, ".txt") {
			t.Errorf("file %s does not have .txt suffix", f)
		}
	}
	if !found {
		t.Errorf("expected to find a.txt in results, got %v", files)
	}
}

func TestGetAllFilesFromDirectory_MultiSuffix(t *testing.T) {
	files, err := GetAllFilesFromDirectory("./", []string{".txt", ".go"})
	if err != nil {
		t.Fatal("GetAllFilesFromDirectory err:", err)
	}
	hasTxt := false
	hasGo := false
	for _, f := range files {
		if hasSuffix(f, ".txt") {
			hasTxt = true
		}
		if hasSuffix(f, ".go") {
			hasGo = true
		}
	}
	if !hasTxt {
		t.Error("expected to find .txt files")
	}
	if !hasGo {
		t.Error("expected to find .go files")
	}
}

func TestGetAllFilesFromDirectory_WildcardSuffix(t *testing.T) {
	files, err := GetAllFilesFromDirectory("./testtmp", []string{"*.txt"})
	if err != nil {
		t.Fatal("GetAllFilesFromDirectory err:", err)
	}
	if len(files) == 0 {
		t.Error("expected at least one .txt file")
	}
	for _, f := range files {
		if !hasSuffix(f, ".txt") {
			t.Errorf("file %s does not match *.txt pattern", f)
		}
	}
}

func TestGetAllFilesFromDirectory_NoMatch(t *testing.T) {
	files, err := GetAllFilesFromDirectory("./testtmp", []string{".xyz"})
	if err != nil {
		t.Fatal("GetAllFilesFromDirectory err:", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files for .xyz suffix, got %d: %v", len(files), files)
	}
}

func TestGetAllFilesFromDirectory_NotExistDir(t *testing.T) {
	_, err := GetAllFilesFromDirectory("./not_exist_dir_12345", []string{"*"})
	if err == nil {
		t.Error("expected error for non-existent directory, got nil")
	}
}

func TestGetAllFilesFromDirectory_EmptySuffixes(t *testing.T) {
	files, err := GetAllFilesFromDirectory("./testtmp", []string{})
	if err != nil {
		t.Fatal("GetAllFilesFromDirectory err:", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files for empty suffixes, got %d: %v", len(files), files)
	}
}

func TestReadFileAllByte(t *testing.T) {
	content, err := ReadFileAllByte("./testtmp/a.txt")
	if err != nil {
		t.Fatal("ReadFileAllByte err:", err)
	}
	if len(content) == 0 {
		t.Error("expected non-empty content from a.txt")
	}
	// a.txt 内容为 "kkk"（可能带换行）
	expected := "kkk"
	got := string(content)
	if got != expected && got != expected+"\n" {
		t.Errorf("ReadFileAllByte result = %q, want %q", got, expected)
	}
}

func TestReadFileAllByte_NotExist(t *testing.T) {
	_, err := ReadFileAllByte("./not_exist_file_12345.txt")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestReadFileAllString(t *testing.T) {
	content, err := ReadFileAllString("./testtmp/a.txt")
	if err != nil {
		t.Fatal("ReadFileAllString err:", err)
	}
	expected := "kkk"
	if content != expected && content != expected+"\n" {
		t.Errorf("ReadFileAllString result = %q, want %q", content, expected)
	}
}

func TestReadFileAllString_NotExist(t *testing.T) {
	_, err := ReadFileAllString("./not_exist_file_12345.txt")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestReadFileAllByte_EmptyFile(t *testing.T) {
	// 创建临时空文件
	tmpFile, err := os.CreateTemp(t.TempDir(), "empty_*.txt")
	if err != nil {
		t.Fatal("create temp file err:", err)
	}
	tmpFile.Close()

	content, err := ReadFileAllByte(tmpFile.Name())
	if err != nil {
		t.Fatal("ReadFileAllByte err:", err)
	}
	if len(content) != 0 {
		t.Errorf("expected empty content for empty file, got %d bytes", len(content))
	}
}

func TestReadFileAllString_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "empty_*.txt")
	if err != nil {
		t.Fatal("create temp file err:", err)
	}
	tmpFile.Close()

	content, err := ReadFileAllString(tmpFile.Name())
	if err != nil {
		t.Fatal("ReadFileAllString err:", err)
	}
	if content != "" {
		t.Errorf("expected empty string for empty file, got %q", content)
	}
}

// hasSuffix 测试辅助函数，检查字符串是否以指定后缀结尾
func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}