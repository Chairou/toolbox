package httphelper

import (
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

// NilValueErr 空值错误
var NilValueErr = errors.New("nil value")

// Result Http请求的结果
type Result interface {
	// BaseResult 返回Http请求的基本结果，包含Status和Body
	BaseResult() *baseResult

	// StdCheck 检查HTTP响应码和Body中的Code
	StdCheck() Result

	// Bind 将返回值存储到Object中
	Bind(object interface{}, path ...interface{}) error

	// Get 通过json path获取子结果
	Get(path ...interface{}) Result

	// Error 返回出现的错误
	Error() error

	// Jsoniter 返回body的json迭代器
	Jsoniter() jsoniter.Any

	// UnmarshalFromBody json UnmarshalFromBody
	UnmarshalFromBody(v interface{}) error
}

type baseResult struct {
	Status    int
	Url       string
	ReqHeader http.Header
	ReqCookie []*http.Cookie
	ReqBody   string
	RetHeader http.Header
	RetCookie []*http.Cookie
	RetBody   string
	Elapsed   string
	BodyLen   int
	Uuid      string
	Err       error
}

// BaseResult 返回Http请求的基本结果，包含Status和Body
func (p *baseResult) BaseResult() *baseResult {
	return p
}

// Error 返回出现的错误
func (p *baseResult) Error() error {
	return p.Err
}

func (p *baseResult) error(err error) *errResult {
	return &errResult{
		baseResult: p,
		error:      err,
	}
}

func (p *baseResult) errorf(format string, a ...interface{}) *errResult {
	return p.error(fmt.Errorf(format, a...))
}

type jsonResult struct {
	*baseResult
	body jsoniter.Any
}

// Jsoniter 返回body的json迭代器
func (p *jsonResult) Jsoniter() jsoniter.Any {
	return p.body
}

// Get 通过json path获取子结果
func (p *jsonResult) Get(path ...interface{}) Result {
	return &jsonResult{
		baseResult: p.baseResult,
		body:       jsoniter.Get([]byte(p.RetBody), path...),
	}
}

// StdCheck 检查HTTP响应码和Body中的Code
func (p *jsonResult) StdCheck() Result {
	if p.Status != http.StatusOK {
		return p.error(fmt.Errorf("response status:%d", p.Status))
	}
	resp := jsoniter.Get([]byte(p.RetBody))
	if code := resp.Get("code").ToInt(); code != 0 {
		return p.error(fmt.Errorf("resp code:%d, meeesage:%s", code, resp.Get("message").ToString()))
	}
	return p
}

// Bind 将返回值存储到Object中
func (p *jsonResult) Bind(object interface{}, path ...interface{}) error {
	res := jsoniter.Get([]byte(p.RetBody), path...)
	if err := res.LastError(); err != nil {
		return fmt.Errorf("parse response body err:%w, body:%s", err, p.RetBody)
	}
	if res.ValueType() == jsoniter.NilValue {
		return NilValueErr
	}
	res.ToVal(object)
	return nil
}

func (p *jsonResult) UnmarshalFromBody(v interface{}) error {
	if err := jsoniter.UnmarshalFromString(p.RetBody, v); err != nil {
		return err
	}
	return nil
}

type errResult struct {
	*baseResult
	error
}

// StdCheck 检查HTTP响应码和Body中的Code
func (p *errResult) StdCheck() Result {
	return p
}

// Bind 将返回值存储到Object中
func (p *errResult) Bind(interface{}, ...interface{}) error {
	return p.error
}

// Get 通过json path获取子结果
func (p *errResult) Get(...interface{}) Result {
	return p
}

// Error 返回出现的错误
func (p *errResult) Error() error {
	return p.error
}

// Jsoniter 返回body的json迭代器
func (p *errResult) Jsoniter() jsoniter.Any {
	return jsoniter.Wrap(nil)
}

// UnmarshalFromBody 返回出现的错误
func (p *errResult) UnmarshalFromBody(v interface{}) error {
	return p.error
}
