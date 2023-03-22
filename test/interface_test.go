package test

import (
	"sync"
	"testing"
)

type Cache interface {
	Put(key, value string) error
	Get(key string) (string, error)
}

type cache struct {
	mu   sync.Mutex
	data map[string]string
}

func (c *cache) Put(key, value string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
	return nil
}

func (c *cache) Get(key string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data[key], nil
}

func newCache() cache {
	tmpCache := cache{}
	tmpCache.data = make(map[string]string, 0)
	return tmpCache
}

func TestCache(t *testing.T) {
	tmpCache := newCache()
	tmpCache.Put("foo", "bar")
	asd, err := tmpCache.Get("foo")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(asd)
	}
}
