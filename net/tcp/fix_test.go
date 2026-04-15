package tcp

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"net"
	"sync"
	"testing"
	"time"
)

// ============================================================
// 辅助函数：创建 TCP 连接对
// ============================================================
func createTCPPipe(t *testing.T) (net.Conn, net.Conn) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("创建 listener 失败: %v", err)
	}

	connCh := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		connCh <- conn
	}()

	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		listener.Close()
		t.Fatalf("创建客户端连接失败: %v", err)
	}

	serverConn := <-connCh
	listener.Close()

	return clientConn, serverConn
}

// ============================================================
// 修正点: Content 函数不修改调用方底层数组
// ============================================================
func TestContent_NotModifyInput(t *testing.T) {
	svr := &Server{}
	original := []byte("hello")
	// 给 original 足够的 cap，如果 Content 直接 append 到 *input 上，会修改底层数组
	input := make([]byte, len(original), len(original)+100)
	copy(input, original)

	// 保存 input 的原始内容
	inputCopy := make([]byte, len(input))
	copy(inputCopy, input)

	output, err := Content(svr, nil, &input)
	if err != nil {
		t.Fatalf("Content() 返回错误: %v", err)
	}

	// 验证输出正确
	expected := "hello, Hello, world!"
	if string(output) != expected {
		t.Errorf("Content() 输出不正确: got %q, want %q", string(output), expected)
	}

	// 验证 input 没有被修改（修正前会被 append 修改底层数组）
	if !bytes.Equal(input, inputCopy) {
		t.Errorf("Content() 修改了输入数据: got %q, want %q", string(input), string(inputCopy))
	}
}

// ============================================================
// 修正点: unPack 中 Tag 校验 — 无效 Tag 返回错误
// ============================================================
func TestUnPack_InvalidTag(t *testing.T) {
	svr := &Server{
		Tag:              "BF",
		TagSize:          2,
		PacketLengthSize: 4,
		HeaderLength:     6,
	}

	clientConn, serverConn := createTCPPipe(t)
	defer clientConn.Close()
	defer serverConn.Close()

	// 发送一个 Tag 为 "XX" 的数据包
	go func() {
		buf := make([]byte, 0)
		buf = append(buf, []byte("XX")...)
		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, uint32(11)) // header(6) + body(5)
		buf = append(buf, lenBuf...)
		buf = append(buf, []byte("hello")...)
		clientConn.Write(buf)
	}()

	_, err := unPack(svr, serverConn.(*net.TCPConn), nil)
	if err == nil {
		t.Fatal("unPack() 应该因为无效 Tag 返回错误")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("invalid tag")) {
		t.Errorf("错误信息应包含 'invalid tag', got: %v", err)
	}
}

// ============================================================
// 修正点: unPack Tag 校验 — 合法 Tag 正常通过
// ============================================================
func TestUnPack_ValidTag(t *testing.T) {
	svr := &Server{
		Tag:              "BF",
		TagSize:          2,
		PacketLengthSize: 4,
		HeaderLength:     6,
	}

	clientConn, serverConn := createTCPPipe(t)
	defer clientConn.Close()
	defer serverConn.Close()

	expectedMsg := "hello"
	go func() {
		buf := make([]byte, 0)
		buf = append(buf, []byte("BF")...)
		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, uint32(6+len(expectedMsg)))
		buf = append(buf, lenBuf...)
		buf = append(buf, []byte(expectedMsg)...)
		clientConn.Write(buf)
	}()

	output, err := unPack(svr, serverConn.(*net.TCPConn), nil)
	if err != nil {
		t.Fatalf("unPack() 返回错误: %v", err)
	}
	if string(output) != expectedMsg {
		t.Errorf("unPack() 输出不正确: got %q, want %q", string(output), expectedMsg)
	}
}

