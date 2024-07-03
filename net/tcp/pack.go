package tcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

func unPack(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	// 首先读取长度前缀
	headerBuf := make([]byte, svr.HeaderLength)
	_, err = io.ReadFull(conn, headerBuf)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error reading length prefix:", err)
		}
		return
	}
	// 通过偏移和长度字段的Size获取长度数据
	lengthBuf := headerBuf[svr.TagSize : svr.TagSize+svr.PacketLengthSize]
	// 解析长度
	var tLength uint64
	switch svr.PacketLengthSize {
	case 2:
		length := binary.BigEndian.Uint16(lengthBuf)
		tLength = uint64(length)
	case 4:
		length := binary.BigEndian.Uint32(lengthBuf)
		tLength = uint64(length)
	case 8:
		length := binary.BigEndian.Uint64(lengthBuf)
		tLength = length
	}
	msgBuf := make([]byte, tLength-uint64(svr.HeaderLength))
	_, err = io.ReadFull(conn, msgBuf)
	if err != nil {
		return nil, err
	}
	fmt.Println("msgBuf: ", string(msgBuf))
	return msgBuf, nil
}

func Pack(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	// 计算长度
	totalLength := len(*input) + svr.HeaderLength
	// 生成tag
	tag := "BF"
	if svr.Tag != "" {
		tag = svr.Tag
	}
	packetBuf := make([]byte, 0, totalLength)
	packetBuf = append(packetBuf, []byte(tag)...)
	switch svr.PacketLengthSize {
	case 2:
		lenBuf := make([]byte, 2)
		binary.BigEndian.PutUint16(lenBuf, uint16(totalLength))
		packetBuf = append(packetBuf, lenBuf...)
	case 4:
		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, uint32(totalLength))
		packetBuf = append(packetBuf, lenBuf...)
	case 8:
		lenBuf := make([]byte, 8)
		binary.BigEndian.PutUint64(lenBuf, uint64(totalLength))
		packetBuf = append(packetBuf, lenBuf...)
	}
	packetBuf = append(packetBuf, *input...)
	_, _ = conn.Write(packetBuf)
	return packetBuf, nil
}

// 读取以结束符结尾的数据
func readUntilEndMarker(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	buffer := make([]byte, 0)
	temp := make([]byte, 65535)
	endMarkerLen := len(svr.EndMarker)

	for {
		n, err := conn.Read(temp)
		if err != nil {
			return []byte{}, err
		}
		buffer = append(buffer, temp[:n]...)

		if len(buffer) >= endMarkerLen && bytes.HasSuffix(buffer, svr.EndMarker) {
			break
		}
	}
	fmt.Println("buffer:", string(buffer))
	return buffer, nil
}

// 写入结束符
func writeWithEndMark(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	output = make([]byte, 0)
	if input != nil {
		output = append(*input, svr.EndMarker...)
	} else {
		output = svr.EndMarker
	}
	_, _ = conn.Write(output)
	return output, err

}
