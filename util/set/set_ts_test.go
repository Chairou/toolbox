package set

import (
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestSet_New(t *testing.T) {
	s := newTS()

	if s.Size() != 0 {
		t.Error("New: calling without any parameters should create a set with zero size")
	}
}

func TestSet_New_parameters(t *testing.T) {
	s := newTS()
	s.Add("string", "another_string", 1, 3.14)

	if s.Size() != 4 {
		t.Error("New: calling with parameters should create a set with size of four")
	}
}

func TestSet_Add(t *testing.T) {
	s := newTS()
	s.Add(1)
	s.Add(2)
	s.Add(2) // duplicate
	s.Add("fatih")
	s.Add("zeynep")
	s.Add("zeynep") // another duplicate

	if s.Size() != 4 {
		t.Error("Add: items are not unique. The set size should be four")
	}

	if !s.Has(1, 2, "fatih", "zeynep") {
		t.Error("Add: added items are not availabile in the set.")
	}
}

func TestSet_Add_multiple(t *testing.T) {
	s := newTS()
	s.Add("ankara", "san francisco", 3.14)

	if s.Size() != 3 {
		t.Error("Add: items are not unique. The set size should be three")
	}

	if !s.Has("ankara", "san francisco", 3.14) {
		t.Error("Add: added items are not availabile in the set.")
	}
}

func TestSet_Remove(t *testing.T) {
	s := newTS()
	s.Add(1)
	s.Add(2)
	s.Add("fatih")

	s.Remove(1)
	if s.Size() != 2 {
		t.Error("Remove: set size should be two after removing")
	}

	s.Remove(1)
	if s.Size() != 2 {
		t.Error("Remove: set size should be not change after trying to remove a non-existing item")
	}

	s.Remove(2)
	s.Remove("fatih")
	if s.Size() != 0 {
		t.Error("Remove: set size should be zero")
	}

	s.Remove("fatih") // try to remove something from a zero length set
}

func TestSet_Remove_multiple(t *testing.T) {
	s := newTS()
	s.Add("ankara", "san francisco", 3.14, "istanbul")
	s.Remove("ankara", "san francisco", 3.14)

	if s.Size() != 1 {
		t.Error("Remove: items are not unique. The set size should be four")
	}

	if !s.Has("istanbul") {
		t.Error("Add: added items are not availabile in the set.")
	}
}

func TestSet_Pop(t *testing.T) {
	s := newTS()
	s.Add(1)
	s.Add(2)
	s.Add("fatih")

	a := s.Pop()
	if s.Size() != 2 {
		t.Error("Pop: set size should be two after popping out")
	}

	if s.Has(a) {
		t.Error("Pop: returned item should not exist")
	}

	s.Pop()
	s.Pop()
	b := s.Pop()
	if b != nil {
		t.Error("Pop: should return nil because set is empty")
	}

	s.Pop() // try to remove something from a zero length set
}

func TestSet_Has(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3", "4")

	if !s.Has("1") {
		t.Error("Has: the item 1 exist, but 'Has' is returning false")
	}

	if !s.Has("1", "2", "3", "4") {
		t.Error("Has: the items all exist, but 'Has' is returning false")
	}
}

func TestSet_Clear(t *testing.T) {
	s := newTS()
	s.Add(1)
	s.Add("istanbul")
	s.Add("san francisco")

	s.Clear()
	if s.Size() != 0 {
		t.Error("Clear: set size should be zero")
	}
}

func TestSet_IsEmpty(t *testing.T) {
	s := newTS()

	empty := s.IsEmpty()
	if !empty {
		t.Error("IsEmpty: set is empty, it should be true")
	}

	s.Add(2)
	s.Add(3)
	notEmpty := s.IsEmpty()

	if notEmpty {
		t.Error("IsEmpty: set is filled, it should be false")
	}
}

func TestSet_IsEqual(t *testing.T) {
	// same size, same content
	s := newTS()
	s.Add("1", "2", "3")
	u := newTS()
	u.Add("1", "2", "3")

	ok := s.IsEqual(u)
	if !ok {
		t.Error("IsEqual: set s and t are equal. However it returns false")
	}

	// same size, different content
	a := newTS()
	a.Add("1", "2", "3")
	b := newTS()
	b.Add("4", "5", "6")

	ok = a.IsEqual(b)
	if ok {
		t.Error("IsEqual: set a and b are now equal (1). However it returns true")
	}

	// different size, similar content
	a = newTS()
	a.Add("1", "2", "3")
	b = newTS()
	b.Add("1", "2", "3", "4")

	ok = a.IsEqual(b)
	if ok {
		t.Error("IsEqual: set s and t are now equal (2). However it returns true")
	}
}

func TestSet_IsSubset(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3", "4")
	u := newTS()
	u.Add("1", "2", "3")

	ok := s.IsSubset(u)
	if !ok {
		t.Error("IsSubset: u is a subset of s. However it returns false")
	}

	ok = u.IsSubset(s)
	if ok {
		t.Error("IsSubset: s is not a subset of u. However it returns true")
	}
}

func TestSet_IsSuperset(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3", "4")
	u := newTS()
	u.Add("1", "2", "3")

	ok := u.IsSuperset(s)
	if !ok {
		t.Error("IsSuperset: s is a superset of u. However it returns false")
	}

	ok = s.IsSuperset(u)
	if ok {
		t.Error("IsSuperset: u is not a superset of u. However it returns true")
	}
}

func TestSet_String(t *testing.T) {
	s := newTS()
	if s.String() != "[]" {
		t.Errorf("String: output is not what is excepted '%s'", s.String())
	}

	if !strings.HasPrefix(s.String(), "[") {
		t.Error("String: output should begin with a square bracket")
	}

	if !strings.HasSuffix(s.String(), "]") {
		t.Error("String: output should end with a square bracket")
	}
}

func TestSet_List(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3", "4")

	// this returns a slice of interface{}
	if len(s.List()) != 4 {
		t.Error("List: slice size should be four.")
	}

	for _, item := range s.List() {
		r := reflect.TypeOf(item)
		if r.Kind().String() != "string" {
			t.Error("List: slice item should be a string")
		}
	}
}

func TestSet_Copy(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3", "4")
	r := s.Copy()

	if !s.IsEqual(r) {
		t.Error("Copy: set s and r are not equal")
	}
}

func TestSet_Merge(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3")
	r := newTS()
	r.Add("3", "4", "5")
	s.Merge(r)

	if s.Size() != 5 {
		t.Error("Merge: the set doesn't have all items in it.")
	}

	if !s.Has("1", "2", "3", "4", "5") {
		t.Error("Merge: merged items are not availabile in the set.")
	}
}

func TestSet_Separate(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3")
	r := newTS()
	r.Add("3", "5")
	s.Separate(r)

	if s.Size() != 2 {
		t.Error("Separate: the set doesn't have all items in it.")
	}

	if !s.Has("1", "2") {
		t.Error("Separate: items after separation are not availabile in the set.")
	}
}

func TestSet_RaceAdd(t *testing.T) {
	// Create two sets. Add concurrently items to each of them. Remove from the
	// other one.
	// "go test -race" should detect this if the library is not thread-safe.
	s := newTS()
	u := newTS()

	go func() {
		for i := 0; i < 1000; i++ {
			item := "item" + strconv.Itoa(i)
			go func(i int) {
				s.Add(item)
				u.Add(item)
			}(i)
		}
	}()

	for i := 0; i < 1000; i++ {
		item := "item" + strconv.Itoa(i)
		go func(i int) {
			s.Add(item)
			u.Add(item)
		}(i)
	}
}

// ============================================================
// 以下测试用例覆盖 set_ts.go 中所有修正的改动
// ============================================================

// ---------- Pop 竞态条件修复测试 ----------

// TestSet_Pop_Concurrent 测试并发 Pop 不会 panic 或产生数据竞争
func TestSet_Pop_Concurrent(t *testing.T) {
	s := newTS()
	for i := 0; i < 100; i++ {
		s.Add(i)
	}

	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func() {
			s.Pop()
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	// 所有元素都应该被 Pop 出去
	if s.Size() != 0 {
		t.Errorf("Pop_Concurrent: expected size 0, got %d", s.Size())
	}
}

// TestSet_Pop_ConcurrentWithAdd 测试并发 Pop 和 Add 不会 panic
func TestSet_Pop_ConcurrentWithAdd(t *testing.T) {
	s := newTS()
	done := make(chan bool, 200)

	for i := 0; i < 100; i++ {
		go func(val int) {
			s.Add(val)
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		go func() {
			s.Pop()
			done <- true
		}()
	}

	for i := 0; i < 200; i++ {
		<-done
	}
}

// ---------- Copy 加锁修复测试 ----------

// TestSet_Copy_Concurrent 测试并发 Copy 不会 panic（修复前会触发 concurrent map iteration and map write）
func TestSet_Copy_Concurrent(t *testing.T) {
	s := newTS()
	for i := 0; i < 100; i++ {
		s.Add(i)
	}

	done := make(chan bool, 200)

	// 并发 Copy
	for i := 0; i < 100; i++ {
		go func() {
			c := s.Copy()
			_ = c.Size()
			done <- true
		}()
	}

	// 并发修改
	for i := 100; i < 200; i++ {
		go func(val int) {
			s.Add(val)
			done <- true
		}(i)
	}

	for i := 0; i < 200; i++ {
		<-done
	}
}

// TestSet_Copy_ReturnsTS 测试 Copy 返回的是线程安全版本
func TestSet_Copy_ReturnsTS(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3")
	c := s.Copy()

	if _, ok := c.(*Set); !ok {
		t.Error("Copy_ReturnsTS: Copy of a thread-safe set should return a thread-safe set")
	}

	if !s.IsEqual(c) {
		t.Error("Copy_ReturnsTS: copied set should be equal to original")
	}

	// 修改副本不影响原集合
	c.Add("4")
	if s.Has("4") {
		t.Error("Copy_ReturnsTS: modifying copy should not affect original")
	}
}

// ---------- IsEqual 嵌套锁修复测试 ----------

// TestSet_IsEqual_TSvsTS 测试两个线程安全 Set 之间的 IsEqual（走直接访问底层 map 的分支）
func TestSet_IsEqual_TSvsTS(t *testing.T) {
	s := newTS()
	s.Add("a", "b", "c")
	u := newTS()
	u.Add("a", "b", "c")

	if !s.IsEqual(u) {
		t.Error("IsEqual_TSvsTS: equal sets should return true")
	}

	u.Add("d")
	if s.IsEqual(u) {
		t.Error("IsEqual_TSvsTS: different size sets should return false")
	}

	// 相同大小但不同内容
	v := newTS()
	v.Add("a", "b", "d")
	if s.IsEqual(v) {
		t.Error("IsEqual_TSvsTS: same size but different content should return false")
	}
}

// TestSet_IsEqual_TSvsNonTS 测试线程安全 Set 与非线程安全 Set 之间的 IsEqual（走接口方法分支）
func TestSet_IsEqual_TSvsNonTS(t *testing.T) {
	s := newTS()
	s.Add("a", "b", "c")
	u := newNonTS()
	u.Add("a", "b", "c")

	if !s.IsEqual(u) {
		t.Error("IsEqual_TSvsNonTS: equal sets should return true")
	}

	u.Add("d")
	if s.IsEqual(u) {
		t.Error("IsEqual_TSvsNonTS: different size sets should return false")
	}

	// 相同大小但不同内容
	v := newNonTS()
	v.Add("a", "b", "d")
	if s.IsEqual(v) {
		t.Error("IsEqual_TSvsNonTS: same size but different content should return false")
	}
}

// TestSet_IsEqual_Concurrent 测试并发 IsEqual 不会 panic
func TestSet_IsEqual_Concurrent(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3")
	u := newTS()
	u.Add("1", "2", "3")

	done := make(chan bool, 200)

	for i := 0; i < 100; i++ {
		go func() {
			s.IsEqual(u)
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		go func(val int) {
			s.Add(strconv.Itoa(val + 100))
			done <- true
		}(i)
	}

	for i := 0; i < 200; i++ {
		<-done
	}
}

// ---------- IsSubset 嵌套锁修复测试 ----------

// TestSet_IsSubset_TSvsTS 测试两个线程安全 Set 之间的 IsSubset（走直接访问底层 map 的分支）
func TestSet_IsSubset_TSvsTS(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3", "4")
	u := newTS()
	u.Add("1", "2", "3")

	if !s.IsSubset(u) {
		t.Error("IsSubset_TSvsTS: u should be a subset of s")
	}

	if u.IsSubset(s) {
		t.Error("IsSubset_TSvsTS: s should not be a subset of u")
	}

	// 空集是任何集合的子集
	empty := newTS()
	if !s.IsSubset(empty) {
		t.Error("IsSubset_TSvsTS: empty set should be a subset of any set")
	}
}

// TestSet_IsSubset_TSvsNonTS 测试线程安全 Set 与非线程安全 Set 之间的 IsSubset（走接口方法分支）
func TestSet_IsSubset_TSvsNonTS(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3", "4")
	u := newNonTS()
	u.Add("1", "2", "3")

	if !s.IsSubset(u) {
		t.Error("IsSubset_TSvsNonTS: u should be a subset of s")
	}

	v := newNonTS()
	v.Add("1", "2", "5")
	if s.IsSubset(v) {
		t.Error("IsSubset_TSvsNonTS: v should not be a subset of s (contains 5)")
	}
}

// TestSet_IsSubset_Concurrent 测试并发 IsSubset 不会 panic
func TestSet_IsSubset_Concurrent(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3", "4")
	u := newTS()
	u.Add("1", "2")

	done := make(chan bool, 200)

	for i := 0; i < 100; i++ {
		go func() {
			s.IsSubset(u)
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		go func(val int) {
			s.Add(strconv.Itoa(val + 100))
			done <- true
		}(i)
	}

	for i := 0; i < 200; i++ {
		<-done
	}
}

// ---------- IsSuperset 新增线程安全重写测试 ----------

// TestSet_IsSuperset_TSvsTS 测试两个线程安全 Set 之间的 IsSuperset（走直接访问底层 map 的分支）
func TestSet_IsSuperset_TSvsTS(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3", "4")
	u := newTS()
	u.Add("1", "2", "3")

	if !u.IsSuperset(s) {
		t.Error("IsSuperset_TSvsTS: s should be a superset of u")
	}

	if s.IsSuperset(u) {
		t.Error("IsSuperset_TSvsTS: u should not be a superset of s")
	}
}

// TestSet_IsSuperset_TSvsNonTS 测试线程安全 Set 与非线程安全 Set 之间的 IsSuperset（走非 *Set 分支）
func TestSet_IsSuperset_TSvsNonTS(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3")
	u := newNonTS()
	u.Add("1", "2", "3", "4")

	if !s.IsSuperset(u) {
		t.Error("IsSuperset_TSvsNonTS: u should be a superset of s")
	}

	v := newNonTS()
	v.Add("1", "2")
	if s.IsSuperset(v) {
		t.Error("IsSuperset_TSvsNonTS: v should not be a superset of s")
	}
}

// TestSet_IsSuperset_Concurrent 测试并发 IsSuperset 不会 panic
func TestSet_IsSuperset_Concurrent(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3")
	u := newTS()
	u.Add("1", "2", "3", "4")

	done := make(chan bool, 200)

	for i := 0; i < 100; i++ {
		go func() {
			s.IsSuperset(u)
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		go func(val int) {
			u.Add(strconv.Itoa(val + 100))
			done <- true
		}(i)
	}

	for i := 0; i < 200; i++ {
		<-done
	}
}

// ---------- IsEmpty 新增线程安全重写测试 ----------

// TestSet_IsEmpty_Concurrent 测试并发 IsEmpty 不会 panic
func TestSet_IsEmpty_Concurrent(t *testing.T) {
	s := newTS()

	done := make(chan bool, 200)

	for i := 0; i < 100; i++ {
		go func(val int) {
			s.Add(val)
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		go func() {
			_ = s.IsEmpty()
			done <- true
		}()
	}

	for i := 0; i < 200; i++ {
		<-done
	}
}

// ---------- String 新增线程安全重写测试 ----------

// TestSet_String_WithItems 测试 String 方法包含所有元素
func TestSet_String_WithItems(t *testing.T) {
	s := newTS()
	s.Add(1)
	str := s.String()

	if !strings.HasPrefix(str, "[") || !strings.HasSuffix(str, "]") {
		t.Errorf("String_WithItems: output should be wrapped in brackets, got '%s'", str)
	}

	if !strings.Contains(str, "1") {
		t.Errorf("String_WithItems: output should contain '1', got '%s'", str)
	}
}

// TestSet_String_Concurrent 测试并发 String 不会 panic
func TestSet_String_Concurrent(t *testing.T) {
	s := newTS()
	s.Add("a", "b", "c")

	done := make(chan bool, 200)

	for i := 0; i < 100; i++ {
		go func() {
			_ = s.String()
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		go func(val int) {
			s.Add(strconv.Itoa(val))
			done <- true
		}(i)
	}

	for i := 0; i < 200; i++ {
		<-done
	}
}

// ---------- Merge 自合并死锁防护 + 跨类型测试 ----------

// TestSet_Merge_Self 测试 s.Merge(s) 不会死锁，且集合不变
func TestSet_Merge_Self(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3")

	// 如果修复前，这里会死锁
	s.Merge(s)

	if s.Size() != 3 {
		t.Errorf("Merge_Self: size should remain 3, got %d", s.Size())
	}

	if !s.Has("1", "2", "3") {
		t.Error("Merge_Self: items should remain unchanged after self-merge")
	}
}

// TestSet_Merge_TSvsNonTS 测试线程安全 Set 合并非线程安全 Set（走 t.Each 分支）
func TestSet_Merge_TSvsNonTS(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3")
	u := newNonTS()
	u.Add("3", "4", "5")

	s.Merge(u)

	if s.Size() != 5 {
		t.Errorf("Merge_TSvsNonTS: size should be 5, got %d", s.Size())
	}

	if !s.Has("1", "2", "3", "4", "5") {
		t.Error("Merge_TSvsNonTS: merged items are not available in the set")
	}
}

// TestSet_Merge_TSvsTS 测试两个线程安全 Set 合并（走直接访问底层 map 的分支）
func TestSet_Merge_TSvsTS(t *testing.T) {
	s := newTS()
	s.Add("a", "b")
	u := newTS()
	u.Add("b", "c", "d")

	s.Merge(u)

	if s.Size() != 4 {
		t.Errorf("Merge_TSvsTS: size should be 4, got %d", s.Size())
	}

	if !s.Has("a", "b", "c", "d") {
		t.Error("Merge_TSvsTS: merged items are not available in the set")
	}

	// 确保 u 没有被修改
	if u.Size() != 3 {
		t.Errorf("Merge_TSvsTS: source set should not be modified, size=%d", u.Size())
	}
}

// TestSet_Merge_Concurrent 测试并发 Merge 不会 panic
func TestSet_Merge_Concurrent(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3")

	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func(val int) {
			u := newTS()
			u.Add(strconv.Itoa(val))
			s.Merge(u)
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}

// ---------- Separate 新增线程安全重写测试 ----------

// TestSet_Separate_Self 测试 s.Separate(s) 不会死锁，且清空集合
func TestSet_Separate_Self(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3")

	// 如果修复前，这里会死锁
	s.Separate(s)

	if s.Size() != 0 {
		t.Errorf("Separate_Self: size should be 0 after self-separate, got %d", s.Size())
	}
}

// TestSet_Separate_TSvsNonTS 测试线程安全 Set 分离非线程安全 Set（走 t.List 分支）
func TestSet_Separate_TSvsNonTS(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3", "4")
	u := newNonTS()
	u.Add("2", "4", "5")

	s.Separate(u)

	if s.Size() != 2 {
		t.Errorf("Separate_TSvsNonTS: size should be 2, got %d", s.Size())
	}

	if !s.Has("1", "3") {
		t.Error("Separate_TSvsNonTS: remaining items should be 1 and 3")
	}
}

// TestSet_Separate_TSvsTS 测试两个线程安全 Set 分离（走直接访问底层 map 的分支）
func TestSet_Separate_TSvsTS(t *testing.T) {
	s := newTS()
	s.Add("a", "b", "c", "d")
	u := newTS()
	u.Add("b", "d", "e")

	s.Separate(u)

	if s.Size() != 2 {
		t.Errorf("Separate_TSvsTS: size should be 2, got %d", s.Size())
	}

	if !s.Has("a", "c") {
		t.Error("Separate_TSvsTS: remaining items should be a and c")
	}

	// 确保 u 没有被修改
	if u.Size() != 3 {
		t.Errorf("Separate_TSvsTS: source set should not be modified, size=%d", u.Size())
	}
}

// TestSet_Separate_EmptySource 测试从空集合分离
func TestSet_Separate_EmptySource(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3")
	u := newTS()

	s.Separate(u)

	if s.Size() != 3 {
		t.Errorf("Separate_EmptySource: size should remain 3, got %d", s.Size())
	}
}

// TestSet_Separate_Concurrent 测试并发 Separate 不会 panic
func TestSet_Separate_Concurrent(t *testing.T) {
	s := newTS()
	for i := 0; i < 1000; i++ {
		s.Add(i)
	}

	done := make(chan bool, 200)

	for i := 0; i < 100; i++ {
		go func(val int) {
			u := newTS()
			u.Add(val)
			s.Separate(u)
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		go func(val int) {
			s.Add(val + 2000)
			done <- true
		}(i)
	}

	for i := 0; i < 200; i++ {
		<-done
	}
}

// ---------- 综合并发竞争测试 ----------

// TestSet_RaceAllMethods 综合测试所有修改过的方法在并发下不会 panic 或产生数据竞争
func TestSet_RaceAllMethods(t *testing.T) {
	s := newTS()
	u := newTS()
	for i := 0; i < 50; i++ {
		s.Add(i)
		u.Add(i + 25)
	}

	var wg sync.WaitGroup
	n := 10

	// 并发 Add
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			s.Add(val + 1000)
		}(i)
	}

	// 并发 Remove
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			s.Remove(val)
		}(i)
	}

	// 并发 Pop
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Pop()
		}()
	}

	// 并发 Has
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			s.Has(val)
		}(i)
	}

	// 并发 Size + IsEmpty
	for i := 0; i < n; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = s.Size()
		}()
		go func() {
			defer wg.Done()
			_ = s.IsEmpty()
		}()
	}

	// 并发 IsEqual + IsSubset + IsSuperset
	for i := 0; i < n; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			s.IsEqual(u)
		}()
		go func() {
			defer wg.Done()
			s.IsSubset(u)
		}()
		go func() {
			defer wg.Done()
			s.IsSuperset(u)
		}()
	}

	// 并发 Each + String + List + Copy（只读操作）
	for i := 0; i < n; i++ {
		wg.Add(4)
		go func() {
			defer wg.Done()
			s.Each(func(item interface{}) bool { return true })
		}()
		go func() {
			defer wg.Done()
			_ = s.String()
		}()
		go func() {
			defer wg.Done()
			_ = s.List()
		}()
		go func() {
			defer wg.Done()
			_ = s.Copy()
		}()
	}

	// 并发 Merge + Separate（使用独立的新 Set，避免与 s 交叉锁竞争）
	for i := 0; i < n; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			v := newTS()
			v.Add("merge_item")
			s.Merge(v)
		}()
		go func() {
			defer wg.Done()
			v := newTS()
			v.Add("merge_item")
			s.Separate(v)
		}()
	}

	wg.Wait()
}

