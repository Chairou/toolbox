package gin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// CorsMiddleware CORS 中间件，允许所有跨域请求
// 使用方法：
//
//	router := gin.Default()
//	router.Use(CorsMiddleware())
func CorsMiddleware(c *Context) {
	// 允许所有来源
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// 允许的请求方法
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD")

	// 允许的请求头
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Token")

	// 允许浏览器访问的响应头
	c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")

	// 允许携带凭证（cookies）
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	// 预检请求的缓存时间（秒）
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")

	// 处理 OPTIONS 预检请求
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()

}

// CustomCorsMiddleware 自定义 CORS 中间件，可配置具体参数
// 使用方法：
//
//	router := gin.Default()
//	router.Use(CustomCorsMiddleware(CorsConfig{
//	    AllowOrigins: []string{"https://example.com"},
//	    AllowMethods: []string{"GET", "POST"},
//	}))
func CustomCorsMiddleware(config CorsConfig) gin.HandlerFunc {
	// 设置默认值
	if len(config.AllowOrigins) == 0 {
		config.AllowOrigins = []string{"*"}
	}
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	}
	if len(config.AllowHeaders) == 0 {
		config.AllowHeaders = []string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "accept", "origin", "Cache-Control", "X-Requested-With", "Token"}
	}
	if len(config.ExposeHeaders) == 0 {
		config.ExposeHeaders = []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"}
	}
	if config.MaxAge == 0 {
		config.MaxAge = 86400 // 默认 24 小时
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查来源是否允许
		allowOrigin := "*"
		if len(config.AllowOrigins) > 0 && config.AllowOrigins[0] != "*" {
			for _, o := range config.AllowOrigins {
				if o == origin {
					allowOrigin = origin
					break
				}
			}
			if allowOrigin == "*" && origin != "" {
				// 如果配置了具体的来源但当前来源不在列表中，则不允许
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", joinStrings(config.AllowMethods, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Headers", joinStrings(config.AllowHeaders, ", "))
		c.Writer.Header().Set("Access-Control-Expose-Headers", joinStrings(config.ExposeHeaders, ", "))

		if config.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if config.MaxAge > 0 {
			c.Writer.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
		}

		// 处理 OPTIONS 预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// CorsConfig CORS 配置结构体
type CorsConfig struct {
	// 允许的来源列表，使用 "*" 表示允许所有来源
	AllowOrigins []string

	// 允许的 HTTP 方法
	AllowMethods []string

	// 允许的请求头
	AllowHeaders []string

	// 允许浏览器访问的响应头
	ExposeHeaders []string

	// 是否允许携带凭证
	AllowCredentials bool

	// 预检请求的缓存时间（秒）
	MaxAge int
}

// joinStrings 辅助函数，用于连接字符串数组
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
