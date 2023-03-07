package bitmap

import (
	"fmt"
	"sync"
)

var bitMapStore sync.Map

type BitMapStruct struct {
	Name       string `json:"Name"`
	ByteBuffer []byte `json:"byteBuffer"`
	Length     uint64 `json:"length"`
}

const byteSize = 8 //定义的bitmap为byte的数组，byte为8bit

func GetBitMap(name string, n uint64) *BitMapStruct {
	inst, ok := bitMapStore.Load(name)
	if ok {
		if inst.(*BitMapStruct).Length == n {
			return inst.(*BitMapStruct)
		}
	}
	newMap := &BitMapStruct{}
	newMap.Name = name
	newMap.Length = n
	newMap.ByteBuffer = make([]byte, n/byteSize+1)
	bitMapStore.Store(newMap.Name, newMap)
	return newMap
}

func (bt *BitMapStruct) Set(n uint64) {
	if n/byteSize > uint64(len(bt.ByteBuffer)) {
		fmt.Println("大小超出bitmap范围")
		return
	}
	byteIndex := n / byteSize   //第x个字节（0,1,2...）
	offsetIndex := n % byteSize //偏移量(0<偏移量<byteSize)
	//bt[byteIndex] = bt[byteIndex] | 1<<offsetIndex //异或1（置位）
	//第x个字节偏移量为offsetIndex的位 置位1
	bt.ByteBuffer[byteIndex] |= 1 << offsetIndex //异或1（置位）
}

func (bt *BitMapStruct) MSet(n ...uint64) {
	for _, v := range n {
		bt.Set(v)
	}
}

func (bt *BitMapStruct) Del(n uint64) {
	if n/byteSize > uint64(len(bt.ByteBuffer)) {
		fmt.Println("大小超出bitmap范围")
		return
	}
	byteIndex := n / byteSize
	offsetIndex := n % byteSize
	bt.ByteBuffer[byteIndex] &= 0 << offsetIndex //清零
}

func (bt *BitMapStruct) MDel(elements ...uint64) {
	for _, v := range elements {
		bt.Del(v)
	}
}

func (bt *BitMapStruct) IsExist(n uint64) bool {
	if n/byteSize > uint64(len(bt.ByteBuffer)) {
		fmt.Println("大小超出bitmap范围")
		return false
	}
	byteIndex := n / byteSize
	offsetIndex := n % byteSize
	//fmt.Println(bt[byteIndex] & (1 << offsetIndex))
	return bt.ByteBuffer[byteIndex]&(1<<offsetIndex) != 0 //TODO：注意：条件是 ！=0，有可能是：16,32等
}

func (bt *BitMapStruct) MExist(elements ...uint64) map[uint64]bool {
	len := len(elements)
	matchElementsList := make(map[uint64]bool, len)
	for _, v := range elements {
		if bt.IsExist(v) {
			matchElementsList[v] = true
		} else {
			matchElementsList[v] = false
		}
	}
	return matchElementsList
}

func (bt *BitMapStruct) PrintAllBits() {
	var i uint64
	for i = 0; i < uint64(len(bt.ByteBuffer)*8); i++ {
		if bt.IsExist(i) {
			fmt.Println("bit: ", i, "valL: ", 1)
		} else {
			fmt.Println("bit: ", i, "valL: ", 0)
		}
	}
}

func (bt *BitMapStruct) Clean() {
	bt.ByteBuffer = make([]byte, bt.Length/byteSize+1)
	//bt.ByteBuffer = bt.ByteBuffer[:0]
}

func (bt *BitMapStruct) Destroy() {
	bt.Length = 0
	bt.ByteBuffer = nil
	bitMapStore.Delete(bt.Name)
	bt.Name = ""
}
