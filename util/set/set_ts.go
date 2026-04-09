package set

import (
	"fmt"
	"strings"
	"sync"
)

// Set defines a thread safe set data structure.
type Set struct {
	set
	l sync.RWMutex // we name it because we don't want to expose it
}

// New creates and initialize a new Set. It's accept a variable number of
// arguments to populate the initial set. If nothing passed a Set with zero
// size is created.
func newTS() *Set {
	s := &Set{}
	s.m = make(map[interface{}]struct{})

	// Ensure interface compliance
	var _ Interface = s

	return s
}

// Add includes the specified items (one or more) to the set. The underlying
// Set s is modified. If passed nothing it silently returns.
func (s *Set) Add(items ...interface{}) {
	s.l.Lock()
	defer s.l.Unlock()
	if len(items) == 0 {
		return
	}

	for _, item := range items {
		s.m[item] = keyExists
	}
}

// Remove deletes the specified items from the set.  The underlying Set s is
// modified. If passed nothing it silently returns.
func (s *Set) Remove(items ...interface{}) {
	s.l.Lock()
	defer s.l.Unlock()
	if len(items) == 0 {
		return
	}

	for _, item := range items {
		delete(s.m, item)
	}
}

// Pop deletes and return an item from the set. The underlying Set s is
// modified. If set is empty, nil is returned.
// 修复：直接使用写锁，消除读锁→释放→写锁之间的竞态窗口
func (s *Set) Pop() interface{} {
	s.l.Lock()
	defer s.l.Unlock()

	for item := range s.m {
		delete(s.m, item)
		return item
	}
	return nil
}

// Has looks for the existence of items passed. It returns false if nothing is
// passed. For multiple items it returns true only if all of  the items exist.
func (s *Set) Has(items ...interface{}) bool {
	s.l.RLock()
	defer s.l.RUnlock()
	// assume checked for empty item, which not exist
	if len(items) == 0 {
		return false
	}

	has := true
	for _, item := range items {
		if _, has = s.m[item]; !has {
			break
		}
	}
	return has
}

// Size returns the number of items in a set.
func (s *Set) Size() int {
	s.l.RLock()
	defer s.l.RUnlock()

	l := len(s.m)
	return l
}

// Clear removes all items from the set.
func (s *Set) Clear() {
	s.l.Lock()
	defer s.l.Unlock()

	s.m = make(map[interface{}]struct{})
}

// IsEmpty reports whether the Set is empty.
// 修复：新增线程安全重写，原实现继承自 set 无锁保护
func (s *Set) IsEmpty() bool {
	s.l.RLock()
	defer s.l.RUnlock()

	return len(s.m) == 0
}

// IsEqual test whether s and t are the same in size and have the same items.
// 修复：当 t 是 *Set 时直接访问底层 map，避免嵌套锁
func (s *Set) IsEqual(t Interface) bool {
	s.l.RLock()
	defer s.l.RUnlock()

	// 如果 t 也是线程安全的 *Set，直接访问底层 map，避免嵌套锁
	if conv, ok := t.(*Set); ok {
		conv.l.RLock()
		defer conv.l.RUnlock()

		if len(s.m) != len(conv.m) {
			return false
		}
		for item := range conv.m {
			if _, ok := s.m[item]; !ok {
				return false
			}
		}
		return true
	}

	// 非线程安全版本，通过接口方法访问
	if sameSize := len(s.m) == t.Size(); !sameSize {
		return false
	}

	equal := true
	t.Each(func(item interface{}) bool {
		_, equal = s.m[item]
		return equal // if false, Each() will end
	})

	return equal
}

// IsSubset tests whether t is a subset of s.
// 修复：当 t 是 *Set 时直接访问底层 map，避免嵌套锁
func (s *Set) IsSubset(t Interface) (subset bool) {
	s.l.RLock()
	defer s.l.RUnlock()

	// 如果 t 也是线程安全的 *Set，直接访问底层 map，避免嵌套锁
	if conv, ok := t.(*Set); ok {
		conv.l.RLock()
		defer conv.l.RUnlock()

		for item := range conv.m {
			if _, ok := s.m[item]; !ok {
				return false
			}
		}
		return true
	}

	subset = true
	t.Each(func(item interface{}) bool {
		_, subset = s.m[item]
		return subset
	})

	return
}

