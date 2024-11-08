package test

import (
	"sync"
	"testing"
)

func TestPoint(t *testing.T) {
	var wg sync.WaitGroup
	var val int = 100
	wg.Add(2)
	go loop3(&val, &wg)
	go loop4(&val, &wg)
	wg.Wait()
	t.Log("val:", val)
}

func loop3(val *int, wg *sync.WaitGroup) {
	*val++
	wg.Done()
}

func loop4(val *int, wg *sync.WaitGroup) {
	*val++
	wg.Done()
}
