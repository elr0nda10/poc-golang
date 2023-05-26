package singleflight

import (
	"reflect"
	"sync"
)

type Cache struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]string),
	}
}

func (c *Cache) Set(key, val string) {
	c.mu.Lock()
	c.data[key] = val
	c.mu.Unlock()
}

func (c *Cache) Get(key string) (string, error) {
	defer c.mu.RUnlock()
	c.mu.RLock()
	if value, ok := c.data[key]; ok {
		return value, nil
	}

	return "", ErrNotExist
}

func (c *Cache) Clear() {
	c.mu.Lock()
	c.data = make(map[string]string)
	c.mu.Unlock()
}

func (c *Cache) eqInternalData(e map[string]string) bool {
	defer c.mu.Unlock()
	c.mu.Lock()
	return reflect.DeepEqual(c.data, e)
}
