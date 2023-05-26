package singleflight

import (
	"errors"
	"golang.org/x/sync/singleflight"
)

var (
	ErrNotExist = errors.New("Data doesn't exist")
	ErrExist    = errors.New("Duplicate Data")
)

type SFStats struct {
	cacheHit  int
	cacheMiss int
	dbHit     int
}

type SingleFlightPoC struct {
	db    *Db
	cache *Cache
	sf    singleflight.Group
	stat  SFStats
}

func NewSingleFlightPoC(db *Db, c *Cache) *SingleFlightPoC {
	return &SingleFlightPoC{
		db:    db,
		cache: c,
	}
}

func (s *SingleFlightPoC) Get(key string) (string, error) {
	if v, errCache := s.cache.Get(key); errCache == nil {
		s.stat.cacheHit++
		switch v[0] {
		case 'S':
			return v[2:], nil
		case 'E':
			return "", ErrNotExist
		}
		return "", ErrNotExist
	}
	s.stat.cacheMiss++

	v, err, _ := s.sf.Do(key, func() (interface{}, error) {
		s.stat.dbHit++
		return s.db.Select(key)
	})

	if err != nil {
		s.cache.Set(key, "E|"+err.Error())
	} else {
		s.cache.Set(key, "S|"+v.(string))
	}

	return v.(string), err
}

func (s *SingleFlightPoC) getStats() SFStats {
	return s.stat
}
