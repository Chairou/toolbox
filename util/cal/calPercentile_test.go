package cal

import (
	"math/rand"
	"testing"
	"time"
)

func TestPercentilesList(t *testing.T) {
	list := []float64{}
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 1000000; i++ {
		list = append(list, rand.Float64())
	}
	now := time.Now()
	h := NewMinHeapList(list)
	t.Log(h.Percentile())
	time.Since(now)
}

func TestPercentiles(t *testing.T) {
	h := NewMaxHeap()
	h.Push(float64(1))
	h.Push(float64(2))
	h.Push(float64(3))
	h.Push(float64(4))
	h.Push(float64(5))
	t.Log(h.Percentile())

}
