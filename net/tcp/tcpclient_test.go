package tcp

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/Chairou/toolbox/logger"
	"io"
	"net"
	"testing"
	"time"
)

var gLog *logger.LogPool

func init() {
	gLog, _ = logger.NewLogPool("test1", "test1.log")
	gLog.SetPrintConsole(false)
}

func TestTcpTlvClient(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println("Failed to connect error: ", err)
		return
	}
	//contentBytes := []byte("hello")
	for {
		//time.Sleep(time.Second)
		//outbytes, err := com(&contentBytes)
		outbytes := []byte("hello")
		if err != nil {
			fmt.Println("compress error: ", err)
			return
		}
		bytes := makeTlvBuffer("BF", outbytes)
		conn.Write(bytes)

		tlvRead(conn)
	}
}

func TestTcpEndMarkClient(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println("Failed to connect error: ", err)
		return
	}
	for {
		raw := []byte("ping")
		comBytes, _ := com(&raw)
		comBytes = append(comBytes, []byte("\r\n")...)
		conn.Write(comBytes)
		buf, _ := readEndMarker(conn.(*net.TCPConn))
		unCom(&buf)
		time.Sleep(time.Second)

	}
}

func makeTlvBuffer(tag string, content []byte) []byte {
	buf := bufferPool.Get().([]byte)
	buf = buf[:0]
	defer bufferPool.Put(buf)
	var length uint32
	// 再打包
	length = uint32(len(content) + 4 + 2)
	//buffer := make([]byte, 0, 4+2)
	buf = append(buf, []byte(tag)...)
	//fmt.Println("package length:", length)
	//lenBuff := make([]byte, 4)
	//fmt.Println("length : ", length)
	//tmp := bufferPool1.Get().([]byte)
	//tmp := []byte{0x00, 0x00, 0x00, 0x00}
	//defer bufferPool1.Put(tmp)
	//tmp = tmp[:4]
	buf = buf[:4+2]
	binary.BigEndian.PutUint32(buf[2:], length)
	//buf = append(buf, tmp...)
	//buffer = append(buffer, []byte(content)...)
	buf = append(buf, []byte(content)...)
	//fmt.Println("package buff:", string(buf))
	return buf
}

func tlvRead(conn net.Conn) {
	// 首先读取长度前缀
	lengthBuf := bufferPool.Get().([]byte)
	lengthBuf = lengthBuf[:4+2]
	defer bufferPool.Put(lengthBuf)
	//lengthBuf := make([]byte, 4+2) // 假设tag字段长度2字节，长度字段是4个字节
	_, err := io.ReadFull(conn, lengthBuf)
	if err != nil {
		//fmt.Println("ReadFull err: ", err.Error())
		if err != io.EOF {
			fmt.Println("Error reading length prefix:", err.Error())
			return
		}
	}
	//fmt.Println("Tlv buffer header: ", lengthBuf)
	// 解析长度
	length := binary.BigEndian.Uint32(lengthBuf[2 : 4+2])

	// 根据长度读取数据
	//fmt.Println("messageBuf length:", length)
	messageBuf := bufferPool.Get().([]byte)
	messageBuf = messageBuf[:length-6]
	//messageBuf := make([]byte, length-6)
	_, err = io.ReadFull(conn, messageBuf)
	if err != nil {
		fmt.Println("Error reading message:", err)
		return
	}
	// 解压缩
	//content, err := unCom(&messageBuf)
	//if err != nil {
	//	fmt.Println("unCom err: ", err.Error())
	//	return
	//}

	//fmt.Printf("Received message: %s\n", string(content))
	// 在这里写日志
	gLog.Infof("Received message: %s\n", string(messageBuf))

}

func readEndMarker(conn *net.TCPConn) (output []byte, err error) {
	buffer := make([]byte, 0)
	temp := make([]byte, 65535)
	endMarkerLen := len([]byte("\r\n"))

	for {
		n, err := conn.Read(temp)
		if err != nil {
			return []byte{}, err
		}
		buffer = append(buffer, temp[:n]...)

		if len(buffer) >= endMarkerLen && bytes.HasSuffix(buffer, []byte("\r\n")) {
			break
		}
	}
	fmt.Println("buffer:", string(buffer))
	return buffer, nil
}

func unCom(input *[]byte) (output []byte, err error) {
	if input == nil {
		return nil, errors.New("input is nil")
	}
	result, err := io.ReadAll(flate.NewReader(bytes.NewReader(*input)))
	//fmt.Println("unCom result:", string(result))
	return result, err
}

func com(input *[]byte) (output []byte, err error) {
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
	output = buf.Bytes()
	return output, err
}
