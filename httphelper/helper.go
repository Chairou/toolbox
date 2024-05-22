package httphelper

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/Chairou/toolbox/logger"
	"github.com/Chairou/toolbox/util/color"
	"github.com/Chairou/toolbox/util/conv"
	uuid2 "github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// Helper http Helper接口的提供的方法
type Helper interface {
	// AddQuery 添加Query参数
	AddQuery(k string, v string) Helper

	// AddQueryMap 通过map来添加Query参数
	AddQueryMap(m map[string]string) Helper

	// AddPathEscapeQuery 添加Query参数,并进行PathEscape转义
	AddPathEscapeQuery(k string, v string) Helper

	AddPathEscapeQueryMap(m map[string]string) Helper
	// AddHeader 添加HTTP header
	AddHeader(k string, v string) Helper

	// SetHeader 设置HTTP header，会覆盖同名的header
	SetHeader(k string, v string) Helper

	// SetTransport 设置RoundTripper，可用于添加中间件
	SetTransport(tripper http.RoundTripper) Helper

	// AddCookies 设置完整cookies
	AddCookies(c []*http.Cookie) Helper

	// AddSimpleCookies 设置只有简单name和value的cookies
	AddSimpleCookies(c map[string]string) Helper

	// AddBasicAuth 设置简单认证
	AddBasicAuth(id string, key string) Helper

	// AddOAuthAccessToken 设置OAuth认证
	AddOAuthAccessToken(token string) Helper

	SetDebug(mode int) Helper

	// AddSign 添加datamore_api的签名方法
	AddSign(appID string, appSecret string) Helper

	// Do 发送请求
	Do() Result

	error() error

	SetUploadFile(fileName string, fileSize int64) Helper
}

const (
	DebugDisabled = 2
	DebugNormal   = 3
	DebugDetail   = 4
	DebugUpload   = 5
)

type httpHelper struct {
	body           io.Reader
	client         http.Client
	req            *http.Request
	debug          int
	Err            error
	UploadFileName string
	UploadFileSize int64
	Uuid           string
}

var Once sync.Once
var log *logger.LogPool

func init() {
	Once.Do(func() {
		logFileName := os.Getenv("httpLogFileName")
		if logFileName == "" {
			logFileName = "/tmp/http.log"
		}
		var err error
		log, err = logger.NewLogPool("http", logFileName)
		if err != nil {
			fmt.Println("Error creating log:", err.Error())
			return
		}
		log.SetLevel(logger.DEBUG_LEVEL)
		log.Infoln("http log init.")
	})
}

// AddQuery 添加Query参数
func (p *httpHelper) AddQuery(k string, v string) Helper {
	query := p.req.URL.Query()
	query.Add(k, v)
	p.req.URL.RawQuery = query.Encode()
	return p
}

// AddQueryMap 通过map来添加Query参数
func (p *httpHelper) AddQueryMap(m map[string]string) Helper {
	query := p.req.URL.Query()
	for k, v := range m {
		query.Add(k, v)
	}
	p.req.URL.RawQuery = query.Encode()
	return p
}

// AddPathEscapeQuery 增加query的kv, kv均用PathEscape转义
func (p *httpHelper) AddPathEscapeQuery(k string, v string) Helper {
	query := p.req.URL.Query()
	query.Add(k, v)
	p.req.URL.RawQuery = PathEscapeEncode(query)
	return p
}

// AddPathEscapeQueryMap 通过map来添加Query参数并进行PathEscape 转义
func (p *httpHelper) AddPathEscapeQueryMap(m map[string]string) Helper {
	query := p.req.URL.Query()
	for k, v := range m {
		query.Add(k, v)
	}
	p.req.URL.RawQuery = PathEscapeEncode(query)
	return p
}

// AddHeader 添加HTTP header
func (p *httpHelper) AddHeader(k string, v string) Helper {
	p.req.Header.Add(k, v)
	return p
}

