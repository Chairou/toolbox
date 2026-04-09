package gin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/Chairou/toolbox/util/encode"
	"github.com/Chairou/toolbox/util/listopt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 预编译 SQL 注入检测正则表达式，避免每次调用都重新编译
var sqlInjectionPatterns []*regexp.Regexp

func init() {
	patterns := []string{
		`(?i)\bunion\b.*\bselect\b`, // 检测UNION SELECT攻击
		`(?i)\binsert\b.*\binto\b`,  // 检测INSERT INTO攻击
		`(?i)\bselect\b.*\bfrom\b`,  // 检测SELECT FROM攻击
		`(?i)\bdelete\b.*\bfrom\b`,  // 检测DELETE FROM攻击
		`(?i)\bupdate\b.*\bset\b`,   // 检测UPDATE SET攻击
		`(?i)\bdrop\b.*\btable\b`,   // 检测DROP TABLE攻击
		`--`,                        // 检测单行注释
		`/\*.*?\*/`,                 // 检测多行注释
	}
	for _, p := range patterns {
		sqlInjectionPatterns = append(sqlInjectionPatterns, regexp.MustCompile(p))
	}
}

func SafeCheck(c *Context) {
	switch c.Request.Method {
	case "GET":
		queryParams := c.Request.URL.Query()
		// 遍历参数并检查
		for key, values := range queryParams {
			for _, value := range values {
				if detectSQLInjection(value) {
					c.JSON(http.StatusForbidden, H{"message": "访问被禁止"})
					fmt.Printf("SQL注入检测 - 参数：%s，值：%s\n", key, value)
					c.Abort()
					return
				}
			}
		}
	case "POST", "PUT", "DELETE", "PATCH":
		// 读取 body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Abort()
			return
		}
		// 重要：读取后必须重新填充 body，否则后续 handler 无法读取
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 解析为 map 后递归遍历
		var jsonData map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
			for key, value := range jsonData {
				if strVal, ok := value.(string); ok {
					if detectSQLInjection(strVal) {
						c.JSON(http.StatusForbidden, H{"message": "访问被禁止"})
						fmt.Printf("SQL注入检测 - 参数：%s，值：%s\n", key, strVal)
						c.Abort()
						return
					}
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
	for _, re := range sqlInjectionPatterns {
		if re.MatchString(input) {
			return true
		}
	}
	return false
}

// checkFields 递归检查结构体字段中是否存在 SQL 注入
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
		return fmt.Errorf("Validation failed: %v", errorList)
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
	if !listopt.IsInStringArr(largeTransferList, blw.ResponseWriter.Header().Get("Content-Description")) {
		fmt.Println(time.Now().Format(time.DateTime), " [Response body]: "+blw.body.String())
	} else {
		bodyStr := blw.body.String()
		if len(bodyStr) > 512 {
			fmt.Println(time.Now().Format(time.DateTime), " [大数据量传输中，只输出头部512字节]: ", "\n", bodyStr[:512])
		} else {
			fmt.Println(time.Now().Format(time.DateTime), " [Response body]: "+bodyStr)
		}
	}
}

// bodyLogWriter 是一个包装了响应体的结构体
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 重写Write方法以捕获响应体数据（使用指针接收者）
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// EscapeString 手动转义 SQL 字符串中的特殊字符
// 注意：反斜杠必须最先替换，避免后续替换产生的反斜杠被二次转义
func EscapeString(value string) string {
	var replacements = []struct {
		old string
		new string
	}{
		{"\\", `\\`}, // 反斜杠必须第一个替换
		{"'", `\'`},
		{"\"", `\"`},
		{"\n", `\n`},
		{"\r", `\r`},
		{"\x00", `\0`},
		{"\x1a", `\Z`},
	}

	for _, r := range replacements {
		value = strings.ReplaceAll(value, r.old, r.new)
	}
	return value
}

// EscapeFields 递归转义结构体中所有字符串字段的 SQL 特殊字符
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
		if value.Kind() == reflect.String {
			value.SetString(EscapeString(value.String()))
		}
	}
}
