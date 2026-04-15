// Package mapsort 提供对 map 按 key 或 value 排序的泛型工具函数
package mapsort

import (
	"cmp"
	"sort"
)

// Pair 表示一个键值对，用于 map 排序后的结果存储
type Pair struct {
	Key   string
	Value int
}

// PairList 是 Pair 的切片类型，实现了 sort.Interface 接口
type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// RankByWordCount 将 map[string]int 按值从大到小排序，返回排序后的 PairList
func RankByWordCount(wordFrequencies map[string]int) PairList {
	pl := make(PairList, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

// KVPair 表示一个泛型键值对，用于泛型 map 排序后的结果存储
type KVPair[K cmp.Ordered, V cmp.Ordered] struct {
	Key   K
	Value V
}

// SortByKey 将 map 按 key 排序，返回排序后的 KVPair 切片。
// ascending 为 true 时正序（从小到大），为 false 时倒序（从大到小）。
func SortByKey[K cmp.Ordered, V cmp.Ordered](m map[K]V, ascending bool) []KVPair[K, V] {
	pairs := make([]KVPair[K, V], 0, len(m))
	for k, v := range m {
		pairs = append(pairs, KVPair[K, V]{Key: k, Value: v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if ascending {
			return cmp.Less(pairs[i].Key, pairs[j].Key)
		}
		return cmp.Less(pairs[j].Key, pairs[i].Key)
	})
	return pairs
}

// SortByValue 将 map 按 value 排序，返回排序后的 KVPair 切片。
// ascending 为 true 时正序（从小到大），为 false 时倒序（从大到小）。
// 当 value 相同时，按 key 正序排列以保证结果稳定。
func SortByValue[K cmp.Ordered, V cmp.Ordered](m map[K]V, ascending bool) []KVPair[K, V] {
	pairs := make([]KVPair[K, V], 0, len(m))
	for k, v := range m {
		pairs = append(pairs, KVPair[K, V]{Key: k, Value: v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Value == pairs[j].Value {
			return cmp.Less(pairs[i].Key, pairs[j].Key)
		}
		if ascending {
			return cmp.Less(pairs[i].Value, pairs[j].Value)
		}
		return cmp.Less(pairs[j].Value, pairs[i].Value)
	})
	return pairs
}