package tcp

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

type operation func(*Server, []byte) ([]byte, error)
type flowControl func([]byte) []byte

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

type Server struct {
	Listener    net.Listener
	Connections map[string]*net.Conn
	RwLock      sync.RWMutex
	IpPort      string
	Process     func(svr *Server, input []byte) (output []byte, err error)
	Unpack      func(conn *net.Conn) (output []byte, err error)
	Dispatch    func(input []byte) (output []byte)
	Pack        func(conn *net.Conn, input []byte) (err error)
	UnCompress  func(input []byte) (output []byte)
	Crypt       func(input []byte) (output []byte)
	DeCrypt     func(input []byte) (output []byte)
	Obfuscate   func(input []byte) (output []byte)
	UnObfuscate func(input []byte) (output []byte)
}

func (s *Server) Listen() error {
	var err error
	s.Listener, err = net.Listen("tcp", s.IpPort)
	if err != nil {
		return errors.New("NewTcpConnection Listen err:" + err.Error())
	}
	return nil
}

func (s *Server) Run() error {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			return errors.New("NewTcpConnection Accept err:" + err.Error())
		}
		if s.Process == nil {
			return fmt.Errorf("Server.Procces is nil")
		}
		remoteAddr := conn.RemoteAddr().String()
		s.RwLock.Lock()
		s.Connections[remoteAddr] = &conn
		s.RwLock.Unlock()
		go s.HandleConnection(conn, s.Process)
	}
}

func (s *Server) HandleConnection(conn net.Conn, op operation) {

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Close Connection, err:" + err.Error())
		}
	}(conn)

	for {
		// 首先读取长度前缀
		lengthBuf := make([]byte, 4) // 假设长度字段是4个字节
		_, err := io.ReadFull(conn, lengthBuf)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading length prefix:", err)
			}
			return
		}

		// 解析长度
		length := binary.BigEndian.Uint32(lengthBuf)

		// 根据长度读取数据
		messageBuf := make([]byte, length)
		fmt.Println("buffer length is", length)
		_, err = io.ReadFull(conn, messageBuf)
		if err != nil {
			fmt.Println("Error reading message:", err)
			return
		}

		// 处理消息
		fmt.Printf("Received message: %s\n", string(messageBuf))
		outBuffer, err := op(s, messageBuf)
		if err != nil {
			fmt.Println("Error")
			return
		}

		// 响应客户端（可选）
		// 这里只是简单地将接收到的消息发送回去
		lengthBuf = append(lengthBuf, outBuffer...)
		_, err = conn.Write(lengthBuf)
		if err != nil {
			fmt.Println("Error writing message:", err)
			return
		}
	}
}

func (s *Server) CustomHandleConnection(conn net.Conn, op operation) error {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Close Connection, err:" + err.Error())
		}
	}(conn)
	for {
		unPackBytes, err := s.Unpack(&conn)
		if err != nil {
			return err
		}

		outBytes, err := op(s, unPackBytes)
		if err != nil {
			fmt.Println("op Error, err:", err.Error())
			return err
		}
		err = s.Pack(&conn, outBytes)
		if err != nil {
			fmt.Println("Unpack Error, err:", err.Error())
			return err
		}
	}
}

func (s *Server) DelimiterHandleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Error closing connection:", err)
		}
	}(conn)

	// 使用 bufio.NewReader 来读取数据流
	reader := bufio.NewReader(conn)

	for {
		// 读取直到遇到新的分隔符（在这个例子中是换行符）
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client closed the connection")
			} else {
				fmt.Println("Error reading from connection:", err)
			}
			return
		}

		// 处理消息
		fmt.Printf("Received message: %s", message) // message 包含分隔符

		// 响应客户端（可选）
		// 这里只是简单地将接收到的消息发送回去
		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Error writing to connection:", err)
			return
		}
	}
}

func extractIP(ipPort string) string {
	startIndex := strings.Index(ipPort, "[") + 1
	endIndex := strings.Index(ipPort, "]")
	if startIndex == -1 || endIndex == -1 {
		return ""
	}
	return ipPort[startIndex:endIndex]
}
