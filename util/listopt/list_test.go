package listopt

import (
	"testing"
)

func TestSplitList(t *testing.T) {
	abcList := make([]string, 0)
	for i := 0; i < 100; i++ {
		abcList = append(abcList, "abc")
	}
	result := SplitList(abcList, 3)
	if len(result) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(result))
	}
	// 验证所有元素总数等于原始数组长度
	total := 0
	for _, seg := range result {
		total += len(seg)
	}
	if total != 100 {
		t.Errorf("expected total 100 elements, got %d", total)
	}
	t.Log(len(result[0]), len(result[1]), len(result[2]))
	// 验证每段长度差不超过1
	for i := 0; i < len(result)-1; i++ {
		diff := len(result[i]) - len(result[i+1])
		if diff < -1 || diff > 1 {
			t.Errorf("segment %d len=%d, segment %d len=%d, diff > 1", i, len(result[i]), i+1, len(result[i+1]))
		}
	}
}

func TestSplitList_NumGreaterThanLen(t *testing.T) {
	arr := []string{"a", "b"}
	result := SplitList(arr, 5)
	if result != nil {
		t.Errorf("expected nil when num > len, got %v", result)
	}
}

func TestSplitList_NumZero(t *testing.T) {
	arr := []string{"a", "b"}
	result := SplitList(arr, 0)
	if result != nil {
		t.Errorf("expected nil when num=0, got %v", result)
	}
}

func TestSplitList_NumNegative(t *testing.T) {
	arr := []string{"a", "b"}
	result := SplitList(arr, -1)
	if result != nil {
		t.Errorf("expected nil when num<0, got %v", result)
	}
}

func TestSplitList_EqualSplit(t *testing.T) {
	arr := []string{"a", "b", "c", "d"}
	result := SplitList(arr, 2)
	if len(result) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(result))
	}
	if len(result[0]) != 2 || len(result[1]) != 2 {
		t.Errorf("expected [2,2], got [%d,%d]", len(result[0]), len(result[1]))
	}
}

func TestSplitList_SingleSegment(t *testing.T) {
	arr := []string{"a", "b", "c"}
	result := SplitList(arr, 1)
	if len(result) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(result))
	}
	if len(result[0]) != 3 {
		t.Errorf("expected segment len 3, got %d", len(result[0]))
	}
}

func TestRemoveRepeatedElement_String(t *testing.T) {
	abcList := []string{"abc", "abc", "gogogo", "we good", "abc"}
	resultList := RemoveRepeatedElement(abcList)
	if len(resultList) != 3 {
		t.Errorf("expected 3 unique elements, got %d", len(resultList))
	}
}

func TestRemoveRepeatedElement_Int(t *testing.T) {
	intList := []int{1, 2, 2, 3, 3, 3}
	resultList := RemoveRepeatedElement(intList)
	if len(resultList) != 3 {
		t.Errorf("expected 3 unique elements, got %d", len(resultList))
	}
}

func TestRemoveRepeatedElement_Float64(t *testing.T) {
	floatList := []float64{1.1, 2.2, 1.1, 3.3}
	resultList := RemoveRepeatedElement(floatList)
	if len(resultList) != 3 {
		t.Errorf("expected 3 unique elements, got %d", len(resultList))
	}
}

func TestRemoveRepeatedElement_UnsupportedType(t *testing.T) {
	boolList := []bool{true, false, true}
	resultList := RemoveRepeatedElement(boolList)
	if len(resultList) != 0 {
		t.Errorf("expected 0 for unsupported type, got %d", len(resultList))
	}
}

func TestRemoveDuplicateString(t *testing.T) {
	input := []string{"abc", "abc", "gogogo", "we good", "abc"}
	result := RemoveDuplicateString(input)
	if len(result) != 3 {
		t.Errorf("expected 3, got %d", len(result))
	}
	// 验证顺序保持
	if result[0] != "abc" || result[1] != "gogogo" || result[2] != "we good" {
		t.Errorf("unexpected order: %v", result)
	}
}

func TestRemoveDuplicateString_Empty(t *testing.T) {
	result := RemoveDuplicateString([]string{})
	if len(result) != 0 {
		t.Errorf("expected 0, got %d", len(result))
	}
}

