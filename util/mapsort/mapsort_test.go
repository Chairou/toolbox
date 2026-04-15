package mapsort

import (
	"testing"
)

// ==================== RankByWordCount 原有测试 ====================

func TestRankByWordCount(t *testing.T) {
	tmpMap := map[string]int{
		"Gopher": 7,
		"Alice":  55,
		"Vera":   24,
		"Bob":    75,
	}
	sortedList := RankByWordCount(tmpMap)
	t.Log(sortedList)

	if len(sortedList) != 4 {
		t.Errorf("expected length 4, got %d", len(sortedList))
	}

	for i := 0; i < len(sortedList)-1; i++ {
		if sortedList[i].Value < sortedList[i+1].Value {
			t.Errorf("not sorted in descending order at index %d: %d < %d",
				i, sortedList[i].Value, sortedList[i+1].Value)
		}
	}

	expectedOrder := []string{"Bob", "Alice", "Vera", "Gopher"}
	for i, expected := range expectedOrder {
		if sortedList[i].Key != expected {
			t.Errorf("expected key %q at index %d, got %q", expected, i, sortedList[i].Key)
		}
	}
}

func TestRankByWordCount_EmptyMap(t *testing.T) {
	tmpMap := map[string]int{}
	sortedList := RankByWordCount(tmpMap)
	if len(sortedList) != 0 {
		t.Errorf("expected empty list, got length %d", len(sortedList))
	}
}

func TestRankByWordCount_NilMap(t *testing.T) {
	var tmpMap map[string]int
	sortedList := RankByWordCount(tmpMap)
	if len(sortedList) != 0 {
		t.Errorf("expected empty list for nil map, got length %d", len(sortedList))
	}
}

func TestRankByWordCount_SingleElement(t *testing.T) {
	tmpMap := map[string]int{"only": 42}
	sortedList := RankByWordCount(tmpMap)
	if len(sortedList) != 1 {
		t.Fatalf("expected length 1, got %d", len(sortedList))
	}
	if sortedList[0].Key != "only" || sortedList[0].Value != 42 {
		t.Errorf("expected {only, 42}, got {%s, %d}", sortedList[0].Key, sortedList[0].Value)
	}
}

// ==================== SortByKey 测试 ====================

func TestSortByKey_StringInt_Ascending(t *testing.T) {
	m := map[string]int{
		"cherry": 3,
		"apple":  1,
		"banana": 2,
	}
	result := SortByKey(m, true)
	if len(result) != 3 {
		t.Fatalf("expected length 3, got %d", len(result))
	}
	expectedKeys := []string{"apple", "banana", "cherry"}
	for i, expected := range expectedKeys {
		if result[i].Key != expected {
			t.Errorf("index %d: expected key %q, got %q", i, expected, result[i].Key)
		}
	}
}

func TestSortByKey_StringInt_Descending(t *testing.T) {
	m := map[string]int{
		"cherry": 3,
		"apple":  1,
		"banana": 2,
	}
	result := SortByKey(m, false)
	if len(result) != 3 {
		t.Fatalf("expected length 3, got %d", len(result))
	}
	expectedKeys := []string{"cherry", "banana", "apple"}
	for i, expected := range expectedKeys {
		if result[i].Key != expected {
			t.Errorf("index %d: expected key %q, got %q", i, expected, result[i].Key)
		}
	}
}

func TestSortByKey_IntString_Ascending(t *testing.T) {
	m := map[int]string{
		3: "cherry",
		1: "apple",
		2: "banana",
	}
	result := SortByKey(m, true)
	if len(result) != 3 {
		t.Fatalf("expected length 3, got %d", len(result))
	}
	expectedKeys := []int{1, 2, 3}
	for i, expected := range expectedKeys {
		if result[i].Key != expected {
			t.Errorf("index %d: expected key %d, got %d", i, expected, result[i].Key)
		}
	}
}

func TestSortByKey_IntString_Descending(t *testing.T) {
	m := map[int]string{
		3: "cherry",
		1: "apple",
		2: "banana",
	}
	result := SortByKey(m, false)
	expectedKeys := []int{3, 2, 1}
	for i, expected := range expectedKeys {
		if result[i].Key != expected {
			t.Errorf("index %d: expected key %d, got %d", i, expected, result[i].Key)
		}
	}
}

func TestSortByKey_Float64Int_Ascending(t *testing.T) {
	m := map[float64]int{
		3.14: 1,
		1.41: 2,
		2.72: 3,
	}
	result := SortByKey(m, true)
	expectedKeys := []float64{1.41, 2.72, 3.14}
	for i, expected := range expectedKeys {
		if result[i].Key != expected {
			t.Errorf("index %d: expected key %f, got %f", i, expected, result[i].Key)
		}
	}
}

func TestSortByKey_EmptyMap(t *testing.T) {
	m := map[string]int{}
	result := SortByKey(m, true)
	if len(result) != 0 {
		t.Errorf("expected empty result, got length %d", len(result))
	}
}

func TestSortByKey_NilMap(t *testing.T) {
	var m map[string]int
	result := SortByKey(m, true)
	if len(result) != 0 {
		t.Errorf("expected empty result for nil map, got length %d", len(result))
	}
}

func TestSortByKey_SingleElement(t *testing.T) {
	m := map[string]int{"only": 42}
	result := SortByKey(m, true)
	if len(result) != 1 {
		t.Fatalf("expected length 1, got %d", len(result))
	}
	if result[0].Key != "only" || result[0].Value != 42 {
		t.Errorf("expected {only, 42}, got {%s, %d}", result[0].Key, result[0].Value)
	}
}

// ==================== SortByValue 测试 ====================

