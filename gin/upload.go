package gin

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strings"
)

func RecUploadFile(c *gin.Context) {
	ErrOk := 0
	ErrInvalidFilename := -11
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"upload file error": err.Error()})
		return
	}
	if isSafeFileName(file.Filename) == false {
		WriteRetJson(c, ErrInvalidFilename, nil, "InvalidFilename = ", file.Filename)
		return
	}

	// 将文件保存到本地
	path := os.Getenv("uploadFilePath")
	if path == "" {
		path = "/tmp/"
	}
	err = c.SaveUploadedFile(file, path+file.Filename)
	if err != nil {
		WriteRetJson(c, ErrOk, nil, "fileName = ", file.Filename)
	}
}

func isSafeFileName(fileName string) (ok bool) {
	// 定义不安全的字符集合
	unsafeChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	for _, v := range unsafeChars {
		if strings.Contains(fileName, v) {
			return false
		}
	}
	return true
}