// SetHeader 设置HTTP header，会覆盖同名的header
func (p *httpHelper) SetHeader(k string, v string) Helper {
	p.req.Header.Set(k, v)
	return p
}

// SetTransport 设置RoundTripper，可用于添加中间件
func (p *httpHelper) SetTransport(tripper http.RoundTripper) Helper {
	p.client.Transport = tripper
	return p
}

func (p *httpHelper) AddCookies(c []*http.Cookie) Helper {
	for _, v := range c {
		p.req.AddCookie(v)
	}
	return p
}

func (p *httpHelper) AddSimpleCookies(c map[string]string) Helper {
	for k, v := range c {
		p.req.AddCookie(&http.Cookie{Name: k, Value: v})
	}
	return p
}

func (p *httpHelper) AddBasicAuth(username string, password string) Helper {
	// 设置 Basic 认证头
	auth := username + ":" + password
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	p.req.Header.Add("Authorization", basicAuth)
	return p
}

func (p *httpHelper) AddOAuthAccessToken(token string) Helper {
	p.req.Header.Add("Authorization", "Bearer "+token)
	return p
}

func (p *httpHelper) AddSign(appID string, appSecret string) Helper {
	timestamp := strconv.Itoa((int)(time.Now().Unix()))
	random := fmt.Sprintf("%v%v%v", rand.Uint64(), rand.Uint64(), rand.Uint64())
	//appID := "tesing"
	//secert := "testing-1234-1234"

	tokenBytes := md5.Sum([]byte(fmt.Sprintf("%v|%s|%v|!&&@@%%#$!*^|%v", appID, timestamp, appSecret, random)))
	token := hex.EncodeToString(tokenBytes[:])

	p.req.Header.Add("Timestamp", timestamp)
	p.req.Header.Add("Random", random)
	p.req.Header.Add("App-Id", appID)
	p.req.Header.Add("Access-Token", token)
	return p
}

func (p *httpHelper) SetDebug(mode int) Helper {
	if mode >= DebugDisabled && mode <= DebugUpload {
		p.debug = mode
	}
	return p
}

func (p *httpHelper) SetUploadFile(fileName string, fileSize int64) Helper {
	p.UploadFileName = fileName
	p.UploadFileSize = fileSize
	return p
}

