package tcp

import (
	"fmt"
	"io"
	"os"
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

func TestSrvInst(t *testing.T) {
	srv := Server{}
	if srv.Dispatch == nil {
		t.Log("Dispatch is nil")
	}
}
