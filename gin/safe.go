package gin

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Chairou/toolbox/util/encode"
	"github.com/Chairou/toolbox/util/listopt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"reflect"
	"regexp"
	"strings"
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
	seq := encode.Sha512([]byte(uuid.New().String()))[16:24]
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

// Recursive function to check fields and nested structures
func checkFields(value reflect.Value, prefix string, errors *[]string) {
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field := value.Field(i)
			fieldType := value.Type().Field(i)
			prefixedFieldName := fmt.Sprintf("%s.%s", prefix, fieldType.Name)
			checkFields(field, prefixedFieldName, errors)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			element := value.Index(i)
			prefixedElementName := fmt.Sprintf("%s[%d]", prefix, i)
			checkFields(element, prefixedElementName, errors)
		}
	default:
		// Custom checks (example: check for empty strings)
		if value.Kind() == reflect.String {
			if detectSQLInjection(value.String()) {
				*errors = append(*errors,
					fmt.Sprintf("Field Name: %s, Field Value: %v has SQL Injection",
						prefix, value.Interface()),
				)
			}
		}
	}
}

func ValidateSql(value reflect.Value, structName string) error {
	var errorList []string
	checkFields(value, structName, &errorList)
	if len(errorList) > 0 {
		return errors.New(fmt.Sprintf("Validation failed: %v", errorList))
	}
	return nil
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
	largeTransferList := []string{"File Transfer"}
	if !listopt.IsInStringArr(largeTransferList, c.GetHeader("Content-Description")) {
		fmt.Println("Response body: " + blw.body.String())
	} else {
		fmt.Println("大文件传输中，不记录response body")
	}

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

// EscapeString 手动转义 SQL 字符串中的特殊字符
func EscapeString(value string) string {
	var replacements = []struct {
		old string
		new string
	}{
		{`'`, `\'`},
		{`"`, `\"`},
		{`\`, `\\`},
		{`\n`, `\\n`},
		{`\r`, `\\r`},
		{`\x00`, `\\0`},
		{`\x1a`, `\\Z`},
		{"'", `\'`},
		{"\"", `\"`},
		{"\\", `\\`},
		{"\n", `\\n`},
		{"\r", `\\r`},
		{"\x00", `\\0`},
		{"\x1a", `\\Z`},
	}

	for _, r := range replacements {
		value = strings.ReplaceAll(value, r.old, r.new)
	}
	return value
}

func EscapeFields(value reflect.Value, prefix string) {
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field := value.Field(i)
			fieldType := value.Type().Field(i)
			prefixedFieldName := fmt.Sprintf("%s.%s", prefix, fieldType.Name)
			EscapeFields(field, prefixedFieldName)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			element := value.Index(i)
			prefixedElementName := fmt.Sprintf("%s[%d]", prefix, i)
			EscapeFields(element, prefixedElementName)
		}
	default:
		// Custom checks (example: check for empty strings)
		if value.Kind() == reflect.String {
			value.SetString(EscapeString(value.String()))
		}
	}
}
