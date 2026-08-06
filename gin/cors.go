package gin

import (
	"net/http"
	"strconv"
	"strings"
)

// CorsMiddleware CORS 中间件，放行所有跨域请求（包括携带凭证的请求）
//
// 实现原理：采用动态反射策略，而非静态白名单。
// 1. Origin：反射请求中的 Origin 头，而非写死 "*"，以此兼容 Access-Control-Allow-Credentials
// 2. Methods：反射预检请求中的 Access-Control-Request-Method，任意 HTTP 方法均可通过
// 3. Headers：反射预检请求中的 Access-Control-Request-Headers，任意自定义头部均可通过
//
// 使用方法：
//
//	router := gin.Default()
//	router.Use(CorsMiddleware())
func CorsMiddleware(c *Context) {
	// 动态反射请求中的 Origin，以支持携带凭证的跨域请求（不能与 "*" 共存）
	origin := c.Request.Header.Get("Origin")
	if origin != "" {
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	}

	// 允许携带凭证（cookies、HTTP认证、TLS客户端证书等）
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	// 允许浏览器读取所有响应头
	c.Writer.Header().Set("Access-Control-Expose-Headers", "*")

	// 预检请求缓存 24 小时，减少不必要的 OPTIONS 请求
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")

	// 处理 OPTIONS 预检请求：客户端要什么就返回什么
	if c.Request.Method == "OPTIONS" {
		// 反射请求方法：放行任意 HTTP 方法
		if reqMethod := c.Request.Header.Get("Access-Control-Request-Method"); reqMethod != "" {
			c.Writer.Header().Set("Access-Control-Allow-Methods", reqMethod)
		}
		// 反射请求头：放行任意自定义头部
		if reqHeaders := c.Request.Header.Get("Access-Control-Request-Headers"); reqHeaders != "" {
			c.Writer.Header().Set("Access-Control-Allow-Headers", reqHeaders)
		}
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
func CustomCorsMiddleware(config CorsConfig) HandlerFunc {
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

	return func(c *Context) {
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
		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
		c.Writer.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))

		if config.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if config.MaxAge > 0 {
			c.Writer.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
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
