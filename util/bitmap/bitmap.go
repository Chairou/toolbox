package bitmap

import (
	"fmt"
	"sync"
)

var bitMapStore sync.Map

type BitMapStruct struct {
	Name       string `json:"Name"`
	ByteBuffer []byte `json:"byteBuffer"`
	BitLength  uint64 `json:"length"`
	mu         sync.RWMutex // 添加读写锁保证线程安全
}

const byteSize = 8 //1 byte = 8 bits

// NewBitMap create the bitmap instance
// NewBitMap 生成bitmap对象
// parameter name name of bitmap
// parameter n the length of the bitmap
func NewBitMap(name string, n uint64) *BitMapStruct {
	newMap := &BitMapStruct{}
	newMap.Name = name
	newMap.BitLength = n
	newMap.ByteBuffer = make([]byte, n/byteSize+1)
	bitMapStore.Store(newMap.Name, newMap)
	return newMap
}

// GetBitMap get the bitmap instance
// GetBitMap 获取bitmap实例
// parameter name name of bitmap instance
func GetBitMap(name string) *BitMapStruct {
	inst, ok := bitMapStore.Load(name)
	if ok {
		return inst.(*BitMapStruct)
	} else {
		return nil
	}
}

// Set set the bit
// Set 设置对应位置的位数为1
// parameter m, the n bit of the bitmap
func (bt *BitMapStruct) Set(n uint64) error {
	if n >= bt.BitLength {
		return fmt.Errorf("bit position %d exceeds bitmap length %d", n, bt.BitLength)
	}
	
	bt.mu.Lock()
	defer bt.mu.Unlock()
	
	byteIndex := n / byteSize   //第x个字节（0,1,2...）
	offsetIndex := n % byteSize //偏移量(0<偏移量<byteSize)
	//第x个字节偏移量为offsetIndex的位 置位1
	bt.ByteBuffer[byteIndex] |= 1 << offsetIndex
	return nil
}

// MSet set multiple bits at once
// Mset 一次设置多个bit
// parameter n bit array
func (bt *BitMapStruct) MSet(n ...uint64) error {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	
	for _, v := range n {
		if v >= bt.BitLength {
			return fmt.Errorf("bit position %d exceeds bitmap length %d", v, bt.BitLength)
		}
		
		byteIndex := v / byteSize
		offsetIndex := v % byteSize
		bt.ByteBuffer[byteIndex] |= 1 << offsetIndex
	}
	return nil
}

// Del unset the bit
// Del 设置对应bit为0
// parameter n the n bit
func (bt *BitMapStruct) Del(n uint64) error {
	if n >= bt.BitLength {
		return fmt.Errorf("bit position %d exceeds bitmap length %d", n, bt.BitLength)
	}
	
	bt.mu.Lock()
	defer bt.mu.Unlock()
	
	byteIndex := n / byteSize
	offsetIndex := n % byteSize
	// 使用位掩码清零特定位，而不是清零整个字节
	bt.ByteBuffer[byteIndex] &^= 1 << offsetIndex
	return nil
}

// MDel Unset multiple bits at once
// MDel 批量删除对应bit
// parameter n bit array
func (bt *BitMapStruct) MDel(elements ...uint64) error {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	
	for _, v := range elements {
		if v >= bt.BitLength {
			return fmt.Errorf("bit position %d exceeds bitmap length %d", v, bt.BitLength)
		}
		
		byteIndex := v / byteSize
		offsetIndex := v % byteSize
		bt.ByteBuffer[byteIndex] &^= 1 << offsetIndex
	}
	return nil
}

// IsExist returns true if the specified
// IsExist 返回是否置位
// parameter n is position of bitmap
func (bt *BitMapStruct) IsExist(n uint64) bool {
	if n >= bt.BitLength {
		return false
	}
	
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	
	byteIndex := n / byteSize
	offsetIndex := n % byteSize
	return bt.ByteBuffer[byteIndex]&(1<<offsetIndex) != 0
}

// MExist returns true if the specified list
// MExist 批量判断是否置位
// parameter elements is position array of bitmap
func (bt *BitMapStruct) MExist(elements ...uint64) map[uint64]bool {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	
	len := len(elements)
	matchElementsList := make(map[uint64]bool, len)
	for _, v := range elements {
		if v < bt.BitLength {
			byteIndex := v / byteSize
			offsetIndex := v % byteSize
			matchElementsList[v] = bt.ByteBuffer[byteIndex]&(1<<offsetIndex) != 0
		} else {
			matchElementsList[v] = false
		}
	}
	return matchElementsList
}

// PrintAllBits Print all bits and all values
// PrintAllBits 打印整个bitmap
func (bt *BitMapStruct) PrintAllBits() {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	
	var i uint64
	for i = 0; i < bt.BitLength; i++ {
		byteIndex := i / byteSize
		offsetIndex := i % byteSize
		if bt.ByteBuffer[byteIndex]&(1<<offsetIndex) != 0 {
			fmt.Println("bit: ", i, "val: ", 1)
		} else {
			fmt.Println("bit: ", i, "val: ", 0)
		}
	}
}

// Clean set all bit zero
// Clean 清零所有位
func (bt *BitMapStruct) Clean() {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	
	// 清零现有内存而不是重新分配
	for i := range bt.ByteBuffer {
		bt.ByteBuffer[i] = 0
	}
}

// Destroy destory current bitmap instance
// Destroy 删除当前bitmap
func (bt *BitMapStruct) Destroy() {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	
	bt.BitLength = 0
	bt.ByteBuffer = nil
	bitMapStore.Delete(bt.Name)
	bt.Name = ""
}