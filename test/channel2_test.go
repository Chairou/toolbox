package test

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestClient2(t *testing.T) {
	for i := 0; i < 999999; i++ {
		i++
	}
	var wg sync.WaitGroup
	bb := make(chan *time.Time)
	wg.Add(1)
	go loop2(bb, &wg)
	now := time.Now()
	bb <- &now
	close(bb)

	wg.Wait()
}

func loop2(aa chan *time.Time, wg *sync.WaitGroup) {
	for {
		before, ok := <-aa
		if !ok {
			return
		} else {
			log.Println(time.Since(*before))
			wg.Done()
		}
	}

}