func TestRemoveDuplicateInt(t *testing.T) {
	input := []int{111, 111, 222, 333, 111}
	result := RemoveDuplicateInt(input)
	if len(result) != 3 {
		t.Errorf("expected 3, got %d", len(result))
	}
	if result[0] != 111 || result[1] != 222 || result[2] != 333 {
		t.Errorf("unexpected order: %v", result)
	}
}

func TestRemoveDuplicateInt_Empty(t *testing.T) {
	result := RemoveDuplicateInt([]int{})
	if len(result) != 0 {
		t.Errorf("expected 0, got %d", len(result))
	}
}

func TestDeleteString(t *testing.T) {
	abcList := []string{"aaa", "bbb", "ccc", "ddd"}
	resultList := DeleteString(abcList, "aaa")
	if len(resultList) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(resultList))
	}
	if In(resultList, "aaa") {
		t.Error("aaa should have been deleted")
	}
}

func TestDeleteString_NotFound(t *testing.T) {
	abcList := []string{"aaa", "bbb", "ccc"}
	resultList := DeleteString(abcList, "zzz")
	if len(resultList) != 3 {
		t.Errorf("expected 3 elements when deleting non-existent, got %d", len(resultList))
	}
}

func TestDeleteString_Empty(t *testing.T) {
	resultList := DeleteString([]string{}, "aaa")
	if len(resultList) != 0 {
		t.Errorf("expected 0 elements for empty list, got %d", len(resultList))
	}
}

func TestDeleteString_LastElement(t *testing.T) {
	abcList := []string{"aaa", "bbb", "ccc"}
	resultList := DeleteString(abcList, "ccc")
	if len(resultList) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(resultList))
	}
	if In(resultList, "ccc") {
		t.Error("ccc should have been deleted")
	}
}

func TestIntersectStr(t *testing.T) {
	firstList := []string{"ab", "bc", "fff"}
	secondList := []string{"cd", "de", "fff"}
	resultList := IntersectStr(firstList, secondList)
	if len(resultList) != 1 || resultList[0] != "fff" {
		t.Errorf("expected [fff], got %v", resultList)
	}
}

func TestIntersectStr_NoCommon(t *testing.T) {
	firstList := []string{"ab", "bc"}
	secondList := []string{"cd", "de"}
	resultList := IntersectStr(firstList, secondList)
	if len(resultList) != 0 {
		t.Errorf("expected empty intersection, got %v", resultList)
	}
}

func TestIntersectStr_WithDuplicates(t *testing.T) {
	// 修复前的bug：slice1中有重复元素时IntersectStr结果不正确
	firstList := []string{"ab", "ab", "fff"}
	secondList := []string{"ab", "fff"}
	resultList := IntersectStr(firstList, secondList)
	if len(resultList) != 2 {
		t.Errorf("expected 2 elements in intersection, got %d: %v", len(resultList), resultList)
	}
	if !In(resultList, "ab") || !In(resultList, "fff") {
		t.Errorf("expected [ab, fff], got %v", resultList)
	}
}

func TestIntersectStr_Empty(t *testing.T) {
	resultList := IntersectStr([]string{}, []string{"a", "b"})
	if len(resultList) != 0 {
		t.Errorf("expected empty, got %v", resultList)
	}
}

func TestUnionStr(t *testing.T) {
	firstList := []string{"ab", "bc"}
	secondList := []string{"cd", "de"}
	resultList := UnionStr(firstList, secondList)
	if len(resultList) != 4 {
		t.Errorf("expected 4, got %d: %v", len(resultList), resultList)
	}
}

func TestUnionStr_WithOverlap(t *testing.T) {
	firstList := []string{"ab", "bc", "fff"}
	secondList := []string{"bc", "fff", "de"}
	resultList := UnionStr(firstList, secondList)
	if len(resultList) != 4 {
		t.Errorf("expected 4, got %d: %v", len(resultList), resultList)
	}
}

func TestUnionStr_Empty(t *testing.T) {
	resultList := UnionStr([]string{}, []string{"a", "b"})
	if len(resultList) != 2 {
		t.Errorf("expected 2, got %d", len(resultList))
	}
}

