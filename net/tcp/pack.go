package tcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/Chairou/toolbox/util/conv"
	"io"
	"net"
	"sync"
)

const MAX_PACKAGE_LENGTH = 65535

var packagePool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 65535)
	},
}

func unPack(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	if conn == nil {
		return []byte{}, errors.New("unPack() conn is nil")
	}

	headerBuf := packagePool.Get().([]byte)
	headerBuf = headerBuf[:svr.HeaderLength]
	defer packagePool.Put(headerBuf)

	_, err = io.ReadFull(conn, headerBuf)
	if err != nil {
		if err == io.EOF {
			fmt.Println("connection closed: ", err)
		} else {
			fmt.Println("Error reading length prefix:", err)
		}
		return []byte{}, err
	}

	lengthBuf := headerBuf[svr.TagSize : svr.TagSize+svr.PacketLengthSize]
	var tLength uint64
	switch svr.PacketLengthSize {
	case 2:
		tLength = uint64(binary.BigEndian.Uint16(lengthBuf))
	case 4:
		tLength = uint64(binary.BigEndian.Uint32(lengthBuf))
	case 8:
		tLength = binary.BigEndian.Uint64(lengthBuf)
	}

	if tLength >= MAX_PACKAGE_LENGTH {
		return []byte{}, errors.New("tLength bigger than MAX_PACKAGE_LENGTH, err = " + conv.String(tLength))
	}

	msgBuf := packagePool.Get().([]byte)
	msgBuf = msgBuf[:tLength-uint64(svr.HeaderLength)]
	defer packagePool.Put(msgBuf)

	_, err = io.ReadFull(conn, msgBuf)
	if err != nil {
		return []byte{}, err
	}
	return msgBuf, nil
}

func Pack(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	if conn == nil {
		return []byte{}, errors.New("Pack() conn is nil")
	}
	if len(*input) <= 0 {
		return []byte{}, errors.New("input zero, quit")
	}

	totalLength := len(*input) + svr.HeaderLength
	tag := "BF"
	if svr.Tag != "" {
		tag = svr.Tag
	}

	packetBuf := packagePool.Get().([]byte)
	packetBuf = packetBuf[:0]
	defer packagePool.Put(packetBuf)

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
	_, err = conn.Write(packetBuf)
	if err != nil {
		return []byte{}, err
	}
	return packetBuf, nil
}

// 读取以结束符结尾的数据
func readUntilEndMarker(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	if conn == nil {
		return []byte{}, errors.New("readUntilEndMarker() conn is nil")
	}

	buffer := packagePool.Get().([]byte)
	buffer = buffer[:0]
	defer packagePool.Put(buffer)

	temp := packagePool.Get().([]byte)
	defer packagePool.Put(temp)

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
	return buffer, nil
}

// 写入结束符
func writeWithEndMark(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	if conn == nil {
		return []byte{}, errors.New("writeWithEndMark() conn is nil")
	}

	outputBuf := packagePool.Get().([]byte)
	outputBuf = outputBuf[:0]
	defer packagePool.Put(outputBuf)

	if input != nil {
		outputBuf = append(*input, svr.EndMarker...)
	} else {
		outputBuf = append(outputBuf, svr.EndMarker...)
	}

	_, err = conn.Write(outputBuf)
	if err != nil {
		return []byte{}, err
	}
	return outputBuf, nil
}
