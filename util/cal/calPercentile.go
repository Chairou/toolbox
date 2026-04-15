package cal

import (
	"container/heap"
)

// MinHeap 定义一个最小堆类型
type MinHeap []float64

// Len 实现 heap.Interface 接口的 Len 方法
func (h MinHeap) Len() int {
	return len(h)
}

// Less 实现 heap.Interface 接口的 Less 方法
func (h MinHeap) Less(i, j int) bool {
	return h[i] < h[j]
}

// Swap 实现 heap.Interface 接口的 Swap 方法
func (h MinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// Push 实现 heap.Interface 接口的 Push 方法
func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(float64))
}

// Pop 实现 heap.Interface 接口的 Pop 方法
func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// Percentile 输出中位数
func (h *MinHeap) Percentile() float64 {
	popNum := 0
	var percentile float64
	if h.Len()&1 == 0 {
		popNum = h.Len() / 2
		for i := 0; i < popNum; i++ {
			percentile = h.Pop().(float64)
		}
		percentile += h.Pop().(float64)
		avgPercentile := percentile / 2
		return avgPercentile
	} else {
		popNum = (h.Len() / 2) + 1
	}
	for i := 0; i < popNum; i++ {
		percentile = h.Pop().(float64)
	}
	return percentile
}

// NewMinHeapList 根据float64切片创建并初始化一个最小堆
func NewMinHeapList(a []float64) *MinHeap {
	h := MinHeap(a)
	heap.Init(&h)
	return &h
}

// NewMinHeap 创建并初始化一个空的最小堆
func NewMinHeap() *MinHeap {
	h := MinHeap{}
	heap.Init(&h)
	return &h
}
