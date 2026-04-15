// Package workpool 提供基于速率限制的 goroutine 池实现，支持任务提交和并发执行
package workpool

import (
	"context"
	"sync"
	"time"

	"github.com/Chairou/toolbox/util/workqueue"
	"github.com/Chairou/toolbox/util/workqueue/runtime"
	"github.com/Chairou/toolbox/util/workqueue/wait"
	"golang.org/x/time/rate"
	"k8s.io/klog/v2"
)

// GoRoutinePool 基于速率限制的 goroutine 池，支持任务提交和并发执行。
// 注意：Run() 方法只能调用一次，调用后池不可复用
type GoRoutinePool struct {
	size   int
	wg     sync.WaitGroup
	stopCh <-chan struct{}
	ctx    context.Context
	queue  workqueue.RateLimitingInterface
	entry  sync.Map
	buff   chan struct{}
	result chan interface{}
}

// GoRoutineFunc 任务处理函数类型
type GoRoutineFunc func(params ...interface{}) (interface{}, error)

// FuncResult 任务执行结果
type FuncResult struct {
	Result interface{}
	Err    error
}

// GoRoutineExecutor 任务执行体，包含任务ID、处理函数和参数
type GoRoutineExecutor struct {
	TaskID string
	GoRoutineFunc
	GoRoutineParams []interface{}
}

// NewRateLimitedGoRoutinePool 创建一个带速率限制的 goroutine 池。
// bucketNum 默认最好为1，这样 qps 才会准确。
// size 为并发 worker 数量，qps 为每秒允许的请求数
func NewRateLimitedGoRoutinePool(size int, stopCh <-chan struct{}, ctx context.Context, qps int, bucketNum int) *GoRoutinePool {
	if size <= 0 {
		size = 1
	}
	if qps <= 0 {
		qps = 1
	}
	if bucketNum <= 0 {
		bucketNum = 1
	}
	buff := make(chan struct{}, size)
	return &GoRoutinePool{
		size:   size,
		buff:   buff,
		wg:     sync.WaitGroup{},
		stopCh: stopCh,
		ctx:    ctx,
		result: make(chan interface{}, size),
		queue:  workqueue.NewRateLimitingQueue(CustomControllerRateLimiter(qps, bucketNum)),
	}
}

// Submit 提交任务到 goroutine 池
func (p *GoRoutinePool) Submit(executor GoRoutineExecutor) {
	p.queue.AddRateLimited(executor.TaskID)
	p.entry.Store(executor.TaskID, executor)
	p.wg.Add(1)
	klog.Infof("submit task %v to goroutine pool", executor.TaskID)
}

// Run 启动 goroutine 池并等待所有任务完成，返回所有任务的执行结果。
// 注意：此方法只能调用一次，调用后池不可复用
func (p *GoRoutinePool) Run() []interface{} {
	defer runtime.HandleCrash()
	defer p.queue.ShutDown()

	for i := 0; i < p.size; i++ {
		go wait.Until(p.worker, time.Second, p.stopCh)
	}
	var receive []interface{}
	done := make(chan struct{})
	go func() {
		for rs := range p.result {
			receive = append(receive, rs)
		}
		close(done)
	}()

	p.wg.Wait()
	close(p.result)
	<-done

	return receive
}

// worker 从队列中循环取出任务并执行
func (p *GoRoutinePool) worker() {
	for p.processNextItem() {
	}
}

// processNextItem 从队列中取出一个任务并执行，返回 false 表示队列已关闭
func (p *GoRoutinePool) processNextItem() bool {
	result := FuncResult{}
	name, quit := p.queue.Get()
	if quit {
		return false
	}
	klog.Infof("goroutine pool exec task: %v", name)
	p.buff <- struct{}{}
	entry, ok := p.entry.Load(name)
	defer func() {
		p.queue.Done(name)
		p.result <- result
		p.entry.Delete(name)
		<-p.buff
		p.wg.Done()
	}()

	if !ok {
		klog.Errorf("processNextItem not found key %v", name)
		return true
	}
	executor := entry.(GoRoutineExecutor)
	fc := executor.GoRoutineFunc
	params := executor.GoRoutineParams

	result.Result, result.Err = fc(params...)
	if result.Err != nil {
		klog.Errorf("processNextItem %v error: %v", name, result.Err)
	}

	return true
}

// CustomControllerRateLimiter 创建自定义的速率限制器，结合指数退避和令牌桶算法
func CustomControllerRateLimiter(qps int, bucketNum int) workqueue.RateLimiter {
	return workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(qps), bucketNum)},
	)
}
