package test

import (
	"log"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	b := make(chan time.Time) // 无缓存通道
	done := make(chan struct{}, 1)
	go loop(b, done) // 消费b中的内容
	producer(b)
	<-done
}

func producer(b chan time.Time) {
	// 发送者
	now := time.Now()
	for {
		ttl := time.Since(now)
		if ttl.Seconds() >= 1 {
			break
		}

		b <- time.Now()
	}

	close(b)
}

func loop(b chan time.Time, done chan struct{}) {
	defer close(done)

	for {
		before, ok := <-b
		if !ok {
			break
		}

		ttl := time.Since(before)
		// log.Println("before: ", before.Format("2006-01-02 15:04:05"))
		log.Println("ttl: ", ttl)
	}
}
