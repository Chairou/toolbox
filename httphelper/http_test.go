package httphelper

//go test -v http_test.go http.go helper.go

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"net/http"
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
	Age     string            `json:"age"`
	Pairent map[string]string `json:"pairent"`
	Address []string          `json:"address"`
	Name    string            `json:"name"`
}

func TestUrlToMap(t *testing.T) {
	url1 := "l5://2342343:324234/main?s=3&g=ff"
	url2 := "s=3&g=ff"
	url3 := "https://news.qq.com:8080/main?s=3&g=中文去"

	fmt.Printf("%s\n", UrlPathEscape(url1))
	fmt.Printf("%s\n", UrlPathEscape(url2))
	fmt.Printf("%s\n", UrlPathEscape(url3))
}

var time1 time.Time = time.Now()

func TestPostHttp(t *testing.T) {
	tmp := &TmpPost{}
	type MyType func(int) int
	var f MyType = func(x int) int { return x * x }
	//client := PostJSON("http://9.135.96.168:8080/post", "{\"qq\":\"win\"}")
	client := PostJSON("http://9.135.96.168:8080/post", f)
	if client.error() != nil {
		t.Log(client.error())
	}
	val, ok := client.(*errHelper)
	if ok {
		t.Log(*val)
	}
	client = PostJSON("http://9.135.96.168:8080/post", "{\"qq\":\"win\"}")
	if client.error() != nil {
		t.Error(client.error())
	}
	client.SetDebug(DEBUG_DETAIL)
	client.AddHeader("aa", "bb").AddHeader("cc", "dd")
	client.AddSimpleCookies(map[string]string{"ee": "ff"})
	ret := client.Do()
	// 把body返回的json串反序列化到结构中去
	err := ret.Bind(tmp)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("####:", ret.BaseResult().ReqBody)
	fmt.Println("retBody:", ret.BaseResult().RetBody)
	fmt.Printf("struct: %+v", tmp)

}

func TestUrlPathEscape(t *testing.T) {
	url := "http://news.qq.com/?a=1&b=吃饭"
	urlEncode := UrlPathEscape(url)
	fmt.Println(urlEncode)
}

func TestGetHttp(t *testing.T) {
	url := "http://9.135.96.168:8080/get?aa=bb&cc=dd"
	client := GET(url)
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
	url := "http://9.135.96.168:8080/get?aa=bb&cc=dd"
	client := GET(url)
	client.AddPathEscapeQuery("key2", "value2;")
	client.AddPathEscapeQuery("key3", "中文")
	ret := client.Do()
	// 显示body内容
	t.Log(ret.BaseResult().RetBody)
	t.Log(ret.BaseResult().ReqBody)
}

func TestJsoniter(t *testing.T) {
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
