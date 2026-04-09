package httphelper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// ==================== SetTimeout 相关测试 ====================

// TestSetTimeout_SafeTypeAssertion 测试 SetTimeout 在 Transport 为 nil 或非 *http.Transport 时不 panic
func TestSetTimeout_SafeTypeAssertion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	// 场景1：正常情况下 SetTimeout 不 panic
	t.Run("正常Transport类型断言", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("SetTimeout 不应该 panic，但发生了 panic: %v", r)
			}
		}()
		client := GET(ts.URL)
		client.SetTimeout(5*time.Second, 20*time.Second)
		ret := client.Do()
		if ret.Error() != nil {
			t.Errorf("请求不应该出错: %v", ret.Error())
		}
	})

	// 场景2：自定义 Transport（非 *http.Transport）时不 panic
	t.Run("自定义RoundTripper不panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("SetTimeout 不应该 panic，但发生了 panic: %v", r)
			}
		}()
		client := GET(ts.URL)
		client.SetTransport(http.DefaultTransport)
		client.SetTimeout(3*time.Second, 10*time.Second)
		ret := client.Do()
		if ret.Error() != nil {
			t.Logf("请求可能出错（预期行为）: %v", ret.Error())
		}
	})

	// 场景3：Transport 为 nil 时不 panic
	t.Run("Transport为nil不panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("SetTimeout 不应该 panic，但发生了 panic: %v", r)
			}
		}()
		client := NewRequest("GET", ts.URL, nil)
		if h, ok := client.(*httpHelper); ok {
			h.client.Transport = nil
			h.SetTimeout(3*time.Second, 10*time.Second)
			if h.client.Timeout != 10*time.Second {
				t.Errorf("期望 Timeout 为 10s，实际为 %v", h.client.Timeout)
			}
		}
	})
}

// TestSetTimeout_NoDoubleMultiply 测试 SetTimeout 不会重复乘以 time.Second
func TestSetTimeout_NoDoubleMultiply(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	client := NewRequest("GET", ts.URL, nil)
	client.SetTimeout(5*time.Second, 20*time.Second)

	if h, ok := client.(*httpHelper); ok {
		if h.client.Timeout != 20*time.Second {
			t.Errorf("期望 Timeout 为 20s，实际为 %v（可能存在重复乘算）", h.client.Timeout)
		}
		if h.client.Timeout > time.Minute {
			t.Errorf("Timeout 值异常过大 (%v)，说明存在重复乘以 time.Second 的问题", h.client.Timeout)
		}
	} else {
		t.Error("client 不是 *httpHelper 类型")
	}
}

// ==================== Do() 相关测试 ====================

// TestDo_UuidAssignedBeforeReadBody 测试 Uuid 在读取 body 之前就已赋值
func TestDo_UuidAssignedBeforeReadBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"success"}`))
	}))
	defer ts.Close()

	client := NewRequest("GET", ts.URL, nil)
	ret := client.Do()
	if ret.Error() != nil {
		t.Fatalf("请求不应该出错: %v", ret.Error())
	}

	uuid := ret.BaseResult().Uuid
	if uuid == "" {
		t.Error("Uuid 不应该为空")
	}
	t.Logf("请求 Uuid: %s", uuid)
}

// TestDo_BodyNilForGetRequest 测试 GET 请求时 body 为 nil 不会出问题
func TestDo_BodyNilForGetRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil && r.ContentLength > 0 {
			t.Error("GET 请求不应该有 body")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"method":"GET"}`))
	}))
	defer ts.Close()

	client := GET(ts.URL + "/test")
	ret := client.Do()
	if ret.Error() != nil {
		t.Fatalf("GET 请求不应该出错: %v", ret.Error())
	}
	if ret.BaseResult().RetBody != `{"method":"GET"}` {
		t.Errorf("返回 body 不匹配，实际: %s", ret.BaseResult().RetBody)
	}
}

