package tcp

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

type operation func(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error)

const TYPE_TLV = 1
const TYPE_ENDMARK = 2
const PACKAGE_LENGTH_TWO_BYTE = 2
const PACKAGE_LENGTH_FOUR_BYTE = 4
const PACKAGE_LENGTH_EIGHT_BYTE = 8

type ServerOption struct {
	Type             int
	Tag              string
	PacketLengthSize int
	EndMarker        []byte
}

type Server struct {
	Type              int
	Listener          net.Listener
	UserConnectionMap map[string]*net.TCPConn
	ConnectionUserMap map[*net.TCPConn]string
	RwLock            sync.RWMutex
	IpPort            string
	OperationList     []operation
	Tag               string
	HeaderLength      int
	TagSize           int
	PacketLengthSize  int
	EndMarker         []byte
}

func Content(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	return append(*input, []byte(", Hello, world!")...), nil
}

func NewTcpServerOption(ipPort string, option ServerOption) (*Server, error) {
	var err error
	srv := &Server{}
	srv.IpPort = ipPort
	srv.Type = option.Type

	if option.Type == TYPE_TLV {
		if option.PacketLengthSize != 2 && option.PacketLengthSize != 4 && option.PacketLengthSize != 8 {
			return nil, errors.New("PacketLengthSize must be 2, 4 or 8")
		}
		if option.Tag != "" {
			srv.SetTag(option.Tag)
		} else {
			srv.SetTag("BF")
		}
		srv.PacketLengthSize = option.PacketLengthSize
		srv.HeaderLength = srv.TagSize + option.PacketLengthSize

	} else if option.Type == TYPE_ENDMARK {
		if len(option.EndMarker) == 0 {
			return nil, errors.New("tag is empty")
		}
		srv.EndMarker = option.EndMarker
	} else {
		return nil, errors.New("type must be TYPE_TLV or TYPE_ENDMARK")
	}

	srv.UserConnectionMap = make(map[string]*net.TCPConn, 1024)
	srv.ConnectionUserMap = make(map[*net.TCPConn]string, 1024)
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
			fmt.Println("主循环错误，退出：", err)
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
		go s.HandleConnection(conn.(*net.TCPConn))
	}
}

func (s *Server) HandleConnection(conn *net.TCPConn) {
	defer func(conn *net.TCPConn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Close Connection, err:" + err.Error())
		}
	}(conn)
	for {
		err := s.process(s.OperationList, s, conn, nil)
		// 有严重错误，直接退出，日志记录放模块内处理
		if err != nil {
			return
		}
	}
}

func (s *Server) SetTag(tag string) {
	s.Tag = tag
	s.TagSize = len(tag)
}

func (s *Server) SetConn(key string, conn *net.TCPConn) {
	s.RwLock.Lock()
	if conn != nil {
		s.UserConnectionMap[key] = conn
		s.ConnectionUserMap[conn] = key
	}
	s.RwLock.Unlock()
}

func (s *Server) GetConn(key string) (conn *net.TCPConn, err error) {
	s.RwLock.Lock()
	defer s.RwLock.Unlock()
	conn, ok := s.UserConnectionMap[key]
	if ok {
		return conn, nil
	} else {
		return nil, errors.New("Connection not found")
	}
}

func (s *Server) DelUserConn(key string) (err error) {
	s.RwLock.Lock()
	defer s.RwLock.Unlock()
	conn, exists := s.UserConnectionMap[key]
	if exists {
		delete(s.UserConnectionMap, key)
		if conn != nil {
			delete(s.ConnectionUserMap, conn)
		} else {
			return errors.New("DelUserConn conn is nil, key=" + key)
		}
		return nil
	} else {
		return errors.New("DelUserConn not found key=" + key)
	}

}

func (s *Server) DelConnUser(conn *net.TCPConn) (err error) {
	s.RwLock.Lock()
	defer s.RwLock.Unlock()
	if conn == nil {
		return errors.New("DelConnUserMap conn is nil")
	}
	key, exists := s.ConnectionUserMap[conn]
	if exists {
		delete(s.UserConnectionMap, key)
		delete(s.ConnectionUserMap, conn)
		return nil
	} else {
		return errors.New("DelUserConn not found key=" + key)
	}

}

func (s *Server) Close(conn *net.TCPConn) (err error) {
	if conn == nil {
		return errors.New("conn is nil，close err")
	}
	_ = s.DelConnUser(conn)
	_ = conn.Close()
	conn = nil
	return nil
}
