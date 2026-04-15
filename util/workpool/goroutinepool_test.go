package workpool

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// ==================== 辅助函数 ====================

func getConstValue(param ...interface{}) (interface{}, error) {
	return 1, nil
}

func getSum(param ...interface{}) (interface{}, error) {
	sum := 0
	for _, p := range param {
		sum += p.(int)
	}
	return sum, nil
}

func alwaysError(param ...interface{}) (interface{}, error) {
	return nil, errors.New("task failed")
}

// ==================== NewRateLimitedGoRoutinePool 测试 ====================

func TestNewRateLimitedGoRoutinePool_Basic(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})
	defer close(stop)

	p := NewRateLimitedGoRoutinePool(5, stop, ctx, 10, 1)
	if p == nil {
		t.Fatal("NewRateLimitedGoRoutinePool returned nil")
	}
	if p.size != 5 {
		t.Errorf("expected size=5, got %d", p.size)
	}
}

func TestNewRateLimitedGoRoutinePool_InvalidSize(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})
	defer close(stop)

	// size <= 0 应被修正为 1
	p := NewRateLimitedGoRoutinePool(0, stop, ctx, 10, 1)
	if p == nil {
		t.Fatal("NewRateLimitedGoRoutinePool returned nil")
	}
	if p.size != 1 {
		t.Errorf("expected size=1 for invalid input, got %d", p.size)
	}

	p2 := NewRateLimitedGoRoutinePool(-5, stop, ctx, 10, 1)
	if p2.size != 1 {
		t.Errorf("expected size=1 for negative input, got %d", p2.size)
	}
}

func TestNewRateLimitedGoRoutinePool_InvalidQPS(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})
	defer close(stop)

	// qps <= 0 不应 panic
	p := NewRateLimitedGoRoutinePool(5, stop, ctx, 0, 1)
	if p == nil {
		t.Fatal("NewRateLimitedGoRoutinePool returned nil for qps=0")
	}

	p2 := NewRateLimitedGoRoutinePool(5, stop, ctx, -1, 0)
	if p2 == nil {
		t.Fatal("NewRateLimitedGoRoutinePool returned nil for negative qps and bucketNum")
	}
}

// ==================== Submit 和 Run 测试 ====================

func TestSubmitAndRun_Basic(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})
	defer close(stop)

	const taskNum = 10
	p := NewRateLimitedGoRoutinePool(5, stop, ctx, 100, 1)

	for i := 0; i < taskNum; i++ {
		task := GoRoutineExecutor{
			TaskID:          fmt.Sprintf("task-%d", i),
			GoRoutineFunc:   getConstValue,
			GoRoutineParams: []interface{}{"hello", 123},
		}
		p.Submit(task)
	}

	results := p.Run()
	if len(results) != taskNum {
		t.Fatalf("expected %d results, got %d", taskNum, len(results))
	}

	cnt := 0
	for _, v := range results {
		fr := v.(FuncResult)
		if fr.Err != nil {
			t.Errorf("unexpected error: %v", fr.Err)
		}
		cnt += fr.Result.(int)
	}
	if cnt != taskNum {
		t.Errorf("expected sum=%d, got %d", taskNum, cnt)
	}
}

func TestSubmitAndRun_SingleTask(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})
	defer close(stop)

	p := NewRateLimitedGoRoutinePool(1, stop, ctx, 100, 1)

	task := GoRoutineExecutor{
		TaskID:          "single-task",
		GoRoutineFunc:   getSum,
		GoRoutineParams: []interface{}{10, 20, 30},
	}
	p.Submit(task)

	results := p.Run()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	fr := results[0].(FuncResult)
	if fr.Err != nil {
		t.Errorf("unexpected error: %v", fr.Err)
	}
	if fr.Result.(int) != 60 {
		t.Errorf("expected result=60, got %v", fr.Result)
	}
}

func TestSubmitAndRun_NoTask(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})
	defer close(stop)

	p := NewRateLimitedGoRoutinePool(5, stop, ctx, 100, 1)

	// 不提交任何任务直接 Run
	results := p.Run()
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// ==================== 错误处理测试 ====================

