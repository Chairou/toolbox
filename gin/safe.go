package gin

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"regexp"
)

func SafeCheck(c *Context) {
	switch c.Request.Method {
	case "GET":
		queryParams := c.Request.URL.Query()
		// 遍历参数并打印
		for key, values := range queryParams {
			for _, value := range values {
				// 这里写检查语句啦
				if detectSQLInjection(value) == true {
					c.Abort()
					c.JSON(http.StatusUnauthorized, H{"message": "访问未授权"})
					fmt.Printf("参数：%s，值：%s\n", key, value)
				}
			}
		}
	}
	// 检查通过，设置seq，退出
	seq := uuid.New()
	c.Set("seq", seq)
	c.Next()
}

func detectSQLInjection(input string) bool {
	// 正则表达式用于检测一些SQL注入的常见模式
	// 注意：这些正则表达式非常简单，实际情况可能需要更复杂的模式
	var sqlInjectionPatterns = []string{
		"(?i)union(.*)select", // 检测UNION SELECT攻击
		"(?i)insert(.*)into",  // 检测INSERT INTO攻击
		"(?i)select(.*)from",  // 检测SELECT FROM攻击
		"(?i)delete(.*)from",  // 检测DELETE FROM攻击
		"(?i)update(.*)set",   // 检测UPDATE SET攻击
		"(?i)drop(.*)table",   // 检测DROP TABLE攻击
		"--",                  // 检测单行注释
		";",                   // 检测分号
		"/\\*.*?\\*/",         // 检测多行注释
		"'",                   // 检测单引号
		"\"",                  // 检测双引号
	}

	for _, pattern := range sqlInjectionPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return true
		}
	}
	return false
}

// ResponseRecorder 中间件用于记录响应数据
func ResponseRecorder(c *Context) {
	// 创建一个新的响应体
	blw := &bodyLogWriter{
		body:           bytes.NewBufferString(""),
		ResponseWriter: c.Writer,
	}
	c.Writer = blw

	// 处理请求
	c.Next()

	// 请求处理完成后，记录响应体
	fmt.Println("Response body: " + blw.body.String())

}

// bodyLogWriter 是一个包装了响应体的结构体
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 重写Write方法以捕获响应体数据
func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
