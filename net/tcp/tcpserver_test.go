package tcp

import (
	"fmt"
	"testing"
)

func TestTcpServer(t *testing.T) {
	svr, err := NewTcpServer("127.0.0.1:8080")
	if err != nil {
		fmt.Println("Error creating tcp server: ", err)
	}
	svr.TagSize = 2
	svr.LengthSize = 4
	svr.HeaderLength = 2 + 4
	svr.OperationList = make([]operation, 0, 16)
	svr.OperationList = append(svr.OperationList, unPack, Content, Pack)
	err = svr.Run()
	if err != nil {
		fmt.Println("Error running tcp server: ", err)
		return
	}

}
