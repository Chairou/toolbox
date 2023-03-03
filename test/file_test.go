package test

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestFileOperations(t *testing.T) {
	fileInfo, err := os.Stat("test.txt")
	if err != nil {
		t.Error(err)
	}

	linuxFileAttr := fileInfo.Sys().(*syscall.Stat_t)
	fmt.Println(time.Unix(linuxFileAttr.Ctimespec.Unix()))
	fmt.Println(fileInfo.ModTime())
}
