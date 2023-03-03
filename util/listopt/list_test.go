package listopt

import (
	"math"
	"reflect"
	"testing"
)

//  go test -v list_test.go list_opt.go

func TestSplitList(t *testing.T) {
	abcList := make([]string, 0)
	for i := 0; i < 100; i++ {
		abcList = append(abcList, "abc")
	}
	qqq := SplitList(abcList, 3)
	t.Log(len(qqq[0]), len(qqq[1]), len(qqq[2]))
	if math.Abs(float64(len(qqq[0])-len(qqq[1]))) > 3 && math.Abs(float64(len(qqq[1])-len(qqq[2]))) > 3 {
		t.Error("SplitList is not avg")
	}
}

func TestRemoveRepeatedString(t *testing.T) {
	abcList := make([]string, 0)
	for i := 0; i < 100; i++ {
		abcList = append(abcList, "abc")
	}
	abcList = append(abcList, "gogogo")
	abcList = append(abcList, "we good")
	resultList := RemoveDuplicateString(abcList)
	if len(resultList) != 3 {
		t.Error("RemoveRepeatedString err")
	}
}

func TestRemoveDuplicateInt(t *testing.T) {
	abcList := make([]int, 0)
	for i := 0; i < 100; i++ {
		abcList = append(abcList, 111)
	}
	abcList = append(abcList, 222)
	abcList = append(abcList, 333)
	resultList := RemoveDuplicateInt(abcList)
	if len(resultList) != 3 {
		t.Error("RemoveRepeatedInt err")
	}
}

func TestDeleteString(t *testing.T) {
	abcList := []string{"aaa", "bbb", "ccc", "ddd"}
	resultList := DeleteString(abcList, "aaa")
	t.Log(resultList)
	if In(resultList, "aaa") {
		t.Error("DeleteString err")
	}
}

func TestRemoveRepeatedElement(t *testing.T) {
	abcList := make([]string, 0)
	for i := 0; i < 100; i++ {
		abcList = append(abcList, "abc")
	}
	abcList = append(abcList, "gogogo")
	abcList = append(abcList, "we good")
	resultList := RemoveRepeatedElement(abcList)
	if len(resultList) != 3 {
		t.Error("RemoveRepeatedElement err")
	}
	for _, v := range resultList {
		t.Log(v, reflect.TypeOf(v))
	}
}

func TestUnion(t *testing.T) {
	firstList := []string{"ab", "bc"}
	secondList := []string{"cd", "de"}
	resultList := UnionStr(firstList, secondList)
	t.Log(resultList)
	if len(resultList) != 4 {
		t.Error("UnionStr err")
	}
}

func TestIntersect(t *testing.T) {
	firstList := []string{"ab", "bc", "fff"}
	secondList := []string{"cd", "de", "fff"}
	resultList := IntersectStr(firstList, secondList)
	t.Log(resultList)
	if resultList[0] != "fff" {
		t.Error("IntersectStr err")
	}
}

func TestDifferenceStr1(t *testing.T) {
	firstList := []string{"ab", "bc", "fff"}
	secondList := []string{"cd", "de", "fff"}
	resultList := DifferenceStr1(firstList, secondList)
	t.Log(resultList)
	if !(In(resultList, "ab") && In(resultList, "bc")) {
		t.Error("Difference err")
	}
}

func TestDifferenceStr2(t *testing.T) {
	firstList := []string{"ab", "bc", "fff"}
	secondList := []string{"cd", "de", "fff"}
	resultList := DifferenceStr2(firstList, secondList)
	t.Log(resultList)
	if !(In(resultList, "ab") && In(resultList, "bc")) {
		t.Error("Difference err")
	}
}

func TestReverse(t *testing.T) {
	firstList := []string{"ab", "bc", "fff"}
	resultList := ReverseStr(firstList)
	t.Log(resultList)
	if resultList[0] != "fff" || resultList[1] != "bc" || resultList[2] != "ab" {
		t.Error("Reverse err")
	}
}
