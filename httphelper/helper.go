package httphelper

import (
	jsoniter "github.com/json-iterator/go"
	"io"
	"net/http"
)

// Helper http Helper接口的提供的方法
type Helper interface {
	// Query 添加Query参数
	Query(k string, v string) Helper

	// QueryMap 通过map来添加Query参数
	QueryMap(m map[string]string) Helper

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

	// Do 发送请求
	Do() Result
}

type httpHelper struct {
	body   io.Reader
	client http.Client
	req    *http.Request
}

// Query 添加Query参数
func (p *httpHelper) Query(k string, v string) Helper {
	query := p.req.URL.Query()
	query.Add(k, v)
	p.req.URL.RawQuery = query.Encode()
	return p
}

// QueryMap 通过map来添加Query参数
func (p *httpHelper) QueryMap(m map[string]string) Helper {
	query := p.req.URL.Query()
	for k, v := range m {
		query.Add(k, v)
	}
	p.req.URL.RawQuery = query.Encode()
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

// Do 发送请求
func (p *httpHelper) Do() Result {
	result := &baseResult{}
	resp, err := http.DefaultClient.Do(p.req)
	if err != nil {
		//return nil, err
		return result.errorf("do http request err: %w", err)
	}
	result.Status = resp.StatusCode
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		//return nil, err
		return result.errorf("read response body err: %w", err)
	}
	result.Body = string(body)
	Log(p.req.Method, p.req.URL.String(), string(body))
	return &jsonResult{
		baseResult: result,
		body:       jsoniter.Get(body),
	}
	//},nil
}

type errHelper struct {
	error
}

func errorHelper(err error) Helper {
	return &errHelper{error: err}
}

// Query 添加Query参数
func (p *errHelper) Query(string, string) Helper {
	return p
}

// QueryMap 通过map来添加Query参数
func (p *errHelper) QueryMap(map[string]string) Helper {
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

// SetTransport 设置RoundTripper，可用于添加中间件
func (p *errHelper) SetTransport(http.RoundTripper) Helper {
	return p
}

// Do 发送请求
//func (p *errHelper) Do() (Result,error) {
//	return &errResult{error: p.error}, p.error
//}

func (p *errHelper) Do() Result {
	return &errResult{error: p.error}
}