// Do 发送请求
func (p *httpHelper) Do() Result {
	startTime := time.Now()
	result := &baseResult{}
	var byteBody []byte
	if p.req.Body != nil {
		var err error
		byteBody, err = io.ReadAll(p.req.Body)
		if err != nil {
			redStr := color.SetColor(color.Red, fmt.Sprintf("%v", err))
			s := fmt.Sprintf("%s  io.ReadAll err: %s", p.Uuid, redStr)
			log.Errorln(s)
			return result.Errorf("io.ReadAll err: %v", err)
		}
	}

	p.Uuid = uuid2.NewString()
	result.ReqBody = string(byteBody)
	switch p.debug {
	case DebugNormal:
		if p.req.Method == "POST" {
			log.Infoln("HTTP REQ:", p.Uuid, p.req.Method, p.req.URL.String(), "\n【reqBODY】:",
				color.SetColor(color.Green, result.ReqBody))
		} else {
			log.Infoln("HTTP REQ:", p.Uuid, p.req.Method, color.SetColor(color.Green, p.req.URL.String()))
		}
	case DebugDetail:
		if p.req.Method == "POST" {
			log.Infoln("HTTP REQ:", p.Uuid, p.req.Method, p.req.URL.String(), p.req.Header, p.req.Cookies(),
				"\n【reqBODY】 :", color.SetColor(color.Green, result.ReqBody))
		} else {
			log.Infoln("HTTP REQ:", p.Uuid, p.req.Method, color.SetColor(color.Green, p.req.URL.String()))
		}
	case DebugUpload:
		log.Infoln("HTTP UPLOAD FILE:", p.Uuid, p.req.Method, p.req.URL.String(), ", fileName:",
			p.UploadFileName, ", fileSize:", p.UploadFileSize)
	}

	newByteBodyReader := bytes.NewReader(byteBody)

	var rc io.ReadCloser
	if newByteBodyReader != nil {
		rc = io.NopCloser(newByteBodyReader)
	}
	p.req.Body = rc

	resp, err := http.DefaultClient.Do(p.req)
	if err != nil {
		redStr := color.SetColor(color.Red, fmt.Sprintf("%v", err))
		s := fmt.Sprintf("%s do http request err: %s", p.Uuid, redStr)
		log.Errorln(s)
		return result.Errorf("do http request err: %v", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			redStr := color.SetColor(color.Red, fmt.Sprintf("%v", err))
			s := fmt.Sprintf("%s body close err: %s", p.Uuid, redStr)
			log.Errorln(s)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		redStr := color.SetColor(color.Red, fmt.Sprintf("%v", err))
		s := fmt.Sprintf("%s read response body err: %s", p.Uuid, redStr)
		log.Errorln(s)
		return result.Errorf("read response body err: %v", err)
	}

	result.Status = resp.StatusCode
	result.Url = p.req.URL.String()
	result.ReqHeader = p.req.Header
	result.ReqCookie = p.req.Cookies()
	result.RetHeader = resp.Header
	result.RetCookie = resp.Cookies()
	result.RetBody = string(body)
	elapsed := time.Since(startTime)
	result.Elapsed = conv.String(elapsed)
	result.BodyLen = len(body)
	result.Uuid = p.Uuid
	switch p.debug {
	case DebugNormal:
		log.Infoln("HTTP RESP:", p.Uuid, "\n【retBody】:", color.SetColor(color.Green, result.RetBody),
			"elapsed :", elapsed)
	case DebugDetail:
		log.Infoln("HTTP RESP:", p.Uuid, "\n【retBody】:", color.SetColor(color.Green, result.RetBody),
			result.RetHeader, result.RetCookie, "elapsed :", elapsed)
	}

	return &jsonResult{
		baseResult: result,
		body:       jsoniter.Get(body),
	}
}

func (p *httpHelper) error() error {
	return p.Err
}

type errHelper struct {
	Err error
}

func errorHelper(err error) Helper {
	return &errHelper{Err: err}
}

// Query 添加Query参数
func (p *errHelper) error() error {
	return p.Err
}

// AddQuery 添加Query参数
func (p *errHelper) AddQuery(string, string) Helper {
	return p
}

// AddQueryMap 通过map来添加Query参数
func (p *errHelper) AddQueryMap(map[string]string) Helper {
	return p
}

func (p *errHelper) AddPathEscapeQuery(k string, v string) Helper {
	return p
}

func (p *errHelper) AddPathEscapeQueryMap(m map[string]string) Helper {
	return p
}

// AddHeader 添加HTTP header
func (p *errHelper) AddHeader(string, string) Helper {
	return p
}

// SetHeader 设置HTTP header，会覆盖同名的header
func (p *errHelper) SetHeader(string, string) Helper {
	return p
}

func (p *errHelper) AddCookies(c []*http.Cookie) Helper {
	return p
}

func (p *errHelper) AddSimpleCookies(c map[string]string) Helper {
	return p
}

func (p *errHelper) AddBasicAuth(username string, password string) Helper {
	return p
}

func (p *errHelper) AddOAuthAccessToken(token string) Helper {
	return p
}

// SetTransport 设置RoundTripper，可用于添加中间件
func (p *errHelper) SetTransport(http.RoundTripper) Helper {
	return p
}

func (p *errHelper) SetDebug(mode int) Helper { return p }

func (p *errHelper) AddSign(appID string, appSecret string) Helper { return p }

// Do 发送请求
//func (p *errHelper) Do() (Result,error) {
//	return &errResult{error: p.error}, p.error
//}

func (p *errHelper) Do() Result {
	return &errResult{error: p.error()}
}

func (p *errHelper) SetUploadFile(fileName string, fileSize int64) Helper { return p }
