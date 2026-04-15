package tcp

import (
	"bufio"
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
	Type             int
	Listener         net.Listener
	IpPort           string
	OperationList    []operation
	Tag              string
	TagBytes         []byte // 预计算的 Tag 字节切片，避免每次 string→[]byte 转换
	HeaderLength     int
	TagSize          int
	PacketLengthSize int
	EndMarker        []byte
}

// connContext 每个连接的上下文，包含带缓冲的 reader/writer，避免每次 I/O 都系统调用
type connContext struct {
	conn   *net.TCPConn
	reader *bufio.Reader
	writer *bufio.Writer
}

// connCtxMap 全局连接上下文映射，通过 conn 指针查找对应的 bufio reader/writer
var connCtxMap sync.Map

// helloSuffix 预分配常量字节切片，避免每次调用时 string→[]byte 转换
var helloSuffix = []byte(", Hello, world!")

func Content(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	inputLen := len(*input)
	result := make([]byte, inputLen+len(helloSuffix))
	copy(result, *input)
	copy(result[inputLen:], helloSuffix)
	return result, nil
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

	srv.Listener, err = srv.Listen()
	if err != nil {
		fmt.Println("NewTcpServer err, ipv4 like 192.168.0.250:8080 ")
		fmt.Println("NewTcpServer err, ipv6 like [2001:0db8:86a3:08d3:1319:8a2e:0370:7344]:8080")
		fmt.Println("NewTcpServer err, both ipv4 and ipv6 like 0:8080")
		return nil, err
	}
	return srv, nil
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 4096)
	},
}

// connCtxPool 复用 connContext 中的 bufio.Reader/Writer
var readerPool = sync.Pool{
	New: func() interface{} {
		return bufio.NewReaderSize(nil, 8192)
	},
}

var writerPool = sync.Pool{
	New: func() interface{} {
		return bufio.NewWriterSize(nil, 8192)
	},
}

// getConnCtx 获取连接上下文，如果不存在则返回 nil（兼容非 Server 管理的连接）
func getConnCtx(conn *net.TCPConn) *connContext {
	if v, ok := connCtxMap.Load(conn); ok {
		return v.(*connContext)
	}
	return nil
}

// 定义处理过程函数，前一个处理函数的输出是后一个处理函数的输入
func (s *Server) process(functions []operation, t *Server, conn *net.TCPConn, input *[]byte) error {
	var err error
	var result []byte
	if input != nil {
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
	// 设置 TCP_NODELAY，禁用 Nagle 算法，减少小包延迟
	conn.SetNoDelay(true)

	// 为每个连接创建带缓冲的 reader/writer，减少系统调用次数
	reader := readerPool.Get().(*bufio.Reader)
	reader.Reset(conn)

	writer := writerPool.Get().(*bufio.Writer)
	writer.Reset(conn)

	ctx := &connContext{
		conn:   conn,
		reader: reader,
		writer: writer,
	}
	// 注册到全局映射，让 operation 函数可以通过 conn 查找 bufio reader/writer
	connCtxMap.Store(conn, ctx)

	defer func() {
		connCtxMap.Delete(conn)
		writer.Flush()
		readerPool.Put(reader)
		writerPool.Put(writer)
		err := conn.Close()
		if err != nil {
			fmt.Println("Close Connection, err:" + err.Error())
		}
	}()

	for {
		err := s.process(s.OperationList, s, conn, nil)
		if err != nil {
			return
		}
	}
}

func (s *Server) SetTag(tag string) {
	s.Tag = tag
	s.TagBytes = []byte(tag)
	s.TagSize = len(tag)
}

func (s *Server) Close(conn *net.TCPConn) (err error) {
	if conn == nil {
		return errors.New("conn is nil，close err")
	}
	return conn.Close()
}