// ============================================================
// 修正点: unPack 中 tLength < HeaderLength 下溢保护
// ============================================================
func TestUnPack_LengthUnderflow(t *testing.T) {
	svr := &Server{
		Tag:              "BF",
		TagSize:          2,
		PacketLengthSize: 4,
		HeaderLength:     6,
	}

	clientConn, serverConn := createTCPPipe(t)
	defer clientConn.Close()
	defer serverConn.Close()

	// 发送一个 tLength=3（小于 HeaderLength=6）的数据包
	go func() {
		buf := make([]byte, 0)
		buf = append(buf, []byte("BF")...)
		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, uint32(3)) // 小于 HeaderLength
		buf = append(buf, lenBuf...)
		clientConn.Write(buf)
	}()

	_, err := unPack(svr, serverConn.(*net.TCPConn), nil)
	if err == nil {
		t.Fatal("unPack() 应该因为 tLength < HeaderLength 返回错误")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("total length less than header length")) {
		t.Errorf("错误信息应包含 'total length less than header length', got: %v", err)
	}
}

// ============================================================
// 修正点: unPack 返回独立拷贝（pool 安全）
// ============================================================
func TestUnPack_ReturnIndependentCopy(t *testing.T) {
	svr := &Server{
		Tag:              "BF",
		TagSize:          2,
		PacketLengthSize: 4,
		HeaderLength:     6,
	}

	clientConn, serverConn := createTCPPipe(t)
	defer clientConn.Close()
	defer serverConn.Close()

	expectedMsg := "test_data"
	go func() {
		buf := make([]byte, 0)
		buf = append(buf, []byte("BF")...)
		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, uint32(6+len(expectedMsg)))
		buf = append(buf, lenBuf...)
		buf = append(buf, []byte(expectedMsg)...)
		clientConn.Write(buf)
	}()

	output, err := unPack(svr, serverConn.(*net.TCPConn), nil)
	if err != nil {
		t.Fatalf("unPack() 返回错误: %v", err)
	}

	// 保存返回值
	savedOutput := string(output)

	// 模拟 pool 被其他 goroutine 使用：从 pool 取出并覆盖数据
	for i := 0; i < 10; i++ {
		buf := packagePool.Get().([]byte)
		buf = buf[:cap(buf)]
		for j := range buf {
			buf[j] = 0xFF
		}
		packagePool.Put(buf)
	}

	// 验证返回值没有被 pool 操作影响
	if string(output) != savedOutput {
		t.Errorf("unPack() 返回值被 pool 操作覆盖: got %q, want %q", string(output), savedOutput)
	}
}

// ============================================================
// 修正点: unPack conn nil 检查
// ============================================================
func TestUnPack_NilConn(t *testing.T) {
	svr := &Server{}
	_, err := unPack(svr, nil, nil)
	if err == nil {
		t.Fatal("unPack(nil conn) 应该返回错误")
	}
}

// ============================================================
// 修正点: Pack conn nil 检查
// ============================================================
func TestPack_NilConn(t *testing.T) {
	svr := &Server{}
	input := []byte("hello")
	_, err := Pack(svr, nil, &input)
	if err == nil {
		t.Fatal("Pack(nil conn) 应该返回错误")
	}
}

// ============================================================
// 修正点: Pack 空 input 检查
// ============================================================
func TestPack_EmptyInput(t *testing.T) {
	svr := &Server{}
	clientConn, serverConn := createTCPPipe(t)
	defer clientConn.Close()
	defer serverConn.Close()

	input := []byte{}
	_, err := Pack(svr, serverConn.(*net.TCPConn), &input)
	if err == nil {
		t.Fatal("Pack(empty input) 应该返回错误")
	}
}