func TestDifferenceStr1(t *testing.T) {
	firstList := []string{"ab", "bc", "fff"}
	secondList := []string{"cd", "de", "fff"}
	resultList := DifferenceStr1(firstList, secondList)
	if len(resultList) != 2 {
		t.Fatalf("expected 2, got %d: %v", len(resultList), resultList)
	}
	if !(In(resultList, "ab") && In(resultList, "bc")) {
		t.Errorf("expected [ab, bc], got %v", resultList)
	}
}

func TestDifferenceStr1_NoOverlap(t *testing.T) {
	firstList := []string{"ab", "bc"}
	secondList := []string{"cd", "de"}
	resultList := DifferenceStr1(firstList, secondList)
	if len(resultList) != 2 {
		t.Errorf("expected 2, got %d: %v", len(resultList), resultList)
	}
}

func TestDifferenceStr2(t *testing.T) {
	firstList := []string{"ab", "bc", "fff"}
	secondList := []string{"cd", "de", "fff"}
	resultList := DifferenceStr2(firstList, secondList)
	if len(resultList) != 2 {
		t.Fatalf("expected 2, got %d: %v", len(resultList), resultList)
	}
	if !(In(resultList, "ab") && In(resultList, "bc")) {
		t.Errorf("expected [ab, bc], got %v", resultList)
	}
}

func TestDifferenceStr2_NoOverlap(t *testing.T) {
	firstList := []string{"ab", "bc"}
	secondList := []string{"cd", "de"}
	resultList := DifferenceStr2(firstList, secondList)
	if len(resultList) != 2 {
		t.Errorf("expected 2, got %d: %v", len(resultList), resultList)
	}
}

func TestIn(t *testing.T) {
	list := []string{"ab", "bc", "fff"}
	if !In(list, "ab") {
		t.Error("expected 'ab' to be found")
	}
	if In(list, "zzz") {
		t.Error("expected 'zzz' to not be found")
	}
}

func TestIn_DoesNotModifySlice(t *testing.T) {
	// 修复前的bug：In函数会sort.Strings修改原切片
	list := []string{"ccc", "aaa", "bbb"}
	original := make([]string, len(list))
	copy(original, list)
	In(list, "aaa")
	for i := range list {
		if list[i] != original[i] {
			t.Errorf("In() modified input slice: expected %v, got %v", original, list)
			break
		}
	}
}

func TestIn_Empty(t *testing.T) {
	if In([]string{}, "a") {
		t.Error("expected false for empty list")
	}
}

func TestReverseStr(t *testing.T) {
	firstList := []string{"ab", "bc", "fff"}
	resultList := ReverseStr(firstList)
	if resultList[0] != "fff" || resultList[1] != "bc" || resultList[2] != "ab" {
		t.Errorf("expected [fff, bc, ab], got %v", resultList)
	}
}

func TestReverseStr_DoesNotModifyOriginal(t *testing.T) {
	// 修复前的bug：ReverseStr会原地修改传入的切片
	original := []string{"ab", "bc", "fff"}
	backup := make([]string, len(original))
	copy(backup, original)
	ReverseStr(original)
	for i := range original {
		if original[i] != backup[i] {
			t.Errorf("ReverseStr() modified input slice: expected %v, got %v", backup, original)
			break
		}
	}
}

func TestReverseStr_Empty(t *testing.T) {
	result := ReverseStr([]string{})
	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}
}

func TestReverseStr_Single(t *testing.T) {
	result := ReverseStr([]string{"only"})
	if len(result) != 1 || result[0] != "only" {
		t.Errorf("expected [only], got %v", result)
	}
}

func TestIsContainsInStringArr(t *testing.T) {
	arr := []string{"hello", "world"}
	if !IsContainsInStringArr("hello world", arr) {
		t.Error("expected true for 'hello world' containing 'hello'")
	}
	if IsContainsInStringArr("foo bar", arr) {
		t.Error("expected false for 'foo bar'")
	}
}

func TestIsContainsInStringArr_EmptyStr(t *testing.T) {
	arr := []string{"hello"}
	if IsContainsInStringArr("", arr) {
		t.Error("expected false for empty string")
	}
}

