package tcp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

type operation func(svr *Server, conn *net.Conn, input []byte) (output []byte, err error)
type flowControl func([]byte) []byte

func unPack(svr *Server, conn *net.Conn, input []byte) (output []byte, err error) {
	// 首先读取长度前缀
	headerBuf := make([]byte, svr.HeaderLength) // 假设长度字段是4个字节
	_, err = io.ReadFull(*conn, headerBuf)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error reading length prefix:", err)
		}
		return
	}
	// 通过偏移和长度字段的Size获取长度数据
	lengthBuf := headerBuf[svr.HeaderOffSet-1 : svr.HeaderOffSet+svr.LengthSize-1]
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
	msgBuf := make([]byte, tLength)
	_, err = io.ReadFull(*conn, msgBuf)
	if err != nil {
		return nil, err
	}
	return msgBuf, nil
}

func Pack(svr *Server, conn *net.Conn, input []byte) (output []byte, err error) {
	totalLength := len(input) + svr.HeaderLength

	// 生成tag
	tag := strings.Repeat("k", svr.HeaderOffSet)
	packetBuf := make([]byte, 0, totalLength)
	packetBuf = append(packetBuf, []byte(tag)...)
	switch svr.LengthSize {
	case 2:
		binary.BigEndian.PutUint16(packetBuf[svr.HeaderOffSet:svr.HeaderOffSet+svr.LengthSize], uint16(totalLength))
	case 4:
		binary.BigEndian.PutUint32(packetBuf[svr.HeaderOffSet:svr.HeaderOffSet+svr.LengthSize], uint32(totalLength))
	case 8:
		binary.BigEndian.PutUint64(packetBuf[svr.HeaderOffSet:svr.HeaderOffSet+svr.LengthSize], uint64(totalLength))
	}
	return packetBuf, nil
}

func NewTcpServer(ipPort string) (*Server, error) {
	srv := &Server{}
	srv.IpPort = ipPort
	err := srv.Listen()
	if err != nil {
		fmt.Println("NewTcpServer err, ipv4 like 192.168.0.250:8080 ")
		fmt.Println("NewTcpServer err, ipv6 like [2001:0db8:86a3:08d3:1319:8a2e:0370:7344]:8080")
		fmt.Println("NewTcpServer err, both ipv4 and ipv6 like 0:8080")
		return nil, err
	}
	return srv, nil
}

// 定义处理过程函数，前一个处理函数的输出是后一个处理函数的输入
func (s *Server) process(funcs []operation, t *Server, conn *net.Conn, input []byte) error {
	var err error
	result := input
	for _, f := range funcs {
		result, err = f(t, conn, result)
		if err != nil {
			return err
		}
	}
	return nil
}

type Server struct {
	Listener      net.Listener
	Connections   map[string]*net.Conn
	RwLock        sync.RWMutex
	IpPort        string
	OperationList []operation
	HeaderLength  int
	HeaderOffSet  int
	LengthSize    int
	Delimiter     string
}

func (s *Server) Listen() error {
	var err error
	s.Listener, err = net.Listen("tcp", s.IpPort)
	if err != nil {
		return errors.New("NewTcpConnection Listen err:" + err.Error())
	}
	return nil
}

//func (s *Server) Run() error {
//	for {
//		conn, err := s.Listener.Accept()
//		if err != nil {
//			return errors.New("NewTcpConnection Accept err:" + err.Error())
//		}
//		if s.Process == nil {
//			return fmt.Errorf("Server.Procces is nil")
//		}
//		remoteAddr := conn.RemoteAddr().String()
//		s.RwLock.Lock()
//		s.Connections[remoteAddr] = &conn
//		s.RwLock.Unlock()
//		go s.HandleConnection(conn)
//	}
//}

//func (s *Server) HandleConnection(conn net.Conn) {
//
//
//	defer func(conn net.Conn) {
//		err := conn.Close()
//		if err != nil {
//			fmt.Println("Close Connection, err:" + err.Error())
//		}
//	}(conn)
//
//	for {
//		// 首先读取长度前缀
//		lengthBuf := make([]byte, 4) // 假设长度字段是4个字节
//		_, err := io.ReadFull(conn, lengthBuf)
//		if err != nil {
//			if err != io.EOF {
//				fmt.Println("Error reading length prefix:", err)
//			}
//			return
//		}
//
//		// 解析长度
//		length := binary.BigEndian.Uint32(lengthBuf)
//
//		// 根据长度读取数据
//		messageBuf := make([]byte, length)
//		fmt.Println("buffer length is", length)
//		_, err = io.ReadFull(conn, messageBuf)
//		if err != nil {
//			fmt.Println("Error reading message:", err)
//			return
//		}
//
//		// 处理消息
//		fmt.Printf("Received message: %s\n", string(messageBuf))
//		//outBuffer, err := op(s, messageBuf)
//		if err != nil {
//			fmt.Println("Error")
//			return
//		}
//
//		// 响应客户端（可选）
//		// 这里只是简单地将接收到的消息发送回去
//		lengthBuf = append(lengthBuf, outBuffer...)
//		_, err = conn.Write(lengthBuf)
//		if err != nil {
//			fmt.Println("Error writing message:", err)
//			return
//		}
//	}
//}

//func (s *Server) CustomHandleConnection(conn net.Conn, op operation) error {
//	defer func(conn net.Conn) {
//		err := conn.Close()
//		if err != nil {
//			fmt.Println("Close Connection, err:" + err.Error())
//		}
//	}(conn)
//	for {
//		unPackBytes, err := s.Unpack(&conn)
//		if err != nil {
//			return err
//		}
//
//		outBytes, err := op(s, unPackBytes)
//		if err != nil {
//			fmt.Println("op Error, err:", err.Error())
//			return err
//		}
//		err = s.Pack(&conn, outBytes)
//		if err != nil {
//			fmt.Println("Unpack Error, err:", err.Error())
//			return err
//		}
//	}
//}

//func (s *Server) DelimiterHandleConnection(conn net.Conn) {
//	defer func(conn net.Conn) {
//		err := conn.Close()
//		if err != nil {
//			fmt.Println("Error closing connection:", err)
//		}
//	}(conn)
//
//	// 使用 bufio.NewReader 来读取数据流
//	reader := bufio.NewReader(conn)
//
//	for {
//		// 读取直到遇到新的分隔符（在这个例子中是换行符）
//		message, err := reader.ReadString('\n')
//		if err != nil {
//			if err == io.EOF {
//				fmt.Println("Client closed the connection")
//			} else {
//				fmt.Println("Error reading from connection:", err)
//			}
//			return
//		}
//
//		// 处理消息
//		fmt.Printf("Received message: %s", message) // message 包含分隔符
//
//		// 响应客户端（可选）
//		// 这里只是简单地将接收到的消息发送回去
//		_, err = conn.Write([]byte(message))
//		if err != nil {
//			fmt.Println("Error writing to connection:", err)
//			return
//		}
//	}
//}

func extractIP(ipPort string) string {
	startIndex := strings.Index(ipPort, "[") + 1
	endIndex := strings.Index(ipPort, "]")
	if startIndex == -1 || endIndex == -1 {
		return ""
	}
	return ipPort[startIndex:endIndex]
}
