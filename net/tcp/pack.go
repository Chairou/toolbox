package tcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/Chairou/toolbox/util/conv"
)

const MAX_PACKAGE_LENGTH = 65535

var packagePool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 65535)
	},
}

// headerPool 专用于小块 header 读取，避免从 65535 的大 pool 中取
var headerPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 16) // 足够容纳最大 header (tag + 8字节长度)
	},
}

func unPack(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	if conn == nil {
		return []byte{}, errors.New("unPack() conn is nil")
	}

	// 优先使用 bufio.Reader，减少系统调用次数
	var reader io.Reader
	if ctx := getConnCtx(conn); ctx != nil {
		reader = ctx.reader
	} else {
		reader = conn
	}

	// 使用专用的小 headerPool
	headerBuf := headerPool.Get().([]byte)
	headerBuf = headerBuf[:svr.HeaderLength]
	defer headerPool.Put(headerBuf)

	_, err = io.ReadFull(reader, headerBuf)
	if err != nil {
		if err == io.EOF {
			fmt.Println("connection closed: ", err)
		} else {
			fmt.Println("Error reading length prefix:", err)
		}
		return []byte{}, err
	}

	// 校验 Tag 字段：使用 bytes.Equal 避免 string 分配
	tagBytes := svr.TagBytes
	if len(tagBytes) == 0 {
		tagBytes = []byte(svr.Tag)
	}
	if !bytes.Equal(headerBuf[:svr.TagSize], tagBytes) {
		return []byte{}, fmt.Errorf("invalid tag: expected %q, got %q", svr.Tag, string(headerBuf[:svr.TagSize]))
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

	// 防止 tLength < HeaderLength 导致 uint64 下溢
	if tLength < uint64(svr.HeaderLength) {
		return []byte{}, errors.New("invalid packet: total length less than header length")
	}

	msgLen := int(tLength) - svr.HeaderLength
	// 直接分配精确大小的 buffer
	result := make([]byte, msgLen)
	_, err = io.ReadFull(reader, result)
	if err != nil {
		return []byte{}, err
	}

	return result, nil
}

func Pack(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	if conn == nil {
		return []byte{}, errors.New("Pack() conn is nil")
	}
	inputLen := len(*input)
	if inputLen <= 0 {
		return []byte{}, errors.New("input zero, quit")
	}

	totalLength := inputLen + svr.HeaderLength

	// 使用预计算的 TagBytes，避免每次 string→[]byte 转换
	tag := svr.TagBytes
	if len(tag) == 0 {
		tag = []byte("BF")
	}

	// 直接分配精确大小的 buffer，一次性组装完整数据包
	packetBuf := make([]byte, totalLength)
	copy(packetBuf, tag)

	// 直接写入长度字段
	switch svr.PacketLengthSize {
	case 2:
		binary.BigEndian.PutUint16(packetBuf[svr.TagSize:], uint16(totalLength))
	case 4:
		binary.BigEndian.PutUint32(packetBuf[svr.TagSize:], uint32(totalLength))
	case 8:
		binary.BigEndian.PutUint64(packetBuf[svr.TagSize:], uint64(totalLength))
	}

	copy(packetBuf[svr.HeaderLength:], *input)

	// 优先使用 bufio.Writer，减少系统调用次数
	if ctx := getConnCtx(conn); ctx != nil {
		_, err = ctx.writer.Write(packetBuf)
		if err != nil {
			return []byte{}, err
		}
		// 立即 Flush，确保数据发送给客户端
		err = ctx.writer.Flush()
		if err != nil {
			return []byte{}, err
		}
	} else {
		_, err = conn.Write(packetBuf)
		if err != nil {
			return []byte{}, err
		}
	}

	return packetBuf, nil
}

// 读取以结束符结尾的数据
func readUntilEndMarker(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	if conn == nil {
		return []byte{}, errors.New("readUntilEndMarker() conn is nil")
	}

	// 优先使用 bufio.Reader
	var rawReader io.Reader
	if ctx := getConnCtx(conn); ctx != nil {
		rawReader = ctx.reader
	} else {
		rawReader = conn
	}

	buffer := packagePool.Get().([]byte)
	buffer = buffer[:0]
	defer packagePool.Put(buffer)

	// 使用栈上固定大小的临时缓冲区
	var temp [4096]byte

	endMarkerLen := len(svr.EndMarker)

	for {
		n, err := rawReader.Read(temp[:])
		if err != nil {
			return []byte{}, err
		}
		buffer = append(buffer, temp[:n]...)

		// 防止无结束符时内存无限增长
		if len(buffer) > MAX_PACKAGE_LENGTH {
			return []byte{}, errors.New("data exceeds MAX_PACKAGE_LENGTH before end marker")
		}

		if len(buffer) >= endMarkerLen && bytes.HasSuffix(buffer, svr.EndMarker) {
			break
		}
	}

	// 拷贝一份独立数据返回
	result := make([]byte, len(buffer))
	copy(result, buffer)
	return result, nil
}

// 写入结束符
func writeWithEndMark(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	if conn == nil {
		return []byte{}, errors.New("writeWithEndMark() conn is nil")
	}

	// 直接分配精确大小的 buffer，一次性组装完整数据
	var inputLen int
	if input != nil {
		inputLen = len(*input)
	}
	result := make([]byte, inputLen+len(svr.EndMarker))
	if input != nil {
		copy(result, *input)
	}
	copy(result[inputLen:], svr.EndMarker)

	// 优先使用 bufio.Writer
	if ctx := getConnCtx(conn); ctx != nil {
		_, err = ctx.writer.Write(result)
		if err != nil {
			return []byte{}, err
		}
		err = ctx.writer.Flush()
		if err != nil {
			return []byte{}, err
		}
	} else {
		_, err = conn.Write(result)
		if err != nil {
			return []byte{}, err
		}
	}

	return result, nil
}
