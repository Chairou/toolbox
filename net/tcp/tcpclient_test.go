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
		conn.Write(makeTlvBuffer("hello world"))
		read(conn)
	}
}

func makeTlvBuffer(content string) []byte {
	var length uint32
	length = uint32(len(content))
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, length)
	buffer = append(buffer, []byte(content)...)
	return buffer
	//var buffer bytes.Buffer
	//binary.Write(&buffer, binary.BigEndian, length)
	//buffer.Write([]byte(content))
	//fmt.Println("Tlv buffer : ", buffer.Bytes())
	//return buffer.Bytes()
}

func read(conn net.Conn) {
	// 首先读取长度前缀
	lengthBuf := make([]byte, 4) // 假设长度字段是4个字节
	_, err := io.ReadFull(conn, lengthBuf)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error reading length prefix:", err)
		}
		return
	}
	fmt.Println("Tlv buffer header: ", lengthBuf)
	// 解析长度
	length := binary.BigEndian.Uint32(lengthBuf)

	// 根据长度读取数据
	messageBuf := make([]byte, length)
	fmt.Println("messageBuf length:", length)
	_, err = io.ReadFull(conn, messageBuf)
	if err != nil {
		fmt.Println("Error reading message:", err)
		return
	}

	// 处理消息
	fmt.Printf("Received message: %s\n", string(messageBuf))
}
