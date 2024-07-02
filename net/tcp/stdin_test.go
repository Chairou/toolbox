package tcp

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func TestStdin(t *testing.T) {
	for {
		text, _ := io.ReadAll(os.Stdin)
		if len(text) > 0 {
			fmt.Println(string(text))
		}
	}
}

//func TestSrvInst(t *testing.T) {
//	srv := Server{}
//	if srv.Dispatch == nil {
//		t.Log("Dispatch is nil")
//	}
//}

func TestTag(t *testing.T) {
	totalLength := 20
	// 生成tag
	tag := strings.Repeat("k", 2)
	packetBuf := make([]byte, 0, totalLength)
	packetBuf = append(packetBuf, []byte(tag)...)
	t.Log(packetBuf)
}
