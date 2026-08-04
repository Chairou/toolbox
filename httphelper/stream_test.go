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

// startSSEServer 启动一个 SSE mock 服务端，按顺序推送若干 data 事件。
func startSSEServer(t *testing.T, events []string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		// 校验请求方法、Content-Type 与 Accept
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			http.Error(w, "expect application/json", http.StatusBadRequest)
			return
		}
		if accept := r.Header.Get("Accept"); accept != "text/event-stream" {
			http.Error(w, "expect text/event-stream", http.StatusBadRequest)
			return
		}
		// 读取并打印请求体，确认 JSON 已正确发送
		body, _ := io.ReadAll(r.Body)
		t.Logf("server received body: %s", string(body))

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		for _, e := range events {
			fmt.Fprintf(w, "data: %s\n\n", e)
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
	})
	return httptest.NewServer(mux)
}

// readSSEEvents 从 *http.Response 中逐条解析 SSE 的 data 字段。
func readSSEEvents(t *testing.T, resp *http.Response) ([]string, error) {
	t.Helper()
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		return nil, fmt.Errorf("unexpected content-type: %s", ct)
	}

	var events []string
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return events, err
		}
		line = strings.TrimRight(line, "\n")
		if strings.HasPrefix(line, "data: ") {
			events = append(events, strings.TrimPrefix(line, "data: "))
		}
		// 空行表示一条事件结束，这里事件都是单行的，简单忽略即可
	}
	return events, nil
}

func TestPostJSONStream_Basic(t *testing.T) {
	want := []string{"hello", "world", "go"}
	srv := startSSEServer(t, want)
	defer srv.Close()

	helper := PostJSONStream(srv.URL+"/sse", map[string]string{"prompt": "hi"})
	helper.SetTimeout(2*time.Second, 5*time.Second)

	resp, err := helper.DoStream()
	if err != nil {
		t.Fatalf("DoStream error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	got, err := readSSEEvents(t, resp)
	if err != nil {
		t.Fatalf("readSSEEvents error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("event count mismatch: got %d (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("event[%d] mismatch: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestPostJSONStream_StringBody(t *testing.T) {
	want := []string{"raw-json"}
	srv := startSSEServer(t, want)
	defer srv.Close()

	// 直接传 string 类型 body
	helper := PostJSONStream(srv.URL+"/sse", `{"k":"v"}`)
	helper.SetTimeout(2*time.Second, 5*time.Second)

	resp, err := helper.DoStream()
	if err != nil {
		t.Fatalf("DoStream error: %v", err)
	}
	got, err := readSSEEvents(t, resp)
	if err != nil {
		t.Fatalf("readSSEEvents error: %v", err)
	}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("unexpected events: %v", got)
	}
}

func TestPostJSONStream_ErrHelper(t *testing.T) {
	// 构造一个 errorHelper，验证 DoStream 的错误路径。
	h := errorHelper(fmt.Errorf("boom"))
	resp, err := h.DoStream()
	if err == nil {
		t.Fatalf("expected error from errHelper.DoStream, got nil")
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %v", resp)
	}
	if err.Error() != "boom" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestPostJSONStream_BadURL(t *testing.T) {
	// 使用一个无法连接的地址，验证网络错误能被返回
	helper := PostJSONStream("http://127.0.0.1:1/sse", map[string]string{"a": "b"})
	helper.SetTimeout(500*time.Millisecond, 1*time.Second)
	resp, err := helper.DoStream()
	if err == nil {
		t.Fatalf("expected connection error, got nil")
	}
	if resp != nil {
		t.Fatalf("expected nil response on error, got %v", resp)
	}
}