// ============================================================
// 修正点: Pack 返回独立拷贝
// ============================================================
func TestPack_ReturnIndependentCopy(t *testing.T) {
	svr := &Server{
		Tag:              "BF",
		TagSize:          2,
		PacketLengthSize: 4,
		HeaderLength:     6,
	}

	clientConn, serverConn := createTCPPipe(t)
	defer clientConn.Close()
	defer serverConn.Close()

	// 启动读取端，防止 Write 阻塞
	go func() {
		buf := make([]byte, 65535)
		for {
			_, err := clientConn.Read(buf)
			if err != nil {
				return
			}
		}
	}()

	input := []byte("hello")
	output, err := Pack(svr, serverConn.(*net.TCPConn), &input)
	if err != nil {
		t.Fatalf("Pack() 返回错误: %v", err)
	}

	savedOutput := make([]byte, len(output))
	copy(savedOutput, output)

	// 模拟 pool 被其他 goroutine 使用
	for i := 0; i < 10; i++ {
		buf := packagePool.Get().([]byte)
		buf = buf[:cap(buf)]
		for j := range buf {
			buf[j] = 0xFF
		}
		packagePool.Put(buf)
	}

	// 验证返回值没有被 pool 操作影响
	if !bytes.Equal(output, savedOutput) {
		t.Errorf("Pack() 返回值被 pool 操作覆盖")
	}
}

// ============================================================
// 修正点: Pack/unPack 完整 TLV 流程（2/4/8 字节长度）
// ============================================================
func TestTLV_PackUnPack_AllLengthSizes(t *testing.T) {
	testCases := []struct {
		name             string
		packetLengthSize int
	}{
		{"2字节长度", 2},
		{"4字节长度", 4},
		{"8字节长度", 8},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svr := &Server{
				Tag:              "BF",
				TagSize:          2,
				PacketLengthSize: tc.packetLengthSize,
				HeaderLength:     2 + tc.packetLengthSize,
			}

			clientConn, serverConn := createTCPPipe(t)
			defer clientConn.Close()
			defer serverConn.Close()

			expectedMsg := "hello_world"

			// 客户端发送 Pack 数据
			go func() {
				input := []byte(expectedMsg)
				_, err := Pack(svr, clientConn.(*net.TCPConn), &input)
				if err != nil {
					t.Errorf("Pack() 错误: %v", err)
				}
			}()

			// 服务端 unPack 读取
			output, err := unPack(svr, serverConn.(*net.TCPConn), nil)
			if err != nil {
				t.Fatalf("unPack() 错误: %v", err)
			}
			if string(output) != expectedMsg {
				t.Errorf("TLV 流程结果不正确: got %q, want %q", string(output), expectedMsg)
			}
		})
	}
}

// ============================================================
// 修正点: readUntilEndMarker OOM 保护
// ============================================================
func TestReadUntilEndMarker_OOMProtection(t *testing.T) {
	svr := &Server{
		EndMarker: []byte("\r\n"),
	}

	clientConn, serverConn := createTCPPipe(t)
	defer clientConn.Close()
	defer serverConn.Close()

	// 发送超过 MAX_PACKAGE_LENGTH 的数据但不发送结束符
	go func() {
		data := bytes.Repeat([]byte("A"), MAX_PACKAGE_LENGTH+100)
		clientConn.Write(data)
		clientConn.Close()
	}()

	_, err := readUntilEndMarker(svr, serverConn.(*net.TCPConn), nil)
	if err == nil {
		t.Fatal("readUntilEndMarker() 应该因为超过 MAX_PACKAGE_LENGTH 返回错误")
	}
}

// ============================================================
// 修正点: readUntilEndMarker 返回独立拷贝
// ============================================================
func TestReadUntilEndMarker_ReturnIndependentCopy(t *testing.T) {
	svr := &Server{
		EndMarker: []byte("\r\n"),
	}

	clientConn, serverConn := createTCPPipe(t)
	defer clientConn.Close()
	defer serverConn.Close()

	expectedData := "hello\r\n"
	go func() {
		clientConn.Write([]byte(expectedData))
	}()

	output, err := readUntilEndMarker(svr, serverConn.(*net.TCPConn), nil)
	if err != nil {
		t.Fatalf("readUntilEndMarker() 返回错误: %v", err)
	}

	savedOutput := string(output)

	// 模拟 pool 被其他 goroutine 使用
	for i := 0; i < 10; i++ {
		buf := packagePool.Get().([]byte)
		buf = buf[:cap(buf)]
		for j := range buf {
			buf[j] = 0xFF
		}
		packagePool.Put(buf)
	}

	if string(output) != savedOutput {
		t.Errorf("readUntilEndMarker() 返回值被 pool 操作覆盖: got %q, want %q", string(output), savedOutput)
	}
}

