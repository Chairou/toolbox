package httphelper

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var (
	// Log 日志输出函数，可修改
	Log = func(v ...interface{}) { fmt.Println(v...) }
)

// NewRequest 创建新的请求
func NewRequest(method string, urlStr string, body io.Reader) Helper {
	var err error
	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 90 * time.Second,
			}).DialContext,
			MaxIdleConns:        300,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     30 * time.Second,
		},
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return errorHelper(fmt.Errorf("new request error: %w", err))
	}

	helper := &httpHelper{
		client: *http.DefaultClient,
		req:    req,
	}
	return helper
}

// GET 创建新的GET请求
func GET(url string) Helper {
	return NewRequest("GET", url, nil)
}

// PostUrlEncode 创建application/x-www-form-urlencoded的POST请求
func PostUrlEncode(url string, values url.Values) Helper {
	return NewRequest("POST", url, strings.NewReader(values.Encode())).SetHeader("Content-Type", "application/x-www-form-urlencoded")
}

// PostJSON 创建application/json的POST请求
func PostJSON(url string, body interface{}) Helper {
	b, _ := jsoniter.Marshal(body)
	return NewRequest("POST", url, bytes.NewReader(b)).SetHeader("Content-Type", "application/json")
}

func PathEscape(param map[string]string) string {
	var buf strings.Builder
	keys := make([]string, 0, len(param))
	for k, v := range param {
		if v != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		keyEscaped := url.PathEscape(k)
		buf.WriteString(keyEscaped)
		buf.WriteByte('=')
		buf.WriteString(url.PathEscape(param[k]))
		buf.WriteByte('&')
	}
	return strings.TrimRight(buf.String(), "&")
}

func UrlToMap(url string) map[string]string {
	argMap := make(map[string]string, 0)
	pos := strings.Index(url, "?")
	argLine := url[pos+1:]
	argList := strings.Split(argLine, "&")

	for _, v := range argList {
		rowList := strings.Split(v, "=")
		argMap[rowList[0]] = rowList[1]
	}
	return argMap
}

func UrlPathEscape(url string) string {
	pos := strings.Index(url, "?")
	urlHeader := url[0 : pos+1]
	urlParam := url[pos+1:]
	encodeStr := PathEscape(UrlToMap(urlParam))
	return urlHeader + encodeStr

}
