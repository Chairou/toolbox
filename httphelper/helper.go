package httphelper

import (
	"bytes"
	"encoding/base64"
	"github.com/Chairou/toolbox/util/conv"
	uuid2 "github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"io"
	"k8s.io/klog/v2"
	"net"
	"net/http"
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

	SetDebug(mode int)

	// Do 发送请求
	Do() Result

	error() error
}

const (
	DEBUG_DISABLED = 2
	DEBUG_NORMAL   = 3
	DEBUG_DETAIL   = 4
)

type httpHelper struct {
	body   io.Reader
	client http.Client
	req    *http.Request
	debug  int
	Err    error
}

var Once sync.Once

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

func (p *httpHelper) SetDebug(mode int) {
	if mode == DEBUG_DISABLED || mode == DEBUG_NORMAL || mode == DEBUG_DETAIL {
		p.debug = mode
	}
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
			return result.Errorf("do http request err: %w", err)
		}
	}

	uuid := uuid2.New()
	result.ReqBody = string(byteBody)
	switch p.debug {
	case DEBUG_NORMAL:
		if p.req.Method == "POST" {
			klog.Infoln("HTTP REQUEST:", uuid.String(), p.req.Method, p.req.URL.String(), "\nBODY :", result.ReqBody)
		} else {
			klog.Infoln("HTTP REQUEST:", uuid.String(), p.req.Method, p.req.URL.String())
		}
	case DEBUG_DETAIL:
		if p.req.Method == "POST" {
			klog.Infoln("HTTP REQUEST:", uuid.String(), p.req.Method, p.req.Header, p.req.Cookies(),
				p.req.URL.String(), "\nBODY :", result.ReqBody)
		} else {
			klog.Infoln("HTTP REQUEST:", uuid.String(), p.req.Method, p.req.URL.String())
		}
	}

	newByteBodyReader := bytes.NewReader(byteBody)

	var rc io.ReadCloser
	if newByteBodyReader != nil {
		rc = io.NopCloser(newByteBodyReader)
	}
	p.req.Body = rc
	Once.Do(func() {
		http.DefaultClient.Transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 60 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          60,
			IdleConnTimeout:       60 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxConnsPerHost:       100,
			ResponseHeaderTimeout: 10 * time.Second,
		}
	})

	resp, err := http.DefaultClient.Do(p.req)
	if err != nil {
		//return nil, err
		return result.Errorf("do http request err: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			klog.Errorln("body close err: ", err)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		//return nil, err
		return result.Errorf("read response body err: %w", err)
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
	result.Uuid = uuid.String()
	switch p.debug {
	case DEBUG_NORMAL:
		klog.Infoln("HTTP RESP:", uuid.String(), "\nretBody:", result.RetBody, "elapsed :", elapsed)
	case DEBUG_DETAIL:
		klog.Infoln("HTTP RESP:", uuid.String(), result.RetHeader, result.RetCookie, "\nretBody:", result.RetBody,
			"elapsed :", elapsed)
	}

	return &jsonResult{
		baseResult: result,
		body:       jsoniter.Get(body),
	}
	//},nil
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

// Query 添加Query参数
func (p *errHelper) AddQuery(string, string) Helper {
	return p
}

// QueryMap 通过map来添加Query参数
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

func (p *errHelper) SetDebug(mode int) {
}

// Do 发送请求
//func (p *errHelper) Do() (Result,error) {
//	return &errResult{error: p.error}, p.error
//}

func (p *errHelper) Do() Result {
	return &errResult{error: p.error()}
}
