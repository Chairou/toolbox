package gin

import (
	"fmt"
	"github.com/Chairou/toolbox/util/conv"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type Ret struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
	Seq  string      `json:"seq"`
}

type H map[string]any

func WriteRetJson(c *gin.Context, code int, data interface{}, args ...interface{}) {
	var msg string
	var ret Ret
	ret.Code = code
	ret.Data = data

	seq, ok := c.Get("seq")
	if ok {
		ret.Seq = conv.String(seq)
	} else {
		ret.Seq = ""
	}

	for _, arg := range args {
		switch v := arg.(type) {
		case error:
			msg += v.Error()
		case string:
			// 处理string类型
			msg += v
		case int:
			// 处理int类型
			msg += fmt.Sprintf("%d", v)
		case int32:
			// 处理int类型
			msg += fmt.Sprintf("%d", v)
		case int64:
			// 处理int类型
			msg += fmt.Sprintf("%d", v)
		case float32:
			// 处理float64类型
			msg += fmt.Sprintf("%f", v)
		case float64:
			// 处理float64类型
			msg += fmt.Sprintf("%f", v)
		case bool:
			// 处理bool类型
			msg += fmt.Sprintf(" %t", v)
		default:
			// 处理其他类型
			msg += fmt.Sprintf(" %v", v)
		}
	}
	ret.Msg = msg
	c.JSON(http.StatusOK, ret)
}

func GetStringDefault(c *gin.Context, param string, defaultValue string) string {
	val := GetParamValue(c, param)
	val = strings.Trim(val, " \t\n\r")
	if val == "" {
		return defaultValue
	}
	return val
}

func GetIntDefault(c *gin.Context, param string, defaultValue int) int {
	val := GetParamValue(c, param)
	val = strings.Trim(val, " \t\n\r")
	intVal, ok := conv.Int(val)
	if !ok {
		return defaultValue
	}
	return intVal
}

// GetParamValue get parameter string value
func GetParamValue(c *gin.Context, name string) string {
	// query string
	v := c.Query(name)

	// post form
	if v == "" {
		v = c.PostForm(name)
	}

	// restful
	if v == "" {
		v = c.Param(name)
	}

	return strings.TrimSpace(v)
}
