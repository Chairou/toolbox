package tcp

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"testing"
)

func TestTcpServer(t *testing.T) {
	svr, err := NewTcpServer("127.0.0.1:8080")
	if err != nil {
		fmt.Println("Error creating tcp server: ", err)
	}
	svr.Process = process
	svr.Unpack = unPack
	err = svr.Run()
	if err != nil {
		fmt.Println("Error running tcp server: ", err)
		return
	}

}

func process(srv *Server, input []byte) (output []byte, err error) {
	if srv.UnCompress != nil {
		// 处理解压工作
	}
	if
	
}

// unPack函数是关键，因为header的混淆是关键，这里就不展示了，免我党学到了：）
func unPack(conn *net.Conn) ([]byte, error) {
	// 首先读取长度前缀
	lengthBuf := make([]byte, 4) // 假设长度字段是4个字节
	_, err := io.ReadFull(*conn, lengthBuf)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error reading length prefix:", err)
		}
		return []byte{}, err
	}

	// 解析长度
	length := binary.BigEndian.Uint32(lengthBuf)

	// 根据长度读取数据
	messageBuf := make([]byte, length)
	fmt.Println("buffer length is", length)
	_, err = io.ReadFull(*conn, messageBuf)
	if err != nil {
		fmt.Println("Error reading message:", err)
		return []byte{}, err
	}
	return messageBuf, nil
}
