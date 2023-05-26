package singleflight

import (
	"reflect"
	"time"
)

type Db struct {
	data  map[string]string
	delay time.Duration
}

func NewDb(delayInMs time.Duration) *Db {
	return &Db{
		data:  make(map[string]string),
		delay: delayInMs,
	}
}

func (db *Db) Insert(key, value string) error {
	time.Sleep(time.Millisecond * db.delay)
	if _, ok := db.data[key]; ok {
		return ErrExist
	}
	db.data[key] = value

	return nil
}

func (db *Db) Select(key string) (string, error) {
	time.Sleep(time.Millisecond * db.delay)
	if value, ok := db.data[key]; ok {
		return value, nil
	}

	return "", ErrNotExist
}

func (db *Db) eqInternalData(e map[string]string) bool {
	return reflect.DeepEqual(db.data, e)
}
