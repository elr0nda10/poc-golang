package singleflight

import "reflect"

type Cache struct {
	data map[string]string
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]string),
	}
}

func (c *Cache) Set(key, val string) {
	c.data[key] = val
}

func (c *Cache) Get(key string) (string, error) {
	if value, ok := c.data[key]; ok {
		return value, nil
	}

	return "", ErrNotExist
}

func (c *Cache) Clear() {
	c.data = make(map[string]string)
}

func (c *Cache) eqInternalData(e map[string]string) bool {
	return reflect.DeepEqual(c.data, e)
}
