package httphelper

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ExamplePostJSONStream 展示如何对接一个真实的 SSE 服务（以 OpenAI / 通义 / 混元等
// 大模型流式接口为原型），进行流式逐行解析。
//
// 真实场景要点：
//  1. 通过 SetTimeout 关闭“整体超时”，否则长连接会被 client.Timeout 切断；
//  2. 必须 defer resp.Body.Close()，否则连接泄漏；
//  3. 服务端通常按 SSE 协议推送：每行以 "field: value" 形式，事件之间以空行分隔；
//  4. [DONE] 是主流大模型约定的流结束标记。
func ExamplePostJSONStream() {
	// ---- 1. 启动一个“真实风格”的 SSE 服务端（生产环境替换为真实 URL）----
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 校验请求体确实是 JSON
		body, _ := io.ReadAll(r.Body)
		fmt.Printf("server received: %s\n", string(body))

		// 设置 SSE 响应头（真实服务由对方设置，这里仅演示）
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		flusher, _ := w.(http.Flusher)

		// 按 SSE 协议推送：event / id / data(可多行) / 空行分隔
		fmt.Fprint(w, "event: message\n")
		fmt.Fprint(w, "id: 1\n")
		fmt.Fprint(w, "data: {\"content\":\"你好")
		fmt.Fprint(w, "，我是\"}\n\n") // 多行 data 会被客户端按 \n 拼接
		flusher.Flush()

		fmt.Fprint(w, "event: message\n")
		fmt.Fprint(w, "id: 2\n")
		fmt.Fprint(w, "data: {\"content\":\"流式助手\"}\n\n")
		flusher.Flush()

		// 流结束标记，主流大模型约定为 [DONE]
		fmt.Fprint(w, "event: done\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer srv.Close()

	// ---- 2. 构造流式 POST 请求 ----
	// 注意：使用确定顺序的 JSON 字符串作为 body，避免 map 序列化顺序不确定影响示例输出。
	helper := PostJSONStream(srv.URL, `{"model":"demo-stream","messages":[{"role":"user","content":"你好"}],"stream":true}`)
	// 流式场景：给足 dial 超时，但不对“整体读取”设超时，避免长连接被切断
	helper.SetTimeout(3*time.Second, 0)

	resp, err := helper.DoStream()
	if err != nil {
		fmt.Println("DoStream error:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("unexpected status:", resp.StatusCode)
		return
	}

	// ---- 3. 逐行解析 SSE 事件 ----
	reader := bufio.NewReader(resp.Body)
	var (
		eventName string
		dataLines []string
	)
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimRight(line, "\n")
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("read error:", err)
			break
		}

		switch {
		case line == "":
			// 空行：一个事件结束，处理并重置
			if len(dataLines) > 0 {
				data := strings.Join(dataLines, "\n")
				fmt.Printf("event=%s data=%s\n", eventName, data)
				if data == "[DONE]" {
					fmt.Println("stream finished")
					return
				}
			}
			eventName = ""
			dataLines = nil
		case strings.HasPrefix(line, "event:"):
			eventName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		case strings.HasPrefix(line, ":"):
			// 注释行（以冒号开头），忽略
		}
	}

	// Output:
	// server received: {"model":"demo-stream","messages":[{"role":"user","content":"你好"}],"stream":true}
	// event=message data={"content":"你好，我是"}
	// event=message data={"content":"流式助手"}
	// event=done data=[DONE]
	// stream finished
}

// TestPostJSONStream_RealisticServer 集成测试：验证真实风格 SSE 服务端的逐事件解析。
func TestPostJSONStream_RealisticServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "text/event-stream" {
			t.Errorf("missing Accept: text/event-stream header")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		for i := 1; i <= 3; i++ {
			fmt.Fprintf(w, "id: %d\ndata: chunk-%d\n\n", i, i)
			flusher.Flush()
			time.Sleep(5 * time.Millisecond)
		}
	}))
	defer srv.Close()

	helper := PostJSONStream(srv.URL, map[string]string{"q": "test"})
	helper.SetTimeout(2*time.Second, 0)

	resp, err := helper.DoStream()
	if err != nil {
		t.Fatalf("DoStream error: %v", err)
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	var events []string
	var id, data string
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimRight(line, "\n")
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("read error: %v", err)
		}
		switch {
		case line == "":
			if data != "" {
				events = append(events, fmt.Sprintf("%s:%s", id, data))
			}
			id, data = "", ""
		case strings.HasPrefix(line, "id:"):
			id = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
		case strings.HasPrefix(line, "data:"):
			data = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
	}

	want := []string{"1:chunk-1", "2:chunk-2", "3:chunk-3"}
	if len(events) != len(want) {
		t.Fatalf("got %v, want %v", events, want)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("event[%d] = %q, want %q", i, events[i], want[i])
		}
	}
}

type IMateReply struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		ClientUuid      string `json:"clientUuid"`
		BotName         string `json:"botName"`
		Avatar          string `json:"avatar"`
		ClientType      string `json:"clientType"`
		Username        string `json:"username"`
		IsOwner         bool   `json:"isOwner"`
		Status          string `json:"status"`
		Model           string `json:"model"`
		OpenclawVersion string `json:"openclawVersion"`
		WebchatStatus   string `json:"webchatStatus"`
		ItfsSpaceId     string `json:"itfsSpaceId"`
		AgentType       string `json:"agentType"`
		EnvType         string `json:"envType"`
		ShareType       string `json:"shareType"`
	} `json:"data"`
}

func TestPostJSONStreamIMate(t *testing.T) {
	iMateToken := " tai_pat_ZZNWVhC0RMcG0ypnNeAD6C8_YWI1NfqihtDNup1RzFQ.S3Ih6sZrZBBVhs29tbVpq6jhLB9Dom08OxOYnWlxS6E"
	baseUrl := "http://imate.woa.com/server/web-api/"
	getIMateUrl := baseUrl + "/api/v1/open/imates"
	iMateHeader := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + iMateToken,
		"X-Username":    "chairou",
	}
	type Ret struct {
		Code    int          `json:"code"`
		Message string       `json:"message"`
		Data    []IMateReply `json:"data"`
	}
	ret := Ret{}
	helper := GET(getIMateUrl)
	helper.AddHeaderMap(iMateHeader)
	resp := helper.Do()
	if resp.Error() != nil {
		t.Fatalf(" helper.Do() error: %v", resp.Error())
	}
	err := resp.UnmarshalFromBody(&ret)
	if err != nil {
		t.Errorf("resp.UnmarshalFromBody(&ret) error: %v", err)
		return
	}
	t.Logf("resp: %+v", resp.BaseResult().RetBody)
	t.Logf("ret: %+v", ret)

}
