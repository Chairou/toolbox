package tcp

import (
	"bytes"
	"compress/flate"
	"errors"
	"io"
	"net"
	"sync"
)

var flateWriterPool = sync.Pool{
	New: func() interface{} {
		buf := bytes.NewBuffer(nil)
		w, _ := flate.NewWriter(buf, flate.DefaultCompression)
		return struct {
			buf *bytes.Buffer
			w   *flate.Writer
		}{buf, w}
	},
}

var flateReaderPool = sync.Pool{
	New: func() interface{} {
		return flate.NewReader(nil)
	},
}

// 压缩插件
func compress(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	if input == nil {
		return nil, errors.New("input is nil")
	}

	poolItem := flateWriterPool.Get().(struct {
		buf *bytes.Buffer
		w   *flate.Writer
	})

	poolItem.buf.Reset()
	poolItem.w.Reset(poolItem.buf)

	_, err = poolItem.w.Write(*input)
	if err != nil {
		flateWriterPool.Put(poolItem)
		return nil, err
	}
	err = poolItem.w.Close()
	if err != nil {
		flateWriterPool.Put(poolItem)
		return nil, err
	}

	// 拷贝一份独立数据返回，避免 pool 回收后 buffer 被 Reset 覆盖
	result := make([]byte, poolItem.buf.Len())
	copy(result, poolItem.buf.Bytes())
	flateWriterPool.Put(poolItem)
	return result, nil
}

// 解压插件
func unCompress(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	if input == nil {
		return nil, errors.New("input is nil")
	}

	r := flateReaderPool.Get().(io.ReadCloser)

	err = r.(flate.Resetter).Reset(bytes.NewReader(*input), nil)
	if err != nil {
		flateReaderPool.Put(r)
		return nil, err
	}

	// 预分配缓冲区，避免 bytes.Buffer 多次扩容
	buf := bytes.NewBuffer(make([]byte, 0, len(*input)*2))
	_, err = io.Copy(buf, r)
	flateReaderPool.Put(r)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