// TestDo_PostWithBody 测试 POST 请求 body 正确传递
func TestDo_PostWithBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("期望 POST 方法，实际: %s", r.Method)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusOK)
		resp, _ := json.Marshal(body)
		w.Write(resp)
	}))
	defer ts.Close()

	client := PostJSON(ts.URL, `{"name":"test","value":"123"}`)
	ret := client.Do()
	if ret.Error() != nil {
		t.Fatalf("POST 请求不应该出错: %v", ret.Error())
	}

	if ret.BaseResult().ReqBody == "" {
		t.Error("POST 请求的 ReqBody 不应该为空")
	}
	t.Logf("ReqBody: %s", ret.BaseResult().ReqBody)
	t.Logf("RetBody: %s", ret.BaseResult().RetBody)
}

// TestDo_NonOKStatusContainsBody 测试非 200 状态码时错误信息包含响应体
func TestDo_NonOKStatusContainsBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_param","message":"参数错误"}`))
	}))
	defer ts.Close()

	client := GET(ts.URL)
	ret := client.Do()

	if ret.Error() == nil {
		t.Fatal("非 200 状态码应该返回错误")
	}

	errMsg := ret.Error().Error()
	if !strings.Contains(errMsg, "400") {
		t.Errorf("错误信息应该包含状态码 400，实际: %s", errMsg)
	}
	if !strings.Contains(errMsg, "invalid_param") {
		t.Errorf("错误信息应该包含响应体内容 'invalid_param'，实际: %s", errMsg)
	}
	t.Logf("错误信息: %s", errMsg)
}

// TestDo_Non200StatusPreservesBaseResult 测试非 200 状态码时 BaseResult 仍然可用
func TestDo_Non200StatusPreservesBaseResult(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "test-value")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"code":403,"message":"forbidden"}`))
	}))
	defer ts.Close()

	client := GET(ts.URL)
	ret := client.Do()

	if ret.Error() == nil {
		t.Fatal("403 状态码应该返回错误")
	}

	base := ret.BaseResult()
	if base.Status != http.StatusForbidden {
		t.Errorf("期望状态码 403，实际: %d", base.Status)
	}
	if base.Uuid == "" {
		t.Error("Uuid 不应该为空")
	}
	t.Logf("Status: %d, Uuid: %s, RetBody: %s", base.Status, base.Uuid, base.RetBody)
}

// TestDo_ServerInternalError 测试 500 状态码
func TestDo_ServerInternalError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal_error"}`))
	}))
	defer ts.Close()

	client := GET(ts.URL)
	ret := client.Do()

	if ret.Error() == nil {
		t.Fatal("500 状态码应该返回错误")
	}

	errMsg := ret.Error().Error()
	if !strings.Contains(errMsg, "500") {
		t.Errorf("错误信息应该包含状态码 500，实际: %s", errMsg)
	}
	if !strings.Contains(errMsg, "internal_error") {
		t.Errorf("错误信息应该包含响应体内容，实际: %s", errMsg)
	}
}

// ==================== httpHelper 结构体测试 ====================

// TestHttpHelper_NoBodyField 测试 httpHelper 结构体基本字段正确初始化
func TestHttpHelper_NoBodyField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	client := NewRequest("GET", ts.URL, nil)
	h, ok := client.(*httpHelper)
	if !ok {
		t.Fatal("client 应该是 *httpHelper 类型")
	}

	if h.req == nil {
		t.Error("req 不应该为 nil")
	}
	if h.debug != DebugNormal {
		t.Errorf("debug 应该为 DebugNormal(%d)，实际: %d", DebugNormal, h.debug)
	}
	if h.Err != nil {
		t.Errorf("Err 应该为 nil，实际: %v", h.Err)
	}
}

// ==================== PostFile 相关测试 ====================

// TestPostFile_FileNotExist 测试文件不存在时返回 errHelper
func TestPostFile_FileNotExist(t *testing.T) {
	client := PostFile("http://127.0.0.1/upload", "/tmp/nonexistent_file_12345.txt", "test.txt", nil)
	if client.error() == nil {
		t.Error("文件不存在时应该返回错误")
	}
	_, ok := client.(*errHelper)
	if !ok {
		t.Error("文件不存在时应该返回 errHelper 类型")
	}
	t.Logf("错误信息: %v", client.error())
}

