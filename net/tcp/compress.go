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
	defer flateWriterPool.Put(poolItem)

	poolItem.buf.Reset()
	poolItem.w.Reset(poolItem.buf)

	_, err = poolItem.w.Write(*input)
	if err != nil {
		return nil, err
	}
	err = poolItem.w.Close()
	if err != nil {
		return nil, err
	}

	return poolItem.buf.Bytes(), nil
}

// 解压插件
func unCompress(svr *Server, conn *net.TCPConn, input *[]byte) (output []byte, err error) {
	if input == nil {
		return nil, errors.New("input is nil")
	}

	r := flateReaderPool.Get().(io.ReadCloser)
	defer flateReaderPool.Put(r)

	err = r.(flate.Resetter).Reset(bytes.NewReader(*input), nil)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