// ============================================================
// 修正点: readUntilEndMarker nil conn 检查
// ============================================================
func TestReadUntilEndMarker_NilConn(t *testing.T) {
	svr := &Server{EndMarker: []byte("\r\n")}
	_, err := readUntilEndMarker(svr, nil, nil)
	if err == nil {
		t.Fatal("readUntilEndMarker(nil conn) 应该返回错误")
	}
}

// ============================================================
// 修正点: readUntilEndMarker 正常流程（分片发送）
// ============================================================
func TestReadUntilEndMarker_Normal(t *testing.T) {
	svr := &Server{
		EndMarker: []byte("\r\n"),
	}

	clientConn, serverConn := createTCPPipe(t)
	defer clientConn.Close()
	defer serverConn.Close()

	go func() {
		// 分多次发送，模拟真实网络场景
		clientConn.Write([]byte("hel"))
		time.Sleep(10 * time.Millisecond)
		clientConn.Write([]byte("lo\r\n"))
	}()

	output, err := readUntilEndMarker(svr, serverConn.(*net.TCPConn), nil)
	if err != nil {
		t.Fatalf("readUntilEndMarker() 返回错误: %v", err)
	}
	if string(output) != "hello\r\n" {
		t.Errorf("readUntilEndMarker() 输出不正确: got %q, want %q", string(output), "hello\r\n")
	}
}

// ============================================================
// 修正点: writeWithEndMark 不修改 *input 底层数组
// ============================================================
func TestWriteWithEndMark_NotModifyInput(t *testing.T) {
	svr := &Server{
		EndMarker: []byte("\r\n"),
	}

	clientConn, serverConn := createTCPPipe(t)
	defer clientConn.Close()
	defer serverConn.Close()

	// 启动读取端
	go func() {
		buf := make([]byte, 65535)
		for {
			_, err := clientConn.Read(buf)
			if err != nil {
				return
			}
		}
	}()

	// 给 input 足够的 cap，如果 writeWithEndMark 直接 append 到 *input，会修改底层数组
	original := []byte("hello")
	input := make([]byte, len(original), len(original)+100)
	copy(input, original)

	inputCopy := make([]byte, len(input))
	copy(inputCopy, input)

	_, err := writeWithEndMark(svr, serverConn.(*net.TCPConn), &input)
	if err != nil {
		t.Fatalf("writeWithEndMark() 返回错误: %v", err)
	}

	// 验证 input 没有被修改
	if !bytes.Equal(input, inputCopy) {
		t.Errorf("writeWithEndMark() 修改了输入数据: got %q, want %q", string(input), string(inputCopy))
	}
}

// ============================================================
// 修正点: writeWithEndMark 返回独立拷贝
// ============================================================
func TestWriteWithEndMark_ReturnIndependentCopy(t *testing.T) {
	svr := &Server{
		EndMarker: []byte("\r\n"),
	}

	clientConn, serverConn := createTCPPipe(t)
	defer clientConn.Close()
	defer serverConn.Close()

	go func() {
		buf := make([]byte, 65535)
		for {
			_, err := clientConn.Read(buf)
			if err != nil {
				return
			}
		}
	}()

	input := []byte("hello")
	output, err := writeWithEndMark(svr, serverConn.(*net.TCPConn), &input)
	if err != nil {
		t.Fatalf("writeWithEndMark() 返回错误: %v", err)
	}

	savedOutput := make([]byte, len(output))
	copy(savedOutput, output)

	// 模拟 pool 被其他 goroutine 使用
	for i := 0; i < 10; i++ {
		buf := packagePool.Get().([]byte)
		buf = buf[:cap(buf)]
		for j := range buf {
			buf[j] = 0xFF
		}
		packagePool.Put(buf)
	}

	if !bytes.Equal(output, savedOutput) {
		t.Errorf("writeWithEndMark() 返回值被 pool 操作覆盖")
	}
}