func TestSortByValue_StringInt_Ascending(t *testing.T) {
	m := map[string]int{
		"cherry": 3,
		"apple":  1,
		"banana": 2,
	}
	result := SortByValue(m, true)
	if len(result) != 3 {
		t.Fatalf("expected length 3, got %d", len(result))
	}
	expectedKeys := []string{"apple", "banana", "cherry"}
	expectedValues := []int{1, 2, 3}
	for i := range result {
		if result[i].Key != expectedKeys[i] {
			t.Errorf("index %d: expected key %q, got %q", i, expectedKeys[i], result[i].Key)
		}
		if result[i].Value != expectedValues[i] {
			t.Errorf("index %d: expected value %d, got %d", i, expectedValues[i], result[i].Value)
		}
	}
}

func TestSortByValue_StringInt_Descending(t *testing.T) {
	m := map[string]int{
		"cherry": 3,
		"apple":  1,
		"banana": 2,
	}
	result := SortByValue(m, false)
	if len(result) != 3 {
		t.Fatalf("expected length 3, got %d", len(result))
	}
	expectedKeys := []string{"cherry", "banana", "apple"}
	expectedValues := []int{3, 2, 1}
	for i := range result {
		if result[i].Key != expectedKeys[i] {
			t.Errorf("index %d: expected key %q, got %q", i, expectedKeys[i], result[i].Key)
		}
		if result[i].Value != expectedValues[i] {
			t.Errorf("index %d: expected value %d, got %d", i, expectedValues[i], result[i].Value)
		}
	}
}

func TestSortByValue_IntString_Ascending(t *testing.T) {
	m := map[int]string{
		1: "cherry",
		2: "apple",
		3: "banana",
	}
	result := SortByValue(m, true)
	expectedValues := []string{"apple", "banana", "cherry"}
	for i, expected := range expectedValues {
		if result[i].Value != expected {
			t.Errorf("index %d: expected value %q, got %q", i, expected, result[i].Value)
		}
	}
}

func TestSortByValue_IntString_Descending(t *testing.T) {
	m := map[int]string{
		1: "cherry",
		2: "apple",
		3: "banana",
	}
	result := SortByValue(m, false)
	expectedValues := []string{"cherry", "banana", "apple"}
	for i, expected := range expectedValues {
		if result[i].Value != expected {
			t.Errorf("index %d: expected value %q, got %q", i, expected, result[i].Value)
		}
	}
}

func TestSortByValue_SameValues_StableByKey(t *testing.T) {
	m := map[string]int{
		"c": 10,
		"a": 10,
		"b": 10,
	}
	result := SortByValue(m, true)
	if len(result) != 3 {
		t.Fatalf("expected length 3, got %d", len(result))
	}
	// value 相同时按 key 正序排列
	expectedKeys := []string{"a", "b", "c"}
	for i, expected := range expectedKeys {
		if result[i].Key != expected {
			t.Errorf("index %d: expected key %q, got %q", i, expected, result[i].Key)
		}
	}
}

func TestSortByValue_NegativeValues(t *testing.T) {
	m := map[string]int{
		"neg":  -5,
		"zero": 0,
		"pos":  3,
	}
	result := SortByValue(m, true)
	if len(result) != 3 {
		t.Fatalf("expected length 3, got %d", len(result))
	}
	// 正序：-5, 0, 3
	expectedValues := []int{-5, 0, 3}
	for i, expected := range expectedValues {
		if result[i].Value != expected {
			t.Errorf("index %d: expected value %d, got %d", i, expected, result[i].Value)
		}
	}
}

func TestSortByValue_EmptyMap(t *testing.T) {
	m := map[string]int{}
	result := SortByValue(m, true)
	if len(result) != 0 {
		t.Errorf("expected empty result, got length %d", len(result))
	}
}

func TestSortByValue_NilMap(t *testing.T) {
	var m map[string]int
	result := SortByValue(m, false)
	if len(result) != 0 {
		t.Errorf("expected empty result for nil map, got length %d", len(result))
	}
}

func TestSortByValue_SingleElement(t *testing.T) {
	m := map[string]int{"only": 42}
	result := SortByValue(m, true)
	if len(result) != 1 {
		t.Fatalf("expected length 1, got %d", len(result))
	}
	if result[0].Key != "only" || result[0].Value != 42 {
		t.Errorf("expected {only, 42}, got {%s, %d}", result[0].Key, result[0].Value)
	}
}

func TestSortByValue_Float64Float64(t *testing.T) {
	m := map[float64]float64{
		1.1: 9.9,
		2.2: 3.3,
		3.3: 6.6,
	}
	result := SortByValue(m, true)
	expectedValues := []float64{3.3, 6.6, 9.9}
	for i, expected := range expectedValues {
		if result[i].Value != expected {
			t.Errorf("index %d: expected value %f, got %f", i, expected, result[i].Value)
		}
	}
}

// ==================== PairList 接口方法测试 ====================

func TestPairList_Len(t *testing.T) {
	pl := PairList{
		{Key: "a", Value: 1},
		{Key: "b", Value: 2},
	}
	if pl.Len() != 2 {
		t.Errorf("expected Len() = 2, got %d", pl.Len())
	}
}

func TestPairList_Less(t *testing.T) {
	pl := PairList{
		{Key: "a", Value: 1},
		{Key: "b", Value: 2},
	}
	if !pl.Less(0, 1) {
		t.Error("expected Less(0, 1) = true")
	}
	if pl.Less(1, 0) {
		t.Error("expected Less(1, 0) = false")
	}
}

func TestPairList_Swap(t *testing.T) {
	pl := PairList{
		{Key: "a", Value: 1},
		{Key: "b", Value: 2},
	}
	pl.Swap(0, 1)
	if pl[0].Key != "b" || pl[1].Key != "a" {
		t.Errorf("Swap failed: got %v", pl)
	}
}