func TestIsContainsInStringArr_EmptyArr(t *testing.T) {
	if IsContainsInStringArr("hello", []string{}) {
		t.Error("expected false for empty array")
	}
}

func TestIsInStringArr(t *testing.T) {
	arr := []string{"Hello", "World"}
	// 大小写不敏感
	if !IsInStringArr(arr, "hello") {
		t.Error("expected true for case-insensitive match")
	}
	if !IsInStringArr(arr, " Hello ") {
		t.Error("expected true for trimmed match")
	}
	if IsInStringArr(arr, "foo") {
		t.Error("expected false for non-existent")
	}
}

func TestIsInStringArr_Empty(t *testing.T) {
	if IsInStringArr([]string{}, "hello") {
		t.Error("expected false for empty array")
	}
}

func TestIsInIntPointerArr(t *testing.T) {
	a, b, c := 1, 2, 3
	arr := []*int{&a, &b, nil, &c}
	if !IsInIntPointerArr(arr, 2) {
		t.Error("expected true for 2")
	}
	if IsInIntPointerArr(arr, 99) {
		t.Error("expected false for 99")
	}
}

func TestIsInIntPointerArr_WithNil(t *testing.T) {
	arr := []*int{nil, nil}
	if IsInIntPointerArr(arr, 0) {
		t.Error("expected false when all elements are nil")
	}
}

func TestIsInArr(t *testing.T) {
	arr := []interface{}{"hello", 123, 1.5}
	if !IsInArr(arr, "hello") {
		t.Error("expected true for 'hello'")
	}
	if !IsInArr(arr, 123) {
		t.Error("expected true for 123")
	}
	if IsInArr(arr, "notfound") {
		t.Error("expected false for 'notfound'")
	}
}

func TestIsInArr_Empty(t *testing.T) {
	if IsInArr([]interface{}{}, "a") {
		t.Error("expected false for empty array")
	}
}

func TestInIntArr(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5}
	if !InIntArr(arr, 3) {
		t.Error("expected true for 3")
	}
	if InIntArr(arr, 99) {
		t.Error("expected false for 99")
	}
}

func TestInIntArr_Empty(t *testing.T) {
	if InIntArr([]int{}, 1) {
		t.Error("expected false for empty array")
	}
}

func TestKeyInIntMap(t *testing.T) {
	m := map[int]interface{}{1: "a", 2: "b", 3: "c"}
	if !KeyInIntMap(m, 1) {
		t.Error("expected true for key 1")
	}
	if KeyInIntMap(m, 99) {
		t.Error("expected false for key 99")
	}
}

func TestKeyInIntMap_Empty(t *testing.T) {
	m := map[int]interface{}{}
	if KeyInIntMap(m, 1) {
		t.Error("expected false for empty map")
	}
}

func TestGetValue(t *testing.T) {
	// 测试nil
	if GetValue(nil) != nil {
		t.Error("expected nil for nil input")
	}

	// 测试字符串（应转小写并trim）
	result := GetValue(" Hello ")
	if result != "hello" {
		t.Errorf("expected 'hello', got %v", result)
	}

	// 测试int
	result = GetValue(42)
	if result != 42 {
		t.Errorf("expected 42, got %v", result)
	}

	// 测试指针
	s := "World"
	result = GetValue(&s)
	if result != "world" {
		t.Errorf("expected 'world', got %v", result)
	}

	// 测试nil指针
	var p *string
	result = GetValue(p)
	if result != nil {
		t.Errorf("expected nil for nil pointer, got %v", result)
	}
}

func TestIsAllSameNum(t *testing.T) {
	if !IsAllSameNum([]int{5, 5, 5, 5}, 5) {
		t.Error("expected true for all same")
	}
	if IsAllSameNum([]int{5, 5, 3, 5}, 5) {
		t.Error("expected false for not all same")
	}
}

func TestIsAllSameNum_Empty(t *testing.T) {
	// 空数组，所有元素都等于val（空真）
	if !IsAllSameNum([]int{}, 5) {
		t.Error("expected true for empty array (vacuous truth)")
	}
}
