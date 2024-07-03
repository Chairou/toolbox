package tcp

import (
	"bytes"
	"compress/flate"
	"errors"
	"fmt"
	"io"
	"net"
)

// 压缩插件
func compress(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	if input == nil {
		return nil, errors.New("input is nil")
	}
	var buf bytes.Buffer
	w, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}
	_, err = w.Write(*input)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	result := buf.Bytes()
	fmt.Println("compress result:", string(result))
	return result, err
}

// 解压插件
func unCompress(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	if input == nil {
		return nil, errors.New("input is nil")
	}
	result, err := io.ReadAll(flate.NewReader(bytes.NewReader(*input)))
	fmt.Println("unCom result:", string(result))
	return result, err
}
