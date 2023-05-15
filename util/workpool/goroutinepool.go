// Copyright 2022 The GDP-CMP Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package workpool

import (
	"context"
	"github.com/Chairou/toolbox/util/workqueue"
	"golang.org/x/time/rate"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

// GoRoutinePool go routine pool
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

// GoRoutineFunc process func
type GoRoutineFunc func(params ...interface{}) (interface{}, error)

// FuncResult struct
type FuncResult struct {
	Result interface{}
	Err    error
}

// GoRoutineExecutor body
type GoRoutineExecutor struct {
	TaskID string
	GoRoutineFunc
	GoRoutineParams []interface{}
}

// NewRateLimitedGoRoutinePool bucketNum 默认最好为1, 这样qps才会准确
func NewRateLimitedGoRoutinePool(size int, stopCH <-chan struct{}, ctx context.Context, qps int, bucketNum int) *GoRoutinePool {
	buff := make(chan struct{}, size)
	return &GoRoutinePool{
		size:   size,
		buff:   buff,
		wg:     sync.WaitGroup{},
		stopCh: stopCH,
		ctx:    ctx,
		result: make(chan interface{}),
		queue:  workqueue.NewRateLimitingQueue(CustomControllerRateLimiter(qps, bucketNum)),
	}
}

// Submit task
func (p *GoRoutinePool) Submit(executor GoRoutineExecutor) {
	p.queue.AddRateLimited(executor.TaskID)
	p.entry.Store(executor.TaskID, executor)
	p.wg.Add(1)
	klog.Infof("submit task %v to goroutine pool", executor.TaskID)
}

// Run pool
func (p *GoRoutinePool) Run() []interface{} {
	defer workqueue.HandleCrash()
	defer p.queue.ShutDown()

	for i := 0; i < p.size; i++ {
		go wait.Until(p.worker, time.Second, p.stopCh)
	}
	var receive []interface{}
	done := make(chan interface{})
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

// run worker
func (p *GoRoutinePool) worker() {
	for p.processNextItem() {
	}
}

// do process
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
	_fc := executor.GoRoutineFunc
	_params := executor.GoRoutineParams

	result.Result, result.Err = _fc(_params...)
	if result.Err != nil {
		klog.Errorf("processNextItem %v error: %v", name, result.Err)
	}

	return true
}

func CustomControllerRateLimiter(qps int, bucketNum int) workqueue.RateLimiter {
	return workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
		// 10 qps, 100 bucket size.  This is only for retry speed and its only the overall factor (not per item)
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(qps), bucketNum)},
	)
}
