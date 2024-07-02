package tcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
)

type operation func(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error)

type Server struct {
	Listener      net.Listener
	Connections   map[string]*net.TCPConn
	RwLock        sync.RWMutex
	IpPort        string
	OperationList []operation
	Tag           string
	HeaderLength  int
	TagSize       int
	LengthSize    int
	EndMarker     []byte
}

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
	lengthBuf := headerBuf[svr.TagSize : svr.TagSize+svr.LengthSize]
	// 解析长度
	var tLength uint64
	switch svr.LengthSize {
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

func Content(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	return append(*input, []byte("Hello, world!")...), nil
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
	switch svr.LengthSize {
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

func NewTcpServer(ipPort string) (*Server, error) {
	var err error
	srv := &Server{}
	srv.IpPort = ipPort
	srv.Connections = make(map[string]*net.TCPConn)
	srv.Listener, err = srv.Listen()
	if err != nil {
		fmt.Println("NewTcpServer err, ipv4 like 192.168.0.250:8080 ")
		fmt.Println("NewTcpServer err, ipv6 like [2001:0db8:86a3:08d3:1319:8a2e:0370:7344]:8080")
		fmt.Println("NewTcpServer err, both ipv4 and ipv6 like 0:8080")
		return nil, err
	}
	return srv, nil
}

// 定义处理过程函数，前一个处理函数的输出是后一个处理函数的输入
func (s *Server) process(functions []operation, t *Server, conn *net.TCPConn, input *[]byte) error {
	var err error
	var result []byte
	if input == nil {
		result = nil
	} else {
		result = *input
	}
	for _, f := range functions {
		result, err = f(t, conn, &result)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Listen() (net.Listener, error) {
	var err error
	listener, err := net.Listen("tcp", s.IpPort)
	if err != nil {
		return listener, errors.New("NewTcpConnection Listen err:" + err.Error())
	}
	return listener, nil
}

func (s *Server) Run() error {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			return errors.New("NewTcpConnection Accept err:" + err.Error())
		}
		remoteAddr := conn.RemoteAddr().String()
		s.RwLock.Lock()
		s.Connections[remoteAddr] = conn.(*net.TCPConn)
		s.RwLock.Unlock()
		go s.HandleTlvConnection(conn.(*net.TCPConn))
	}
}

func (s *Server) HandleTlvConnection(conn *net.TCPConn) {
	defer func(conn *net.TCPConn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Close Connection, err:" + err.Error())
		}
	}(conn)
	for {
		err := s.process(s.OperationList, s, conn, nil)
		if err != nil {
			return
		}
	}
}

func (s *Server) SetTag(tag string) {
	s.Tag = tag
	s.TagSize = len(s.Tag)
}

// 读取以结束符结尾的数据
func readUntilEndMarker(svr *Server, conn *net.TCPConn, input []byte) (output []byte, err error) {
	buffer := make([]byte, 0)
	temp := make([]byte, 1024)
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
func makeWithEndMark(svr *Server, conn *net.TCPConn, input []byte) (output []byte, err error) {
	input = append(input, svr.EndMarker...)
	return input, nil
}
