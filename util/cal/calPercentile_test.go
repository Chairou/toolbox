package cal

import (
	"math/rand"
	"testing"
)

func TestPercentilesList(t *testing.T) {
	list := []float64{}
	for i := 0; i < 1000000; i++ {
		list = append(list, rand.Float64())
	}
	h := NewMinHeapList(list)
	t.Log(h.Percentile())
}

func TestPercentiles(t *testing.T) {
	h := NewMinHeap()
	h.Push(float64(1))
	h.Push(float64(2))
	h.Push(float64(3))
	h.Push(float64(4))
	h.Push(float64(5))
	t.Log(h.Percentile())

}