// ============================================================
// 修正点: writeWithEndMark nil conn 检查
// ============================================================
func TestWriteWithEndMark_NilConn(t *testing.T) {
	svr := &Server{EndMarker: []byte("\r\n")}
	input := []byte("hello")
	_, err := writeWithEndMark(svr, nil, &input)
	if err == nil {
		t.Fatal("writeWithEndMark(nil conn) 应该返回错误")
	}
}

// ============================================================
// 修正点: writeWithEndMark nil input 只写结束符
// ============================================================
func TestWriteWithEndMark_NilInput(t *testing.T) {
	svr := &Server{
		EndMarker: []byte("\r\n"),
	}

	clientConn, serverConn := createTCPPipe(t)
	defer clientConn.Close()
	defer serverConn.Close()

	// 读取端
	resultCh := make(chan []byte, 1)
	go func() {
		buf := make([]byte, 65535)
		n, _ := clientConn.Read(buf)
		resultCh <- buf[:n]
	}()

	output, err := writeWithEndMark(svr, serverConn.(*net.TCPConn), nil)
	if err != nil {
		t.Fatalf("writeWithEndMark(nil input) 返回错误: %v", err)
	}

	// 验证只写入了结束符
	if string(output) != "\r\n" {
		t.Errorf("writeWithEndMark(nil input) 输出不正确: got %q, want %q", string(output), "\r\n")
	}

	received := <-resultCh
	if string(received) != "\r\n" {
		t.Errorf("接收到的数据不正确: got %q, want %q", string(received), "\r\n")
	}
}

// ============================================================
// 修正点: compress 返回独立拷贝
// ============================================================
func TestCompress_ReturnIndependentCopy(t *testing.T) {
	input := []byte("hello world, this is a test for compress")
	output, err := compress(nil, nil, &input)
	if err != nil {
		t.Fatalf("compress() 返回错误: %v", err)
	}

	savedOutput := make([]byte, len(output))
	copy(savedOutput, output)

	// 模拟 pool 被其他 goroutine 使用：取出 writer pool 并 Reset
	for i := 0; i < 5; i++ {
		poolItem := flateWriterPool.Get().(struct {
			buf *bytes.Buffer
			w   *flate.Writer
		})
		poolItem.buf.Reset()
		flateWriterPool.Put(poolItem)
	}

	// 验证返回值没有被 pool 操作影响
	if !bytes.Equal(output, savedOutput) {
		t.Errorf("compress() 返回值被 pool 操作覆盖")
	}
}

// ============================================================
// 修正点: compress/unCompress 往返测试
// ============================================================
func TestCompressUnCompress_RoundTrip(t *testing.T) {
	testData := []string{
		"hello",
		"hello world, this is a longer test string for compression",
		"abcdefghijklmnopqrstuvwxyz0123456789",
	}

	for _, data := range testData {
		input := []byte(data)
		compressed, err := compress(nil, nil, &input)
		if err != nil {
			t.Fatalf("compress(%q) 返回错误: %v", data, err)
		}

		decompressed, err := unCompress(nil, nil, &compressed)
		if err != nil {
			t.Fatalf("unCompress() 返回错误: %v", err)
		}

		if string(decompressed) != data {
			t.Errorf("压缩/解压往返测试失败: got %q, want %q", string(decompressed), data)
		}
	}
}

// ============================================================
// 修正点: compress/unCompress nil input 检查
// ============================================================
func TestCompress_NilInput(t *testing.T) {
	_, err := compress(nil, nil, nil)
	if err == nil {
		t.Fatal("compress(nil) 应该返回错误")
	}
}

