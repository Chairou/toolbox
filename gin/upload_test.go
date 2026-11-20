package gin

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

// buildUploadRequest 构造一个包含 file 字段的 multipart 请求
func buildUploadRequest(t *testing.T, fieldName, fileName string, content []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if fieldName != "" {
		fw, err := mw.CreateFormFile(fieldName, fileName)
		if err != nil {
			t.Fatalf("CreateFormFile: %v", err)
		}
		if _, err := fw.Write(content); err != nil {
			t.Fatalf("write form file: %v", err)
		}
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

// runRecUploadFile 以 gin 测试上下文执行 RecUploadFile，返回响应记录器
func runRecUploadFile(t *testing.T, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	gc, _ := gin.CreateTestContext(w)
	gc.Request = req
	RecUploadFile(&Context{Context: gc})
	return w
}

// parseRet 解析 WriteRetJson 输出的统一返回结构
func parseRet(t *testing.T, w *httptest.ResponseRecorder) Ret {
	t.Helper()
	var ret Ret
	if err := json.Unmarshal(w.Body.Bytes(), &ret); err != nil {
		t.Fatalf("unmarshal response %q: %v", w.Body.String(), err)
	}
	return ret
}

// TestRecUploadFile_Success 正常上传应落盘到 uploadFilePath 目录
func TestRecUploadFile_Success(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("uploadFilePath", dir)

	content := []byte("hello upload")
	req := buildUploadRequest(t, "file", "hello.txt", content)
	w := runRecUploadFile(t, req)

	if ret := parseRet(t, w); ret.Code != 0 {
		t.Errorf("code = %d, want 0", ret.Code)
	}
	got, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatalf("uploaded file not found: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("content mismatch: got %q, want %q", got, content)
	}
}

// TestRecUploadFile_ChineseFileName 中文文件名应被允许
func TestRecUploadFile_ChineseFileName(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("uploadFilePath", dir)

	req := buildUploadRequest(t, "file", "测试文件.txt", []byte("x"))
	runRecUploadFile(t, req)

	if _, err := os.Stat(filepath.Join(dir, "测试文件.txt")); err != nil {
		t.Errorf("chinese file name not saved: %v", err)
	}
}

// TestRecUploadFile_PathTraversalStripped 路径穿越文件名会被 multipart 取 basename，
// 文件只能落在上传目录内，不能写到上级目录
func TestRecUploadFile_PathTraversalStripped(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "uploads")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("uploadFilePath", dir)

	req := buildUploadRequest(t, "file", "../evil.txt", []byte("x"))
	runRecUploadFile(t, req)

	if _, err := os.Stat(filepath.Join(root, "evil.txt")); err == nil {
		t.Fatal("file escaped upload directory")
	}
}

// TestRecUploadFile_UnsafeFileName 含非法字符的文件名应被拒绝
func TestRecUploadFile_UnsafeFileName(t *testing.T) {
	cases := []string{
		"a?b.txt",
		"a*b.txt",
		".hidden",
		"CON.txt",
		"NUL",
		"a b.txt",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("uploadFilePath", dir)

			req := buildUploadRequest(t, "file", name, []byte("x"))
			w := runRecUploadFile(t, req)

			if ret := parseRet(t, w); ret.Code != -11 {
				t.Errorf("code = %d, want -11", ret.Code)
			}
			entries, _ := os.ReadDir(dir)
			if len(entries) != 0 {
				t.Errorf("unsafe file was written: %d entries", len(entries))
			}
		})
	}
}

// TestRecUploadFile_MissingFileField 缺少 file 字段应返回 400
func TestRecUploadFile_MissingFileField(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("uploadFilePath", dir)

	req := buildUploadRequest(t, "", "", nil)
	w := runRecUploadFile(t, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// TestRecUploadFile_EmptyFile 空文件应被拒绝
func TestRecUploadFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("uploadFilePath", dir)

	req := buildUploadRequest(t, "file", "empty.txt", []byte{})
	w := runRecUploadFile(t, req)

	if ret := parseRet(t, w); ret.Code != -13 {
		t.Errorf("code = %d, want -13", ret.Code)
	}
	if _, err := os.Stat(filepath.Join(dir, "empty.txt")); err == nil {
		t.Error("empty file should not be saved")
	}
}

// TestRecUploadFile_UploadDirNotExist 上传目录不存在应返回 -14
func TestRecUploadFile_UploadDirNotExist(t *testing.T) {
	t.Setenv("uploadFilePath", filepath.Join(t.TempDir(), "missing"))

	req := buildUploadRequest(t, "file", "a.txt", []byte("x"))
	w := runRecUploadFile(t, req)

	if ret := parseRet(t, w); ret.Code != -14 {
		t.Errorf("code = %d, want -14", ret.Code)
	}
}
