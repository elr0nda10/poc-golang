package singleflight

import (
	"reflect"
	"sync"
	"time"
)

type Db struct {
	data  map[string]string
	delay time.Duration
	mu    sync.RWMutex
}

func NewDb(delayInMs time.Duration) *Db {
	return &Db{
		data:  make(map[string]string),
		delay: delayInMs,
	}
}

func (db *Db) Insert(key, value string) error {
	time.Sleep(time.Millisecond * db.delay)
	db.mu.RLock()
	if _, ok := db.data[key]; ok {
		db.mu.RUnlock()
		return ErrExist
	}
	db.mu.RUnlock()

	db.mu.Lock()
	db.data[key] = value
	db.mu.Unlock()

	return nil
}

func (db *Db) Select(key string) (string, error) {
	time.Sleep(time.Millisecond * db.delay)
	db.mu.RLock()
	defer db.mu.RUnlock()
	if value, ok := db.data[key]; ok {
		return value, nil
	}

	return "", ErrNotExist
}

func (db *Db) eqInternalData(e map[string]string) bool {
	return reflect.DeepEqual(db.data, e)
}
