package test

import (
	"fmt"
	"sync"
	"testing"
)
import "github.com/panjf2000/ants/v2"



func myFunc(i interface{}) {
	n := i.(int32)
	fmt.Printf("run with %d\n", n )
}


func TestGoroutine(t *testing.T) {
	wg := sync.WaitGroup{}
	runTimes := 100
	// 初始化协程池
	p, _ := ants.NewPoolWithFunc(10, func(i interface{} ) {
		myFunc(i)
		wg.Done()
	})
	// 释放协程池
	defer p.Release()
	// 提交任务
	for i  := 0; i < runTimes; i++ {
		wg.Add(1)
		_ = p.Invoke(int32(i))
	}
	wg.Wait()
	fmt.Printf("running goroutines: %d\n", p.Running())
}
