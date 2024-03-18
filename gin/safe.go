package gin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
)

func SafeCheck() gin.HandlerFunc {
	return func(c *gin.Context) {

		switch c.Request.Method {
		case "GET":
			queryParams := c.Request.URL.Query()
			// 遍历参数并打印
			for key, values := range queryParams {
				for _, value := range values {
					// 这里写检查语句啦
					fmt.Printf("参数：%s，值：%s\n", key, value)
					if detectSQLInjection(value) == true {
						c.Abort()
						c.JSON(http.StatusUnauthorized, gin.H{"message": "访问未授权"})
					}
				}
			}
			// 检查通过，退出
			c.Next()
		}
	}
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
