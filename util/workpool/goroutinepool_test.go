package workpool

import (
	"context"
	"fmt"
	"github.com/Chairou/toolbox/util/conv"
	"testing"
	"time"
)

var (
	totalTaskNum int = 100
)

func TestGoroutineRateLimit(t *testing.T) {
	cnt := 0
	result := work()
	for _, v := range result {
		cnt += v.(FuncResult).Result.(int)
	}
	if cnt == totalTaskNum {
		t.Log(cnt)
	} else {
		t.Error(cnt)
	}

}

func getConstValue(param ...interface{}) (interface{}, error) {
	fmt.Println(param[0].(string), param[1].(int))
	return 1, nil
}

func work() []interface{} {
	ctx := context.Background()
	stop := make(chan struct{})
	poolSize := 10
	qps := 15
	bucketNum := 1

	p := NewRateLimitedGoRoutinePool(poolSize, stop, ctx, qps, bucketNum)

	for i := 0; i < totalTaskNum; i++ {
		task := GoRoutineExecutor{
			TaskID:          "taskID-" + conv.String(i),
			GoRoutineFunc:   getConstValue,
			GoRoutineParams: []interface{}{"123", 456},
		}
		p.Submit(task)
	}
	now := time.Now()
	results := p.Run()
	fmt.Println(time.Since(now))
	return results
}