func TestUnCompress_NilInput(t *testing.T) {
	_, err := unCompress(nil, nil, nil)
	if err == nil {
		t.Fatal("unCompress(nil) 应该返回错误")
	}
}

// ============================================================
// 修正点: process 中 result 拷贝安全（管道数据传递）
// ============================================================
func TestProcess_ResultCopySafety(t *testing.T) {
	svr := &Server{}

	// 定义一个 operation 链：第一个函数返回数据，第二个函数验证数据完整性
	var capturedInput []byte
	op1 := func(svr *Server, conn *net.TCPConn, input *[]byte) ([]byte, error) {
		return []byte("step1_output"), nil
	}
	op2 := func(svr *Server, conn *net.TCPConn, input *[]byte) ([]byte, error) {
		capturedInput = make([]byte, len(*input))
		copy(capturedInput, *input)
		return []byte("step2_output"), nil
	}

	svr.OperationList = []operation{op1, op2}
	err := svr.process(svr.OperationList, svr, nil, nil)
	if err != nil {
		t.Fatalf("process() 返回错误: %v", err)
	}

	// 验证第二个函数收到的输入是第一个函数的输出
	if string(capturedInput) != "step1_output" {
		t.Errorf("process 管道传递数据不正确: got %q, want %q", string(capturedInput), "step1_output")
	}
}

// ============================================================
// 修正点: process 三步管道数据完整性
// ============================================================
func TestProcess_ThreeStepPipeline(t *testing.T) {
	svr := &Server{}

	var step2Input, step3Input string

	op1 := func(svr *Server, conn *net.TCPConn, input *[]byte) ([]byte, error) {
		return []byte("aaa"), nil
	}
	op2 := func(svr *Server, conn *net.TCPConn, input *[]byte) ([]byte, error) {
		step2Input = string(*input)
		return []byte("bbb"), nil
	}
	op3 := func(svr *Server, conn *net.TCPConn, input *[]byte) ([]byte, error) {
		step3Input = string(*input)
		return []byte("ccc"), nil
	}

	svr.OperationList = []operation{op1, op2, op3}
	err := svr.process(svr.OperationList, svr, nil, nil)
	if err != nil {
		t.Fatalf("process() 返回错误: %v", err)
	}

	if step2Input != "aaa" {
		t.Errorf("step2 收到的输入不正确: got %q, want %q", step2Input, "aaa")
	}
	if step3Input != "bbb" {
		t.Errorf("step3 收到的输入不正确: got %q, want %q", step3Input, "bbb")
	}
}

// ============================================================
// 修正点: Close 方法
// ============================================================
func TestClose_NilConn(t *testing.T) {
	svr := &Server{}
	err := svr.Close(nil)
	if err == nil {
		t.Fatal("Close(nil) 应该返回错误")
	}
}

func TestClose_ValidConn(t *testing.T) {
	svr := &Server{}
	clientConn, serverConn := createTCPPipe(t)
	defer clientConn.Close()

	err := svr.Close(serverConn.(*net.TCPConn))
	if err != nil {
		t.Fatalf("Close() 返回错误: %v", err)
	}

	// 验证连接已关闭：再次写入应该失败
	_, err = serverConn.Write([]byte("test"))
	if err == nil {
		t.Error("连接关闭后写入应该失败")
	}
}

// ============================================================
// 修正点: NewTcpServerOption 参数校验
// ============================================================
func TestNewTcpServerOption_InvalidPacketLengthSize(t *testing.T) {
	opt := ServerOption{
		Type:             TYPE_TLV,
		PacketLengthSize: 3, // 无效值
	}
	_, err := NewTcpServerOption("127.0.0.1:0", opt)
	if err == nil {
		t.Fatal("NewTcpServerOption(PacketLengthSize=3) 应该返回错误")
	}
}

