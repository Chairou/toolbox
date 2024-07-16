package tcp

import (
	"fmt"
	"testing"
)

func TestTcpTlvServer(t *testing.T) {
	opt := ServerOption{}
	opt.Tag = "FF"
	opt.PacketLengthSize = PACKAGE_LENGTH_FOUR_BYTE
	opt.Type = TYPE_TLV

	svr, err := NewTcpServerOption("127.0.0.1:8080", opt)
	if err != nil {
		fmt.Println("Error creating tcp server: ", err)
	}
	svr.OperationList = make([]operation, 0)
	svr.OperationList = append(svr.OperationList, unPack, unCompress, Content, compress, Pack)
	err = svr.Run()
	if err != nil {
		fmt.Println("Error running tcp server: ", err)
		return
	}
}

func TestTcpEndMarkServer(t *testing.T) {
	opt := ServerOption{}
	opt.EndMarker = []byte("\r\n")
	opt.Type = TYPE_ENDMARK

	svr, err := NewTcpServerOption("127.0.0.1:8080", opt)
	if err != nil {
		fmt.Println("Error creating tcp server: ", err)
	}
	svr.OperationList = make([]operation, 0, 16)
	svr.OperationList = append(svr.OperationList, readUntilEndMarker, unCompress, Content, compress, writeWithEndMark)
	err = svr.Run()
	if err != nil {
		fmt.Println("Error running tcp server: ", err)
		return
	}
}