// IsSuperset tests whether t is a superset of s.
// 修复：新增线程安全重写，原实现继承自 set 无锁保护
func (s *Set) IsSuperset(t Interface) bool {
	s.l.RLock()
	defer s.l.RUnlock()

	// 如果 t 也是线程安全的 *Set，直接访问底层 map，避免嵌套锁
	if conv, ok := t.(*Set); ok {
		conv.l.RLock()
		defer conv.l.RUnlock()

		for item := range s.m {
			if _, ok := conv.m[item]; !ok {
				return false
			}
		}
		return true
	}

	// 非线程安全版本，需要释放锁后委托调用，避免死锁
	// 先复制 s 的元素列表
	items := make([]interface{}, 0, len(s.m))
	for item := range s.m {
		items = append(items, item)
	}
	s.l.RUnlock()

	// 检查 s 的所有元素是否都在 t 中
	result := true
	for _, item := range items {
		if !t.Has(item) {
			result = false
			break
		}
	}

	s.l.RLock() // 重新获取锁以匹配 defer RUnlock
	return result
}

// Each traverses the items in the Set, calling the provided function for each
// set member. Traversal will continue until all items in the Set have been
// visited, or if the closure returns false.
func (s *Set) Each(f func(item interface{}) bool) {
	s.l.RLock()
	defer s.l.RUnlock()

	for item := range s.m {
		if !f(item) {
			break
		}
	}
}

// String returns a string representation of s
// 修复：新增线程安全重写，原实现继承自 set 无锁保护
func (s *Set) String() string {
	s.l.RLock()
	defer s.l.RUnlock()

	items := make([]string, 0, len(s.m))
	for item := range s.m {
		items = append(items, fmt.Sprintf("%v", item))
	}

	return fmt.Sprintf("[%s]", strings.Join(items, ", "))
}

// List returns a slice of all items. There is also StringSlice() and
// IntSlice() methods for returning slices of type string or int.
func (s *Set) List() []interface{} {
	s.l.RLock()
	defer s.l.RUnlock()

	list := make([]interface{}, 0, len(s.m))

	for item := range s.m {
		list = append(list, item)
	}

	return list
}

// Copy returns a new Set with a copy of s.
// 修复：添加读锁保护遍历，防止并发 map 读写 panic
func (s *Set) Copy() Interface {
	s.l.RLock()
	defer s.l.RUnlock()

	u := newTS()
	for item := range s.m {
		u.m[item] = keyExists
	}
	return u
}

// Merge is like Union, however it modifies the current set it's applied on
// with the given t set.
// 修复：当 t 是 *Set 时直接访问底层 map，避免嵌套锁；防止自合并死锁
func (s *Set) Merge(t Interface) {
	s.l.Lock()
	defer s.l.Unlock()

	// 如果 t 也是线程安全的 *Set，直接访问底层 map
	if conv, ok := t.(*Set); ok {
		// 防止自合并死锁：s.Merge(s) 是空操作
		if conv == s {
			return
		}
		conv.l.RLock()
		defer conv.l.RUnlock()

		for item := range conv.m {
			s.m[item] = keyExists
		}
		return
	}

	t.Each(func(item interface{}) bool {
		s.m[item] = keyExists
		return true
	})
}

// Separate removes the set items containing in t from set s.
// 修复：新增线程安全重写，原实现继承自 set 无锁保护
func (s *Set) Separate(t Interface) {
	s.l.Lock()
	defer s.l.Unlock()

	// 如果 t 也是线程安全的 *Set，直接访问底层 map
	if conv, ok := t.(*Set); ok {
		// 防止自分离死锁：s.Separate(s) 等价于清空集合
		if conv == s {
			s.m = make(map[interface{}]struct{})
			return
		}
		conv.l.RLock()
		defer conv.l.RUnlock()

		for item := range conv.m {
			delete(s.m, item)
		}
		return
	}

	// 非线程安全版本，获取列表后逐个删除
	items := t.List()
	for _, item := range items {
		delete(s.m, item)
	}
}