func TestNewTcpServerOption_InvalidType(t *testing.T) {
	opt := ServerOption{
		Type: 99, // 无效类型
	}
	_, err := NewTcpServerOption("127.0.0.1:0", opt)
	if err == nil {
		t.Fatal("NewTcpServerOption(Type=99) 应该返回错误")
	}
}

func TestNewTcpServerOption_EndMarkEmpty(t *testing.T) {
	opt := ServerOption{
		Type:      TYPE_ENDMARK,
		EndMarker: []byte{}, // 空结束符
	}
	_, err := NewTcpServerOption("127.0.0.1:0", opt)
	if err == nil {
		t.Fatal("NewTcpServerOption(empty EndMarker) 应该返回错误")
	}
}

func TestNewTcpServerOption_TLV_DefaultTag(t *testing.T) {
	opt := ServerOption{
		Type:             TYPE_TLV,
		PacketLengthSize: 4,
		// Tag 为空，应使用默认值 "BF"
	}
	svr, err := NewTcpServerOption("127.0.0.1:0", opt)
	if err != nil {
		t.Fatalf("NewTcpServerOption() 返回错误: %v", err)
	}
	defer svr.Listener.Close()

	if svr.Tag != "BF" {
		t.Errorf("默认 Tag 应为 'BF', got %q", svr.Tag)
	}
	if svr.HeaderLength != 6 {
		t.Errorf("HeaderLength 应为 6, got %d", svr.HeaderLength)
	}
}

func TestNewTcpServerOption_TLV_CustomTag(t *testing.T) {
	opt := ServerOption{
		Type:             TYPE_TLV,
		Tag:              "FF",
		PacketLengthSize: 4,
	}
	svr, err := NewTcpServerOption("127.0.0.1:0", opt)
	if err != nil {
		t.Fatalf("NewTcpServerOption() 返回错误: %v", err)
	}
	defer svr.Listener.Close()

	if svr.Tag != "FF" {
		t.Errorf("Tag 应为 'FF', got %q", svr.Tag)
	}
}

// ============================================================
// 修正点: EndMarker 完整读写流程
// ============================================================
func TestEndMarker_WriteAndRead(t *testing.T) {
	svr := &Server{
		EndMarker: []byte("\r\n"),
	}

	clientConn, serverConn := createTCPPipe(t)
	defer clientConn.Close()
	defer serverConn.Close()

	expectedMsg := "hello world"
	// 写入端
	go func() {
		input := []byte(expectedMsg)
		_, err := writeWithEndMark(svr, clientConn.(*net.TCPConn), &input)
		if err != nil {
			t.Errorf("writeWithEndMark() 错误: %v", err)
		}
	}()

	// 读取端
	output, err := readUntilEndMarker(svr, serverConn.(*net.TCPConn), nil)
	if err != nil {
		t.Fatalf("readUntilEndMarker() 返回错误: %v", err)
	}

	// 输出应包含原始数据 + 结束符
	expected := expectedMsg + "\r\n"
	if string(output) != expected {
		t.Errorf("EndMarker 读写流程结果不正确: got %q, want %q", string(output), expected)
	}
}

