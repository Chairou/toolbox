package gin

import (
	"net/http"
	"strconv"
	"strings"
)

// 常用 CORS 响应头常量，避免魔法字符串
const (
	headerOrigin           = "Origin"
	headerAllowOrigin      = "Access-Control-Allow-Origin"
	headerAllowMethods     = "Access-Control-Allow-Methods"
	headerAllowHeaders     = "Access-Control-Allow-Headers"
	headerAllowCredentials = "Access-Control-Allow-Credentials"
	headerExposeHeaders    = "Access-Control-Expose-Headers"
	headerMaxAge           = "Access-Control-Max-Age"
	headerRequestMethod    = "Access-Control-Request-Method"
	headerRequestHeaders   = "Access-Control-Request-Headers"
	headerVary             = "Vary"

	// 预检请求缓存时长：24 小时，减少浏览器重复发送 OPTIONS
	defaultMaxAge = 86400
)

// CorsMiddleware CORS 中间件——放通所有跨域请求（包括携带凭证的请求）
//
// 实现原理：采用「动态反射」策略，而非静态白名单。
//  1. Origin：反射请求中的 Origin 头（不能写死 "*"，否则无法与 Credentials 共存）
//  2. Methods：反射预检请求中的 Access-Control-Request-Method（任意 HTTP 方法均可）
//  3. Headers：反射预检请求中的 Access-Control-Request-Headers（任意自定义头部均可）
//  4. Vary：告知 CDN/代理缓存需按 Origin/请求头 分桶缓存，避免缓存污染
//
// 使用方法：
//
//	router := gin.Default()
//	router.Use(CorsMiddleware)
func CorsMiddleware(c *Context) {
	origin := c.Request.Header.Get(headerOrigin)

	// 反射 Origin，兼容 Credentials；无 Origin（同源或非浏览器请求）时降级为 "*"
	if origin != "" {
		c.Writer.Header().Set(headerAllowOrigin, origin)
	} else {
		c.Writer.Header().Set(headerAllowOrigin, "*")
	}

	// 告知下游缓存：响应内容依赖 Origin，防止 CDN 缓存串号
	c.Writer.Header().Add(headerVary, headerOrigin)

	// 允许携带凭证（Cookies、HTTP 认证、TLS 客户端证书）
	c.Writer.Header().Set(headerAllowCredentials, "true")

	// 允许浏览器 JS 读取所有响应头
	c.Writer.Header().Set(headerExposeHeaders, "*")

	// 预检请求缓存
	c.Writer.Header().Set(headerMaxAge, strconv.Itoa(defaultMaxAge))

	// 处理 OPTIONS 预检：客户端要什么就返回什么
	if c.Request.Method == http.MethodOptions {
		// 反射请求方法：放通任意 HTTP 方法
		if reqMethod := c.Request.Header.Get(headerRequestMethod); reqMethod != "" {
			c.Writer.Header().Set(headerAllowMethods, reqMethod)
			c.Writer.Header().Add(headerVary, headerRequestMethod)
		} else {
			c.Writer.Header().Set(headerAllowMethods, "GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD")
		}
		// 反射请求头：放通任意自定义头部
		if reqHeaders := c.Request.Header.Get(headerRequestHeaders); reqHeaders != "" {
			c.Writer.Header().Set(headerAllowHeaders, reqHeaders)
			c.Writer.Header().Add(headerVary, headerRequestHeaders)
		}
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}

// CorsConfig CORS 配置结构体
type CorsConfig struct {
	// AllowOrigins 允许的来源列表，使用 []string{"*"} 表示放通所有
	AllowOrigins []string

	// AllowOriginFunc 动态判定回调（优先级高于 AllowOrigins），返回 true 表示允许
	// 适用于按后缀/正则匹配等复杂场景，例如：
	//   func(o string) bool { return strings.HasSuffix(o, ".example.com") }
	AllowOriginFunc func(origin string) bool

	// AllowMethods 允许的 HTTP 方法
	AllowMethods []string

	// AllowHeaders 允许的请求头
	AllowHeaders []string

	// ExposeHeaders 允许浏览器 JS 访问的响应头
	ExposeHeaders []string

	// AllowCredentials 是否允许携带凭证。开启此项时，AllowOrigins 不能为 "*"
	AllowCredentials bool

	// MaxAge 预检请求的缓存时间（秒），<=0 时不设置
	MaxAge int
}

// CustomCorsMiddleware 自定义 CORS 中间件，可精细化配置
//
// 使用方法：
//
//	router.Use(CustomCorsMiddleware(CorsConfig{
//	    AllowOrigins:     []string{"https://example.com"},
//	    AllowMethods:     []string{"GET", "POST"},
//	    AllowCredentials: true,
//	}))
//
// 或使用动态判定：
//
//	router.Use(CustomCorsMiddleware(CorsConfig{
//	    AllowOriginFunc: func(o string) bool {
//	        return strings.HasSuffix(o, ".example.com")
//	    },
//	    AllowCredentials: true,
//	}))
func CustomCorsMiddleware(config CorsConfig) HandlerFunc {
	// 默认值填充
	if len(config.AllowOrigins) == 0 && config.AllowOriginFunc == nil {
		config.AllowOrigins = []string{"*"}
	}
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	}
	if len(config.AllowHeaders) == 0 {
		config.AllowHeaders = []string{
			"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token",
			"Authorization", "Accept", "Origin", "Cache-Control", "X-Requested-With", "Token",
		}
	}
	if len(config.ExposeHeaders) == 0 {
		config.ExposeHeaders = []string{"Content-Length", "Content-Type"}
	}
	if config.MaxAge == 0 {
		config.MaxAge = defaultMaxAge
	}

	// 预计算：拼接字符串只做一次，避免每次请求都执行 strings.Join
	allowMethodsStr := strings.Join(config.AllowMethods, ", ")
	allowHeadersStr := strings.Join(config.AllowHeaders, ", ")
	exposeHeadersStr := strings.Join(config.ExposeHeaders, ", ")
	maxAgeStr := strconv.Itoa(config.MaxAge)

	// 预判：是否为「放通所有」通配模式
	allowAll := false
	for _, o := range config.AllowOrigins {
		if o == "*" {
			allowAll = true
			break
		}
	}

	return func(c *Context) {
		origin := c.Request.Header.Get(headerOrigin)
		allowOrigin := resolveAllowOrigin(origin, allowAll, &config)

		// Origin 不在白名单：直接拒绝预检 & 请求
		if allowOrigin == "" && origin != "" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// 特殊约束：Credentials=true 时，Origin 不能返回 "*"，必须反射具体来源
		if allowOrigin == "*" && config.AllowCredentials && origin != "" {
			allowOrigin = origin
		}

		if allowOrigin != "" {
			c.Writer.Header().Set(headerAllowOrigin, allowOrigin)
		}
		// 只要不是纯 "*"，就要加 Vary: Origin，避免缓存串号
		if allowOrigin != "*" {
			c.Writer.Header().Add(headerVary, headerOrigin)
		}

		c.Writer.Header().Set(headerAllowMethods, allowMethodsStr)
		c.Writer.Header().Set(headerAllowHeaders, allowHeadersStr)
		c.Writer.Header().Set(headerExposeHeaders, exposeHeadersStr)

		if config.AllowCredentials {
			c.Writer.Header().Set(headerAllowCredentials, "true")
		}
		if config.MaxAge > 0 {
			c.Writer.Header().Set(headerMaxAge, maxAgeStr)
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// resolveAllowOrigin 根据配置解析最终返回的 Access-Control-Allow-Origin 值
//   - 返回 ""：表示禁止此来源
//   - 返回 "*"：表示放通所有
//   - 返回具体域名：表示反射该 Origin
func resolveAllowOrigin(origin string, allowAll bool, config *CorsConfig) string {
	// 优先使用动态回调
	if config.AllowOriginFunc != nil {
		if origin != "" && config.AllowOriginFunc(origin) {
			return origin
		}
		return ""
	}
	if allowAll {
		return "*"
	}
	for _, o := range config.AllowOrigins {
		if o == origin {
			return origin
		}
	}
	return ""
}