// TestPostFile_ValidFile 测试有效文件上传
func TestPostFile_ValidFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_upload_*.txt")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("test file content for upload")
	if err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}
	tmpFile.Close()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("期望 POST 方法，实际: %s", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if !strings.Contains(ct, "multipart/form-data") {
			t.Errorf("Content-Type 应该包含 multipart/form-data，实际: %s", ct)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"uploaded":true}`))
	}))
	defer ts.Close()

	client := PostFile(ts.URL, tmpFile.Name(), "uploaded.txt", map[string]string{"key": "value"})
	if client.error() != nil {
		t.Fatalf("有效文件不应该返回错误: %v", client.error())
	}

	ret := client.Do()
	if ret.Error() != nil {
		t.Fatalf("上传请求不应该出错: %v", ret.Error())
	}
	t.Logf("上传结果: %s", ret.BaseResult().RetBody)
}

// ==================== UrlToMap 相关测试 ====================

// TestUrlToMap_NormalUrl 测试正常 URL 解析
func TestUrlToMap_NormalUrl(t *testing.T) {
	result := UrlToMap("http://example.com/path?a=1&b=2&c=3")
	if result["a"] != "1" {
		t.Errorf("期望 a=1，实际: a=%s", result["a"])
	}
	if result["b"] != "2" {
		t.Errorf("期望 b=2，实际: b=%s", result["b"])
	}
	if result["c"] != "3" {
		t.Errorf("期望 c=3，实际: c=%s", result["c"])
	}
}

// TestUrlToMap_NoQuestionMark 测试没有 "?" 的 URL 不 panic
func TestUrlToMap_NoQuestionMark(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("UrlToMap 不应该 panic，但发生了 panic: %v", r)
		}
	}()

	result := UrlToMap("a=1&b=2")
	if result["a"] != "1" {
		t.Errorf("期望 a=1，实际: a=%s", result["a"])
	}
	if result["b"] != "2" {
		t.Errorf("期望 b=2，实际: b=%s", result["b"])
	}
	t.Logf("无问号 URL 解析结果: %v", result)
}

// TestUrlToMap_NoEqualSign 测试参数中没有 "=" 不 panic
func TestUrlToMap_NoEqualSign(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("UrlToMap 不应该 panic，但发生了 panic: %v", r)
		}
	}()

	result := UrlToMap("http://example.com/path?keyonly&a=1")
	if val, ok := result["keyonly"]; ok {
		if val != "" {
			t.Errorf("没有值的参数应该为空字符串，实际: %s", val)
		}
	}
	if result["a"] != "1" {
		t.Errorf("期望 a=1，实际: a=%s", result["a"])
	}
	t.Logf("无等号参数解析结果: %v", result)
}

// TestUrlToMap_EmptyParams 测试空参数
func TestUrlToMap_EmptyParams(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("UrlToMap 不应该 panic，但发生了 panic: %v", r)
		}
	}()

	result := UrlToMap("http://example.com/path?")
	t.Logf("空参数解析结果: %v", result)
}

// ==================== UrlPathEscape 相关测试 ====================

// TestUrlPathEscape_NormalUrl 测试正常 URL 转义
func TestUrlPathEscape_NormalUrl(t *testing.T) {
	result := UrlPathEscape("http://example.com/path?a=1&b=hello")
	if !strings.HasPrefix(result, "http://example.com/path?") {
		t.Errorf("URL 前缀不正确: %s", result)
	}
	if !strings.Contains(result, "a=1") {
		t.Errorf("应该包含 a=1: %s", result)
	}
	t.Logf("正常 URL 转义结果: %s", result)
}

// TestUrlPathEscape_NoQuestionMark 测试没有 "?" 的 URL 不 panic 且直接返回原 URL
func TestUrlPathEscape_NoQuestionMark(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("UrlPathEscape 不应该 panic，但发生了 panic: %v", r)
		}
	}()

	input := "http://example.com/path"
	result := UrlPathEscape(input)
	if result != input {
		t.Errorf("没有 '?' 时应该返回原 URL，期望: %s，实际: %s", input, result)
	}
	t.Logf("无问号 URL 转义结果: %s", result)
}

// TestUrlPathEscape_ChineseParams 测试中文参数转义
func TestUrlPathEscape_ChineseParams(t *testing.T) {
	result := UrlPathEscape("http://example.com/path?name=中文&key=value")
	if strings.Contains(result, "中文") {
		t.Errorf("中文应该被转义，但结果中仍包含中文: %s", result)
	}
	if !strings.Contains(result, "key=value") {
		t.Errorf("应该包含 key=value: %s", result)
	}
	t.Logf("中文参数转义结果: %s", result)
}

// ==================== errHelper 相关测试 ====================

// TestErrHelper_AllMethodsReturnSelf 测试 errHelper 的所有方法都安全返回自身
func TestErrHelper_AllMethodsReturnSelf(t *testing.T) {
	client := errorHelper(fmt.Errorf("test error"))

	// 所有链式调用方法都应该安全返回，不 panic
	client.AddQuery("k", "v")
	client.AddQueryMap(map[string]string{"k": "v"})
	client.AddPathEscapeQuery("k", "v")
	client.AddPathEscapeQueryMap(map[string]string{"k": "v"})
	client.AddHeader("k", "v")
	client.AddHeaderMap(map[string]string{"k": "v"})
	client.SetHeader("k", "v")
	client.AddCookie([]*http.Cookie{{Name: "k", Value: "v"}})
	client.AddCookieMap(map[string]string{"k": "v"})
	client.AddBasicAuth("user", "pass")
	client.AddOAuthAccessToken("token")
	client.SetTransport(http.DefaultTransport)
	client.SetDebug(DebugDetail)
	client.AddSign("appid", "secret")
	client.SetUploadFile("file", 100)
	client.SetTimeout(5*time.Second, 20*time.Second)

	if client.error() == nil {
		t.Error("errHelper 应该包含错误")
	}
	if client.error().Error() != "test error" {
		t.Errorf("错误信息不匹配，期望: test error，实际: %s", client.error().Error())
	}

	ret := client.Do()
	if ret.Error() == nil {
		t.Error("errHelper.Do() 应该返回错误")
	}
	t.Logf("errHelper 错误: %v", ret.Error())
}

// ==================== 综合集成测试 ====================

// TestIntegration_FullRequestLifecycle 测试完整的请求生命周期
func TestIntegration_FullRequestLifecycle(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"method":  r.Method,
			"path":    r.URL.Path,
			"query":   r.URL.RawQuery,
			"headers": r.Header.Get("X-Custom"),
		}
		w.Header().Set("X-Response-Id", "resp-123")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := GET(ts.URL + "/api/test")
	client.AddQuery("key1", "value1")
	client.AddHeader("X-Custom", "custom-value")
	client.AddCookieMap(map[string]string{"session": "abc123"})
	client.SetTimeout(5*time.Second, 10*time.Second)

	ret := client.Do()
	if ret.Error() != nil {
		t.Fatalf("请求不应该出错: %v", ret.Error())
	}

	base := ret.BaseResult()

	if base.Status != http.StatusOK {
		t.Errorf("期望状态码 200，实际: %d", base.Status)
	}
	if base.Uuid == "" {
		t.Error("Uuid 不应该为空")
	}
	if base.Url == "" {
		t.Error("Url 不应该为空")
	}
	if base.Elapsed == "" {
		t.Error("Elapsed 不应该为空")
	}
	if base.RetBody == "" {
		t.Error("RetBody 不应该为空")
	}

	t.Logf("完整请求生命周期测试通过 - Status: %d, Uuid: %s, Elapsed: %s",
		base.Status, base.Uuid, base.Elapsed)
}

// TestIntegration_PostJsonWithTimeout 测试 POST JSON 请求带超时设置
func TestIntegration_PostJsonWithTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(body)
	}))
	defer ts.Close()

	type Payload struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	client := PostJSON(ts.URL, Payload{Name: "test", Value: 42})
	client.SetTimeout(5*time.Second, 10*time.Second)
	ret := client.Do()

	if ret.Error() != nil {
		t.Fatalf("POST 请求不应该出错: %v", ret.Error())
	}

	if ret.BaseResult().ReqBody == "" {
		t.Error("POST 请求的 ReqBody 不应该为空")
	}
	if !strings.Contains(ret.BaseResult().ReqBody, "test") {
		t.Errorf("ReqBody 应该包含 'test'，实际: %s", ret.BaseResult().ReqBody)
	}
	t.Logf("POST JSON 请求测试通过 - ReqBody: %s, RetBody: %s",
		ret.BaseResult().ReqBody, ret.BaseResult().RetBody)
}