func TestSubmitAndRun_WithErrors(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})
	defer close(stop)

	const taskNum = 5
	p := NewRateLimitedGoRoutinePool(3, stop, ctx, 100, 1)

	for i := 0; i < taskNum; i++ {
		task := GoRoutineExecutor{
			TaskID:          fmt.Sprintf("err-task-%d", i),
			GoRoutineFunc:   alwaysError,
			GoRoutineParams: []interface{}{},
		}
		p.Submit(task)
	}

	results := p.Run()
	if len(results) != taskNum {
		t.Fatalf("expected %d results, got %d", taskNum, len(results))
	}

	for _, v := range results {
		fr := v.(FuncResult)
		if fr.Err == nil {
			t.Error("expected error, got nil")
		}
		if fr.Result != nil {
			t.Errorf("expected nil result, got %v", fr.Result)
		}
	}
}

func TestSubmitAndRun_MixedSuccessAndError(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})
	defer close(stop)

	p := NewRateLimitedGoRoutinePool(5, stop, ctx, 100, 1)

	// 提交成功任务
	for i := 0; i < 3; i++ {
		task := GoRoutineExecutor{
			TaskID:          fmt.Sprintf("ok-task-%d", i),
			GoRoutineFunc:   getSum,
			GoRoutineParams: []interface{}{1, 2},
		}
		p.Submit(task)
	}

	// 提交失败任务
	for i := 0; i < 2; i++ {
		task := GoRoutineExecutor{
			TaskID:          fmt.Sprintf("fail-task-%d", i),
			GoRoutineFunc:   alwaysError,
			GoRoutineParams: []interface{}{},
		}
		p.Submit(task)
	}

	results := p.Run()
	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}

	successCnt := 0
	errorCnt := 0
	for _, v := range results {
		fr := v.(FuncResult)
		if fr.Err != nil {
			errorCnt++
		} else {
			successCnt++
		}
	}
	if successCnt != 3 {
		t.Errorf("expected 3 successes, got %d", successCnt)
	}
	if errorCnt != 2 {
		t.Errorf("expected 2 errors, got %d", errorCnt)
	}
}

// ==================== 并发安全测试 ====================

func TestSubmitAndRun_ConcurrentResults(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})
	defer close(stop)

	const taskNum = 20
	var counter int64

	counterFunc := func(param ...interface{}) (interface{}, error) {
		atomic.AddInt64(&counter, 1)
		return atomic.LoadInt64(&counter), nil
	}

	p := NewRateLimitedGoRoutinePool(10, stop, ctx, 100, 1)

	for i := 0; i < taskNum; i++ {
		task := GoRoutineExecutor{
			TaskID:          fmt.Sprintf("concurrent-%d", i),
			GoRoutineFunc:   counterFunc,
			GoRoutineParams: []interface{}{},
		}
		p.Submit(task)
	}

	results := p.Run()
	if len(results) != taskNum {
		t.Fatalf("expected %d results, got %d", taskNum, len(results))
	}

	finalCount := atomic.LoadInt64(&counter)
	if finalCount != taskNum {
		t.Errorf("expected counter=%d, got %d", taskNum, finalCount)
	}
}

// ==================== 速率限制测试 ====================

func TestSubmitAndRun_RateLimit(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})
	defer close(stop)

	const taskNum = 10
	const qps = 50

	noopFunc := func(param ...interface{}) (interface{}, error) {
		return "ok", nil
	}

	p := NewRateLimitedGoRoutinePool(5, stop, ctx, qps, 1)

	for i := 0; i < taskNum; i++ {
		task := GoRoutineExecutor{
			TaskID:          fmt.Sprintf("rate-task-%d", i),
			GoRoutineFunc:   noopFunc,
			GoRoutineParams: []interface{}{},
		}
		p.Submit(task)
	}

	start := time.Now()
	results := p.Run()
	elapsed := time.Since(start)

	if len(results) != taskNum {
		t.Fatalf("expected %d results, got %d", taskNum, len(results))
	}

	for _, v := range results {
		fr := v.(FuncResult)
		if fr.Err != nil {
			t.Errorf("unexpected error: %v", fr.Err)
		}
	}

	t.Logf("completed %d tasks in %v with qps=%d", taskNum, elapsed, qps)
}

// ==================== FuncResult 测试 ====================

func TestFuncResult_ZeroValue(t *testing.T) {
	fr := FuncResult{}
	if fr.Result != nil {
		t.Errorf("expected nil Result, got %v", fr.Result)
	}
	if fr.Err != nil {
		t.Errorf("expected nil Err, got %v", fr.Err)
	}
}

