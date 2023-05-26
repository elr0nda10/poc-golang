package singleflight

import (
	"errors"
	"golang.org/x/sync/singleflight"
	"sync"
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
	mu    sync.RWMutex
}

func NewSingleFlightPoC(db *Db, c *Cache) *SingleFlightPoC {
	return &SingleFlightPoC{
		db:    db,
		cache: c,
	}
}

func (s *SingleFlightPoC) Get(key string) (string, error) {
	if v, errCache := s.cache.Get(key); errCache == nil {
		s.cacheHitInc()
		switch v[0] {
		case 'S':
			return v[2:], nil
		case 'E':
			return "", ErrNotExist
		}
		return "", ErrNotExist
	}
	s.cacheMissInc()

	v, err, _ := s.sf.Do(key, func() (interface{}, error) {
		s.dbHitInc()
		return s.db.Select(key)
	})

	if err != nil {
		s.cache.Set(key, "E|"+err.Error())
	} else {
		s.cache.Set(key, "S|"+v.(string))
	}

	return v.(string), err
}

func (s *SingleFlightPoC) cacheHitInc() {
	s.mu.Lock()
	s.stat.cacheHit++
	s.mu.Unlock()
}

func (s *SingleFlightPoC) cacheMissInc() {
	s.mu.Lock()
	s.stat.cacheMiss++
	s.mu.Unlock()
}

func (s *SingleFlightPoC) dbHitInc() {
	s.mu.Lock()
	s.stat.dbHit++
	s.mu.Unlock()
}

func (s *SingleFlightPoC) getStats() SFStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stat
}
