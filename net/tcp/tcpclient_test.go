package tcp

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func TestTcpClient(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println("Failed to connect error: ", err)
		return
	}
	for {
		time.Sleep(time.Second)
		conn.Write(makeTlvBuffer("BF", "hello"))
		read(conn)
	}
}

func makeTlvBuffer(tag string, content string) []byte {
	var length uint32
	length = uint32(len(content) + 4 + 2)
	buffer := make([]byte, 0, 4+2)
	buffer = append(buffer, []byte(tag)...)
	fmt.Println("package length:", length)
	lenBuff := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuff, length)
	buffer = append(buffer, lenBuff...)
	buffer = append(buffer, []byte(content)...)
	fmt.Println("package buff:", buffer)
	return buffer
}

func read(conn net.Conn) {
	// 首先读取长度前缀
	lengthBuf := make([]byte, 4+2) // 假设tag字段长度2字节，长度字段是4个字节
	_, err := io.ReadFull(conn, lengthBuf)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error reading length prefix:", err)
		}
		return
	}
	fmt.Println("Tlv buffer header: ", lengthBuf)
	// 解析长度
	length := binary.BigEndian.Uint32(lengthBuf[2 : 4+2])

	// 根据长度读取数据
	fmt.Println("messageBuf length:", length)
	messageBuf := make([]byte, length-6)
	_, err = io.ReadFull(conn, messageBuf)
	if err != nil {
		fmt.Println("Error reading message:", err)
		return
	}

	// 处理消息
	fmt.Printf("Received message: %s\n", string(messageBuf))
}