func TestFuncResult_WithValues(t *testing.T) {
	fr := FuncResult{
		Result: "success",
		Err:    nil,
	}
	if fr.Result != "success" {
		t.Errorf("expected Result='success', got %v", fr.Result)
	}

	fr2 := FuncResult{
		Result: nil,
		Err:    errors.New("failed"),
	}
	if fr2.Err == nil || fr2.Err.Error() != "failed" {
		t.Errorf("expected Err='failed', got %v", fr2.Err)
	}
}

// ==================== CustomControllerRateLimiter 测试 ====================

func TestCustomControllerRateLimiter(t *testing.T) {
	limiter := CustomControllerRateLimiter(10, 1)
	if limiter == nil {
		t.Fatal("CustomControllerRateLimiter returned nil")
	}
}

func TestCustomControllerRateLimiter_DifferentParams(t *testing.T) {
	testCases := []struct {
		name      string
		qps       int
		bucketNum int
	}{
		{"low_qps", 1, 1},
		{"high_qps", 1000, 10},
		{"single_bucket", 50, 1},
		{"multi_bucket", 50, 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			limiter := CustomControllerRateLimiter(tc.qps, tc.bucketNum)
			if limiter == nil {
				t.Errorf("CustomControllerRateLimiter(%d, %d) returned nil",
					tc.qps, tc.bucketNum)
			}
		})
	}
}

// ==================== GoRoutineExecutor 测试 ====================

func TestGoRoutineExecutor_ZeroValue(t *testing.T) {
	executor := GoRoutineExecutor{}
	if executor.TaskID != "" {
		t.Errorf("expected empty TaskID, got %q", executor.TaskID)
	}
	if executor.GoRoutineFunc != nil {
		t.Error("expected nil GoRoutineFunc")
	}
	if executor.GoRoutineParams != nil {
		t.Error("expected nil GoRoutineParams")
	}
}

// ==================== 大批量任务测试 ====================

func TestSubmitAndRun_LargeBatch(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})
	defer close(stop)

	const taskNum = 50

	noopFunc := func(param ...interface{}) (interface{}, error) {
		return param[0], nil
	}

	p := NewRateLimitedGoRoutinePool(10, stop, ctx, 200, 1)

	for i := 0; i < taskNum; i++ {
		task := GoRoutineExecutor{
			TaskID:          fmt.Sprintf("batch-%d", i),
			GoRoutineFunc:   noopFunc,
			GoRoutineParams: []interface{}{i},
		}
		p.Submit(task)
	}

	results := p.Run()
	if len(results) != taskNum {
		t.Fatalf("expected %d results, got %d", taskNum, len(results))
	}

	// 验证所有结果都存在（顺序可能不同）
	resultSet := make(map[int]bool)
	for _, v := range results {
		fr := v.(FuncResult)
		if fr.Err != nil {
			t.Errorf("unexpected error: %v", fr.Err)
			continue
		}
		resultSet[fr.Result.(int)] = true
	}
	if len(resultSet) != taskNum {
		t.Errorf("expected %d unique results, got %d", taskNum, len(resultSet))
	}
}

// ==================== 单 worker 测试 ====================

func TestSubmitAndRun_SingleWorker(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})
	defer close(stop)

	const taskNum = 5
	p := NewRateLimitedGoRoutinePool(1, stop, ctx, 100, 1)

	for i := 0; i < taskNum; i++ {
		task := GoRoutineExecutor{
			TaskID:          fmt.Sprintf("single-worker-%d", i),
			GoRoutineFunc:   getSum,
			GoRoutineParams: []interface{}{i, 1},
		}
		p.Submit(task)
	}

	results := p.Run()
	if len(results) != taskNum {
		t.Fatalf("expected %d results, got %d", taskNum, len(results))
	}

	totalSum := 0
	for _, v := range results {
		fr := v.(FuncResult)
		if fr.Err != nil {
			t.Errorf("unexpected error: %v", fr.Err)
			continue
		}
		totalSum += fr.Result.(int)
	}
	// sum(i+1 for i in 0..4) = 1+2+3+4+5 = 15
	expectedSum := 15
	if totalSum != expectedSum {
		t.Errorf("expected total sum=%d, got %d", expectedSum, totalSum)
	}
}