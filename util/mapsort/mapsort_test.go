package mapsort

import (
	"sort"
	"testing"
)

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func rankByWordCount(wordFrequencies map[string]int) PairList {
	pl := make(PairList, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	s := []int{5, 2, 6, 3, 1, 4} // 未排序的切片数据
	sort.Sort(sort.Reverse(sort.IntSlice(s)))

	return pl
}

func TestRankByWordCount(t *testing.T) {
	tmpMap := make(map[string]int, 0)
	tmpMap["Gopher"] = 7
	tmpMap["Alice"] = 55
	tmpMap["Vera"] = 24
	tmpMap["Bob"] = 75
	sortedList := rankByWordCount(tmpMap)
	t.Log(sortedList)
}

// TestSortSample 排序的例子, 见倒数第二行
func TestSortSample(t *testing.T) {
	var people = []struct {
		Name string
		Age  int
	}{
		{"Gopher", 7},
		{"Alice", 55},
		{"Vera", 24},
		{"Bob", 75},
	}

	sort.SliceStable(people, func(i, j int) bool { return people[i].Age > people[j].Age }) // 按年龄降序排序
	t.Log("Sort by age:", people)
	if people[0].Age < people[1].Age {
		t.Error("sort err:", people)
	}
}
