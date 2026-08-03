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

	jsoniter "github.com/json-iterator/go"
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
}

type IMateChatListRet struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		List []struct {
			ChatKey       string `json:"chatKey"`
			ClientUuid    string `json:"clientUuid"`
			Creator       string `json:"creator"`
			ChatName      string `json:"chatName"`
			LastMessageId string `json:"lastMessageId"`
			Status        int    `json:"status"`
			ChatType      string `json:"chatType"`
			LastSessionId string `json:"lastSessionId"`
			Model         string `json:"model"`
			CreatedAt     string `json:"createdAt"`
			UpdatedAt     string `json:"updatedAt"`
		} `json:"list"`
		Total string `json:"total"`
	} `json:"data"`
}

type TextMessage struct {
	Text                string `json:"text"`
	UploadMediaRevision int    `json:"upload_media_revision"`
	Files               []struct {
		Url  string `json:"url"`
		Id   string `json:"id"`
		Mime string `json:"mime"`
		Name string `json:"name"`
		Size int    `json:"size"`
	} `json:"files"`
}
type TextMessageReply struct {
	Seq     int `json:"seq"`
	Content struct {
		Type    string `json:"type"`
		Payload struct {
			Delta string `json:"delta"`
		} `json:"payload"`
		Timestamp int64 `json:"timestamp"`
	} `json:"content"`
}

func TestPostJSONStreamIMate(t *testing.T) {
	iMateToken := "ut_prod_chairou_thjakloo83avxqh8wds0vzm3e"
	baseUrl := "http://imate.woa.com/server/web-api/"
	getIMateUrl := baseUrl + "/api/v1/open/imates"
	iMateHeader := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + iMateToken,
		"X-Username":    "chairou",
	}
	type IMateRet struct {
		Code    int          `json:"code"`
		Message string       `json:"message"`
		Data    []IMateReply `json:"data"`
	}
	iMateRet := IMateRet{}
	helper := GET(getIMateUrl)
	helper.AddHeaderMap(iMateHeader)
	resp := helper.Do()
	if resp.Error() != nil {
		t.Fatalf(" helper.Do() error: %v", resp.Error())
	}
	err := resp.UnmarshalFromBody(&iMateRet)
	if err != nil {
		t.Errorf("resp.UnmarshalFromBody(&iMateRet) error: %v", err)
		return
	}
	t.Logf("iMateRet: %+v", resp.BaseResult().RetBody)
	t.Logf("iMateRet: %+v", iMateRet)

	// 获取会话列表
	//GET /api/v1/open/imates/{clientUuid}/chats
	chatListUrl := baseUrl + "/api/v1/open/imates/" + iMateRet.Data[0].ClientUuid + "/chats"
	helper = GET(chatListUrl)
	helper.AddHeaderMap(iMateHeader)
	resp = helper.Do()
	chatListRet := IMateChatListRet{}
	if resp.Error() != nil {
		t.Fatalf(" helper.Do() error: %v", resp.Error())
	}
	err = resp.UnmarshalFromBody(&chatListRet)
	if err != nil {
		t.Errorf("resp.UnmarshalFromBody(&chatListRet) error: %v", err)
		return
	}
	t.Logf("chatListRet: %+v", chatListRet)
	// 进行对话
	chatUrl := baseUrl + "/api/v1/open/imates/" + iMateRet.Data[0].ClientUuid + "/chats/" + chatListRet.Data.List[0].ChatKey + "/messages/stream"
	messageReq := TextMessage{
		Text: "你好，请输出全国各大城市今天的天气情况",
	}
	helper = PostJSONStream(chatUrl, messageReq)
	helper.AddHeaderMap(iMateHeader)
	// 流式场景：dial 超时 5s 足够，整体超时必须设为 0（不限制），
	// 否则 client.Timeout 会把长连接一并计入超时，导致中途被切断。
	helper.SetTimeout(5*time.Second, 0)

	streamResp, err := helper.DoStream()
	if err != nil {
		t.Fatalf(" helper.DoStream() error: %v", err)
	}
	defer streamResp.Body.Close()

	if streamResp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d, %s", streamResp.StatusCode, streamResp.Status)
	}
	reader := bufio.NewReader(streamResp.Body)
	var events []string
	var id, data, finalData string
	// flush 将当前累积的 (id, data) 作为一条事件保存并重置
	flush := func() {
		if data != "" {
			events = append(events, fmt.Sprintf("%s:%s", id, data))
			textMessageReply := &TextMessageReply{}
			err = jsoniter.UnmarshalFromString(data, &textMessageReply)
			if err != nil {
				t.Errorf("jsoniter.UnmarshalFromString(data, &textMessageReply) error: %v", err)
				return
			}
			finalData += textMessageReply.Content.Payload.Delta
			t.Logf("SSE event => id=%s data=%s", id, data)
		}
		id, data = "", ""
	}
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimRight(line, "\n")
		if err != nil {
			if err == io.EOF {
				// EOF 前把最后一条未落盘的事件冲出去
				flush()
				break
			}
			t.Fatalf("read error: %v", err)
		}
		switch {
		case line == "":
			flush()
		case strings.HasPrefix(line, "id:"):
			id = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
		case strings.HasPrefix(line, "data:"):
			data = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
	}
	t.Logf("total SSE events: %d", len(events))
	t.Log("finalData: ", finalData)
}