// TestSet_RaceCopyAndModify 测试并发 Copy 与修改操作不会 panic（修复前 Copy 无锁会 fatal error）
func TestSet_RaceCopyAndModify(t *testing.T) {
	s := newTS()
	for i := 0; i < 100; i++ {
		s.Add(i)
	}

	done := make(chan bool, 300)

	// 并发 Copy
	for i := 0; i < 100; i++ {
		go func() {
			c := s.Copy()
			_ = c.List()
			done <- true
		}()
	}

	// 并发 Add
	for i := 0; i < 100; i++ {
		go func(val int) {
			s.Add(val + 1000)
			done <- true
		}(i)
	}

	// 并发 Remove
	for i := 0; i < 100; i++ {
		go func(val int) {
			s.Remove(val)
			done <- true
		}(i)
	}

	for i := 0; i < 300; i++ {
		<-done
	}
}

// TestSet_RaceMergeSeparate 测试并发 Merge 和 Separate 不会 panic
func TestSet_RaceMergeSeparate(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3", "4", "5")

	done := make(chan bool, 200)

	for i := 0; i < 100; i++ {
		go func(val int) {
			u := newTS()
			u.Add(strconv.Itoa(val))
			s.Merge(u)
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		go func(val int) {
			u := newTS()
			u.Add(strconv.Itoa(val))
			s.Separate(u)
			done <- true
		}(i)
	}

	for i := 0; i < 200; i++ {
		<-done
	}
}

// TestSet_RaceIsEqualBothDirections 测试两个 Set 并发互相 IsEqual 不会死锁
func TestSet_RaceIsEqualBothDirections(t *testing.T) {
	s := newTS()
	s.Add("1", "2", "3")
	u := newTS()
	u.Add("1", "2", "3")

	done := make(chan bool, 200)

	for i := 0; i < 100; i++ {
		go func() {
			s.IsEqual(u)
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		go func() {
			u.IsEqual(s)
			done <- true
		}()
	}

	for i := 0; i < 200; i++ {
		<-done
	}
}