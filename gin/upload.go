package gin

import (
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// safeNameRe 文件名白名单：首字符为字母、数字或汉字（\x{4e00}-\x{9fa5}），
// 其余允许字母、数字、汉字及 . _ - 连接符，总长 1~64 个字符。
// 包级预编译，避免每次校验都重新编译正则。
// var safeNameRe = regexp.MustCompile(`^[A-Za-z0-9\x{4e00}-\x{9fa5}][A-Za-z0-9\x{4e00}-\x{9fa5}._-]{0,63}$`)
var safeNameRe = regexp.MustCompile(`^[A-Za-z0-9\p{Han}][A-Za-z0-9\p{Han}._-]{0,63}$`)

func RecUploadFile(c *Context) {
	const (
		ErrOk              = 0
		ErrInvalidFilename = -11
		ErrSaveFileFailed  = -12
		ErrFileTooLarge    = -13
		ErrUploadDir       = -14
	)

	var maxUploadSize int64
	maxUploadSize = 1024 * 1024 * 100
	var metaHeaderSize int64
	metaHeaderSize = 1024 * 1024
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize+metaHeaderSize)
	if err := c.Request.ParseMultipartForm(maxUploadSize); err != nil {
		WriteRetJson(c, ErrFileTooLarge, nil, "upload rejected: ", err)
		return
	}
	if c.Request.MultipartForm != nil {
		defer func(MultipartForm *multipart.Form) {
			_ = MultipartForm.RemoveAll()
		}(c.Request.MultipartForm)
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrInvalidFilename, "message": "missing file field"})
		return
	}
	if file.Size <= 0 || file.Size > maxUploadSize {
		WriteRetJson(c, ErrFileTooLarge, nil, "file size out of range")
		return
	}
	if isSafeFileName(file.Filename) == false {
		WriteRetJson(c, ErrInvalidFilename, nil, "InvalidFilename")
		return
	}

	// 将文件保存到本地
	path := os.Getenv("uploadFilePath")
	if path == "" {
		path = os.TempDir()
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		WriteRetJson(c, ErrUploadDir, nil, "upload dir: "+err.Error())
		return
	}
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		WriteRetJson(c, ErrUploadDir, nil, "upload dir is not a usable directory")
		return
	}
	err = c.SaveUploadedFile(file, filepath.Join(abs, file.Filename))
	if err != nil {
		WriteRetJson(c, ErrSaveFileFailed, nil, "ErrSaveFileFailed")
		return
	}
	WriteRetJson(c, ErrOk, nil, "ok")
}

func isSafeFileName(fileName string) (ok bool) {
	if !safeNameRe.MatchString(fileName) {
		return false
	}
	// 文件名长度受文件系统按字节限制（NAME_MAX=255），中文 UTF-8 占 3 字节
	if len(fileName) > 200 {
		return false
	}
	if fileName == "." || fileName == ".." || strings.HasPrefix(fileName, ".") {
		return false
	}
	stem := strings.ToUpper(strings.TrimSuffix(fileName, filepath.Ext(fileName)))
	switch stem {
	case "CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		return false
	}
	return true
}
