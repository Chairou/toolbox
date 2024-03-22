package httphelper

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var (
	// Log 日志输出函数，可修改
	Log = func(v ...interface{}) { fmt.Println(v...) }
)
var once sync.Once
var gCookies sync.Map

// NewRequest 创建新的请求
func NewRequest(method string, urlStr string, body io.Reader) Helper {
	once.Do(func() {
		http.DefaultClient = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 90 * time.Second,
				}).DialContext,
				MaxIdleConns:          300,
				MaxIdleConnsPerHost:   100,
				IdleConnTimeout:       30 * time.Second,
				ResponseHeaderTimeout: 5 * time.Second,
				ForceAttemptHTTP2:     true,
				ReadBufferSize:        65536,
				WriteBufferSize:       65536,
			},
			Timeout: 30 * time.Second,
		}
	})
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return errorHelper(fmt.Errorf("new request error: %w", err))
		//return &httpHelper{
		//	client: *http.DefaultClient,
		//	req:    req,
		//	debug:  DEBUG_NORMAL,
		//	Err:    err,
		//}
	}

	helper := &httpHelper{
		client: *http.DefaultClient,
		req:    req,
		debug:  DebugNormal,
		Err:    nil,
	}
	return helper
}

// GET 创建新的GET请求
func GET(url string) Helper {
	return NewRequest("GET", url, nil)
}

// PostUrlEncode 创建application/x-www-form-urlencoded的POST请求
func PostUrlEncode(url string, values url.Values) Helper {
	return NewRequest("POST", url, strings.NewReader(values.Encode())).SetHeader("Content-Type",
		"application/x-www-form-urlencoded")
}

// PostJSON 创建application/json的POST请求
func PostJSON(url string, body interface{}) Helper {
	switch value := body.(type) {
	case string:
		return NewRequest("POST", url, strings.NewReader(value)).SetHeader("Content-Type",
			"application/json")
	default:
		byteBody, err := jsoniter.Marshal(body)
		if err != nil {
			return errorHelper(fmt.Errorf("new request error: %w", err))
		}
		return NewRequest("POST", url, bytes.NewReader(byteBody)).SetHeader("Content-Type",
			"application/json")

	}
}

func PostFile(url string, fullPathSourceFileName string, DstFileName string) Helper {
	// 打开要上传的文件
	file, err := os.Open(fullPathSourceFileName)
	if err != nil {
		fmt.Println("无法打开文件:", err)
		return errorHelper(fmt.Errorf("无法打开文件: %w", err))
	}
	stat, _ := file.Stat()
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("close file err:", err)
		}
	}(file)

	// 创建一个缓冲区来存储请求体
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建一个表单字段，将文件内容写入其中
	fileField, err := writer.CreateFormFile("file", DstFileName)
	if err != nil {
		fmt.Println("创建表单字段失败:", err)
		return errorHelper(fmt.Errorf("创建表单字段失败: %w", err))
	}
	_, err = io.Copy(fileField, file)
	if err != nil {
		fmt.Println("写入文件内容失败:", err)
		return errorHelper(fmt.Errorf("创建表单字段失败: %w", err))
	}

	// 完成表单写入
	err = writer.Close()
	if err != nil {
		fmt.Println("关闭表单写入失败:", err)
		return errorHelper(fmt.Errorf("关闭表单写入失败: %w", err))
	}

	// 创建一个POST请求
	httpClient := NewRequest("POST", url, body).SetHeader("Content-Type",
		writer.FormDataContentType())
	httpClient.SetDebug(DebugUpload)
	httpClient.SetUploadFile(fullPathSourceFileName, stat.Size())

	return httpClient
}

func PathEscapeEncode(v url.Values) string {
	if v == nil {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		keyEscaped := url.PathEscape(k)
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(keyEscaped)
			buf.WriteByte('=')
			buf.WriteString(url.PathEscape(v))
		}
	}
	return buf.String()
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
	argMap := make(map[string]string)
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

func SetGlobalCookie(key string, c []*http.Cookie) error {
	if key == "" {
		return errors.New("key is required")
	}
	gCookies.Store(key, c)
	return nil
}

func GetGlobalCookie(key string) ([]*http.Cookie, error) {
	if key == "" {
		return nil, errors.New("key is required")
	}
	cookies, ok := gCookies.Load(key)
	if !ok {
		return nil, errors.New("gCookies.Load() is failed")
	}
	return cookies.([]*http.Cookie), nil
}
