package test

import (
	"os"
	"testing"
)

func TestFileOperations(t *testing.T) {
	fileInfo, err := os.Stat("test.txt")
	if err != nil {
		t.Error(err)
	}
	t.Log(fileInfo)
	//linuxFileAttr := fileInfo.Sys().(*syscall.Stat_t)
	//fmt.Println(time.Unix(linuxFileAttr.Ctimespec.Unix()))
	//fmt.Println(fileInfo.ModTime())
}
