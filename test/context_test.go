package test

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	go subProcess(ctx)
	time.Sleep(5* time.Second)
	cancel()
	fmt.Println("MAIN DONE")
	time.Sleep(100* time.Second)

}


func subProcess(ctx context.Context) {
	for true {
		select {
		case <-ctx.Done():
			fmt.Println("subProcess ctx.Done()", ctx.Err())
			return
		default:
			time.Sleep(1* time.Second)
			fmt.Println("sub loop")
		}

	}
}
