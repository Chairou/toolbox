package set

import (
	"fmt"
	"strings"
)

// Provides a common set baseline for both threadsafe and non-ts Sets.
type set struct {
	m map[interface{}]struct{} // struct{} doesn't take up space
}

// SetNonTS defines a non-thread safe set data structure.
type SetNonTS struct {
	set
}

// NewNonTS creates and initializes a new non-threadsafe Set.
func newNonTS() *SetNonTS {
	s := &SetNonTS{}
	s.m = make(map[interface{}]struct{})

	// Ensure interface compliance
	var _ Interface = s

	return s
}

// Add includes the specified items (one or more) to the set. The underlying
// Set s is modified. If passed nothing it silently returns.
func (s *set) Add(items ...interface{}) {
	if len(items) == 0 {
		return
	}

	for _, item := range items {
		s.m[item] = keyExists
	}
}

// Remove deletes the specified items from the set.  The underlying Set s is
// modified. If passed nothing it silently returns.
func (s *set) Remove(items ...interface{}) {
	if len(items) == 0 {
		return
	}

	for _, item := range items {
		delete(s.m, item)
	}
}

// Pop  deletes and return an item from the set. The underlying Set s is
// modified. If set is empty, nil is returned.
func (s *set) Pop() interface{} {
	for item := range s.m {
		delete(s.m, item)
		return item
	}
	return nil
}

// Has looks for the existence of items passed. It returns false if nothing is
// passed. For multiple items it returns true only if all of  the items exist.
func (s *set) Has(items ...interface{}) bool {
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
func (s *set) Size() int {
	return len(s.m)
}

// Clear removes all items from the set.
func (s *set) Clear() {
	s.m = make(map[interface{}]struct{})
}

// IsEmpty reports whether the Set is empty.
func (s *set) IsEmpty() bool {
	return s.Size() == 0
}

// IsEqual test whether s and t are the same in size and have the same items.
func (s *set) IsEqual(t Interface) bool {
	// 如果 t 是线程安全的 *Set，直接访问底层 map，避免嵌套锁
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
func (s *set) IsSubset(t Interface) (subset bool) {
	// 如果 t 是线程安全的 *Set，直接访问底层 map，避免嵌套锁
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
func (s *set) IsSuperset(t Interface) bool {
	return t.IsSubset(s)
}

// Each traverses the items in the Set, calling the provided function for each
// set member. Traversal will continue until all items in the Set have been
// visited, or if the closure returns false.
func (s *set) Each(f func(item interface{}) bool) {
	for item := range s.m {
		if !f(item) {
			break
		}
	}
}

// Copy returns a new Set with a copy of s.
func (s *set) Copy() Interface {
	u := newNonTS()
	for item := range s.m {
		u.Add(item)
	}
	return u
}

// String returns a string representation of s
func (s *set) String() string {
	list := s.List()
	t := make([]string, 0, len(list))
	for _, item := range list {
		t = append(t, fmt.Sprintf("%v", item))
	}

	return fmt.Sprintf("[%s]", strings.Join(t, ", "))
}

// List returns a slice of all items. There is also StringSlice() and
// IntSlice() methods for returning slices of type string or int.
func (s *set) List() []interface{} {
	list := make([]interface{}, 0, len(s.m))

	for item := range s.m {
		list = append(list, item)
	}

	return list
}

// Merge is like Union, however it modifies the current set it's applied on
// with the given t set.
func (s *set) Merge(t Interface) {
	t.Each(func(item interface{}) bool {
		s.m[item] = keyExists
		return true
	})
}

// Separate removes the set items containing in t from set s.
// It's not the opposite of Merge. Items in t that are not in s are ignored.
func (s *set) Separate(t Interface) {
	s.Remove(t.List()...)
}