// ============================================================
// 修正点: 并发安全测试 — unPack 返回值不被 pool 覆盖
// ============================================================
func TestUnPack_ConcurrentPoolSafety(t *testing.T) {
	svr := &Server{
		Tag:              "BF",
		TagSize:          2,
		PacketLengthSize: 4,
		HeaderLength:     6,
	}

	const goroutines = 10
	var wg sync.WaitGroup
	errCh := make(chan string, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			clientConn, serverConn := createTCPPipe(t)
			defer clientConn.Close()
			defer serverConn.Close()

			msg := bytes.Repeat([]byte{byte('A' + id%26)}, 100)
			go func() {
				buf := make([]byte, 0)
				buf = append(buf, []byte("BF")...)
				lenBuf := make([]byte, 4)
				binary.BigEndian.PutUint32(lenBuf, uint32(6+len(msg)))
				buf = append(buf, lenBuf...)
				buf = append(buf, msg...)
				clientConn.Write(buf)
			}()

			output, err := unPack(svr, serverConn.(*net.TCPConn), nil)
			if err != nil {
				errCh <- err.Error()
				return
			}

			// 短暂等待，让其他 goroutine 有机会操作 pool
			time.Sleep(time.Millisecond)

			if !bytes.Equal(output, msg) {
				errCh <- "数据不匹配"
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for errMsg := range errCh {
		t.Errorf("并发测试失败: %s", errMsg)
	}
}

// ============================================================
// 修正点: compress 并发安全测试
// ============================================================
func TestCompress_ConcurrentSafety(t *testing.T) {
	const goroutines = 10
	var wg sync.WaitGroup
	errCh := make(chan string, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			original := bytes.Repeat([]byte{byte('A' + id%26)}, 200)
			input := make([]byte, len(original))
			copy(input, original)

			compressed, err := compress(nil, nil, &input)
			if err != nil {
				errCh <- "compress error: " + err.Error()
				return
			}

			// 短暂等待，让其他 goroutine 有机会操作 pool
			time.Sleep(time.Millisecond)

			decompressed, err := unCompress(nil, nil, &compressed)
			if err != nil {
				errCh <- "unCompress error: " + err.Error()
				return
			}

			if !bytes.Equal(decompressed, original) {
				errCh <- "压缩/解压数据不匹配"
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for errMsg := range errCh {
		t.Errorf("并发测试失败: %s", errMsg)
	}
}

// ============================================================
// 修正点: 完整 TLV 端到端集成测试（Server 启动 + 客户端通信）
// ============================================================
func TestTLV_EndToEnd_Integration(t *testing.T) {
	opt := ServerOption{
		Type:             TYPE_TLV,
		Tag:              "BF",
		PacketLengthSize: PACKAGE_LENGTH_FOUR_BYTE,
	}

	svr, err := NewTcpServerOption("127.0.0.1:0", opt)
	if err != nil {
		t.Fatalf("创建服务器失败: %v", err)
	}
	defer svr.Listener.Close()

	svr.OperationList = []operation{unPack, Content, Pack}

	// 启动服务器
	go svr.Run()

	// 等待服务器就绪
	time.Sleep(50 * time.Millisecond)

	// 客户端连接
	conn, err := net.Dial("tcp", svr.Listener.Addr().String())
	if err != nil {
		t.Fatalf("客户端连接失败: %v", err)
	}
	defer conn.Close()

	// 构造 TLV 数据包发送
	msg := "hello"
	sendBuf := make([]byte, 0)
	sendBuf = append(sendBuf, []byte("BF")...)
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(6+len(msg)))
	sendBuf = append(sendBuf, lenBuf...)
	sendBuf = append(sendBuf, []byte(msg)...)
	_, err = conn.Write(sendBuf)
	if err != nil {
		t.Fatalf("发送数据失败: %v", err)
	}

	// 读取服务器响应的 TLV 数据包
	headerBuf := make([]byte, 6)
	_, err = conn.Read(headerBuf)
	if err != nil {
		t.Fatalf("读取响应头失败: %v", err)
	}

	// 校验 Tag
	if string(headerBuf[:2]) != "BF" {
		t.Errorf("响应 Tag 不正确: got %q, want %q", string(headerBuf[:2]), "BF")
	}

	// 解析长度
	respLen := binary.BigEndian.Uint32(headerBuf[2:6])
	bodyLen := respLen - 6
	bodyBuf := make([]byte, bodyLen)
	_, err = conn.Read(bodyBuf)
	if err != nil {
		t.Fatalf("读取响应体失败: %v", err)
	}

	// 验证响应内容
	expected := "hello, Hello, world!"
	if string(bodyBuf) != expected {
		t.Errorf("端到端响应不正确: got %q, want %q", string(bodyBuf), expected)
	}
}
