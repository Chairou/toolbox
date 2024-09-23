package httphelper

//go test -v http_test.go http.go helper.go

import (
	"encoding/base64"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"net/http"
	"net/url"

	"testing"
	"time"
)

type Ret struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    TmpPost `json:"data"`
}
type TmpPost struct {
	Sex     string            `json:"sex"`
	Age     int               `json:"age"`
	Pairent map[string]string `json:"pairent"`
	Address []string          `json:"address"`
	Name    string            `json:"name"`
}

func TestUrlToMap(t *testing.T) {
	url1 := "l5://2342343:324234/main?s=3&g=ff"
	url2 := "s=3&g=ff"
	url3 := "https://news.qq.com:8080/main?s=3&g=中文去"

	t.Logf("%s\n", UrlPathEscape(url1))
	t.Logf("%s\n", UrlPathEscape(url2))
	t.Logf("%s\n", UrlPathEscape(url3))
}

var time1 time.Time = time.Now()

func TestPostFuncHttp(t *testing.T) {
	type MyType func(int) int
	var f MyType = func(x int) int { return x * x }
	client := PostJSON("http://127.0.0.1/postBody", f)
	if client.error() != nil {
		t.Log(client.error())
	}
	val, ok := client.(*errHelper)
	if ok {
		t.Log(*val)
	} else {
		t.Error("expected errHelper")
	}
}

func TestPostString(t *testing.T) {
	client := PostJSON("http://127.0.0.1/postBody", `{"name":"win"}`)
	client.SetDebug(DebugDetail)
	client.AddHeader("aa", "bb").AddHeader("cc", "dd")
	client.AddSimpleCookies(map[string]string{"ee": "ff"})
	ret := client.Do()
	asd := ret.Get("name")
	q := asd.JsonIterAny().ToString()
	t.Log(q)
}

func TestPostSturctInstance(t *testing.T) {
	type Simple struct {
		Name string `json:"name"`
	}
	var instance Simple
	instance.Name = "AAAA"
	client := PostJSON("http://127.0.0.1/postBody", instance)
	ret := client.Do()
	instance.Name = "BBBB"
	err := ret.Bind(&instance)
	if err != nil {
		t.Error(err)
	}
	t.Logf("simple : %#v", instance)
}

func TestPostJsonRet(t *testing.T) {
	type Simple struct {
		Name string `json:"name"`
	}
	req := Simple{Name: "asd"}
	ret := Simple{}
	err := PostJsonRet("http://127.0.0.1/api/postBody", req, &ret)
	if err != nil {
		t.Error(err)
	}
	t.Logf("simple : %#v", ret)
}

func TestUrlPathEscape(t *testing.T) {
	url1 := "http://news.qq.com/?a=1&b=吃饭"
	urlEncode := UrlPathEscape(url1)
	t.Log(urlEncode)
}

func TestGetHttp(t *testing.T) {
	url1 := "http://127.0.0.1/get?aa=bb&cc=dd"
	client := GET(url1)
	client.AddQuery("key1", "value1")
	client.AddHeader("envSelector", "test")
	cookies := make([]*http.Cookie, 0)
	c1 := &http.Cookie{Name: "ee", Value: "ff"}
	c2 := &http.Cookie{Name: "gg", Value: "hh"}
	cookies = append(cookies, c1, c2)
	client.AddCookies(cookies)
	ret := client.Do()
	// 显示body内容
	fmt.Println(ret.BaseResult().RetBody)
}

func TestPathEscaped(t *testing.T) {
	url1 := "http://127.0.0.1/get?aa=bb&cc=dd"
	client := GET(url1)
	client.AddPathEscapeQuery("key2", "value2")
	client.AddPathEscapeQuery("key3", "中文")
	ret := client.Do()
	// 显示body内容
	t.Log(ret.BaseResult().RetBody)
	t.Log(ret.BaseResult().ReqBody)
}

func TestJsonIter(t *testing.T) {
	aaa := baseResult{}
	aaa.RetBody = "{\"code\":0,\"message\":\"hello\",\"data\":{}}"
	bbb := jsonResult{
		baseResult: &aaa,
		body:       jsoniter.Get([]byte{}),
	}
	ret := &Ret{}
	err := bbb.UnmarshalFromBody(ret)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(ret)
}

func TestGlobalCookie(t *testing.T) {
	cookie1 := http.Cookie{}
	cookie1.Name = "auth"
	cookie1.Value = "chair"
	cookie2 := http.Cookie{}
	cookie2.Name = "passwd"
	cookie2.Value = "bbb"
	cookies := make([]*http.Cookie, 0)
	cookies = append(cookies, &cookie1)
	cookies = append(cookies, &cookie2)
	err := SetGlobalCookie("test1", cookies)
	if err != nil {
		t.Error(err)
	}
	cookieback, err := GetGlobalCookie("test1")
	if err != nil {
		t.Error("GetGlobalCookie failed :", err)
	}
	t.Log(cookieback[0].Name)
}

func TestPostUrlEncode(t *testing.T) {
	url1 := "http://127.0.0.1/get?aa=bb&cc=dd"
	data := url.Values{}
	data.Set("workspace_id", "123")
	data.Set("name", "test1")
	data.Set("description", "test all method")
	client := PostUrlEncode(url1, data)
	// 设置 Basic 认证头
	username := "nofight"   // 替换为您的用户名
	password := "F-4-1-E-A" // 替换为您的密码
	auth := username + ":" + password
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	client.AddHeader("Authorization", basicAuth)
	ret := client.Do()
	if ret.Error() != nil {
		t.Log("sendRtxOversea err: ", ret.Error())
	}
	t.Log(ret.BaseResult().RetBody)
	t.Log(ret.BaseResult().ReqBody)
}

func TestBasicAuth(t *testing.T) {
	url1 := "http://127.0.0.1/get?aa=bb&cc=dd"
	username := "nofight"   // 替换为您的用户名
	password := "F-4-1-E-A" //}
	client := GET(url1)
	client.AddBasicAuth(username, password)
	ret := client.Do()
	t.Log(ret.BaseResult().RetBody)
	t.Log(ret.BaseResult().ReqBody)
}

func TestUploadFile(t *testing.T) {
	url1 := "http://127.0.0.1/upload"
	client := PostFile(url1, "/tmp/abc.txt", "test1.txt")
	ret := client.Do()
	if ret.Error() != nil {
		t.Log(ret.Error())
	}
	t.Log(ret.BaseResult().RetBody)
}

func TestJsonGet(t *testing.T) {
	type Sample struct {
		Name string `json:"name"`
	}
	//var sample = Sample{}
	body := `{"name":"abc"}`
	asd := jsoniter.Get([]byte(body), "name")
	t.Logf("%v", (asd).ToString())
}
