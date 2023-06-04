package singleflight

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

func (c *Cache) test_internalDataStructure(expectedVal map[string]string, t *testing.T) {
	defer c.mu.Unlock()
	c.mu.Lock()
	if isEq := reflect.DeepEqual(c.data, expectedVal); !isEq {
		t.Fatalf("expecting total data %v but got %v", expectedVal, c.data)
	}
}

func TestCache_Set(t *testing.T) {
	tests := []struct {
		name     string
		input    []testKeyValue
		expected testExpectedOutput
	}{{
		name: "Insert 1 data",
		input: []testKeyValue{
			{"k1", "v1"},
		},
		expected: testExpectedOutput{
			Val: map[string]string{
				"k1": "v1",
			},
			Err: nil,
		},
	}, {
		name: "Insert 2 datas",
		input: []testKeyValue{
			{"k1", "v1"},
			{"k2", "v2"},
		},
		expected: testExpectedOutput{
			Val: map[string]string{
				"k1": "v1",
				"k2": "v2",
			},
			Err: nil,
		},
	}, {
		name: "Insert 2 datas with same key",
		input: []testKeyValue{
			{"k1", "v1"},
			{"k1", "v2"},
		},
		expected: testExpectedOutput{
			Val: map[string]string{
				"k1": "v2",
			},
			Err: nil,
		},
	}, {
		name: "Insert 2 datas with same value",
		input: []testKeyValue{
			{"k1", "v1"},
			{"k2", "v1"},
		},
		expected: testExpectedOutput{
			Val: map[string]string{
				"k1": "v1",
				"k2": "v1",
			},
			Err: nil,
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cache := NewCache()
			for _, ti := range test.input {
				cache.Set(ti.Key, ti.Val)
			}

			cache.test_internalDataStructure(test.expected.Val.(map[string]string), t)
		})
	}
}

func TestCache_Get(t *testing.T) {
	cache := NewCache()
	for _, testData := range []testKeyValue{
		{"k1", "v1"},
		{"k2", "v2"},
		{"k1", "v3"},
	} {
		cache.Set(testData.Key, testData.Val)
	}

	tests := []struct {
		name     string
		input    string
		expected testExpectedOutput
	}{{
		name:  "element exist",
		input: "k2",
		expected: testExpectedOutput{
			Val: "v2",
			Err: nil,
		},
	}, {
		name:  "duplicate key should return latest data",
		input: "k1",
		expected: testExpectedOutput{
			Val: "v3",
			Err: nil,
		},
	}, {
		name:  "element not exist should return ErrNotExist",
		input: "k5",
		expected: testExpectedOutput{
			Val: "",
			Err: ErrNotExist,
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := cache.Get(test.input)
			if !errors.Is(err, test.expected.Err) {
				t.Fatalf("expected error to be %v but got %v", test.expected.Err, err)
			}
			if err == nil && res != test.expected.Val {
				t.Fatalf("expected value to be %v but got %v", test.expected.Val, res)
			}
		})
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache()
	for _, testData := range []testKeyValue{
		{"k1", "v1"},
		{"k2", "v2"},
		{"k1", "v3"},
	} {
		cache.Set(testData.Key, testData.Val)
	}

	cache.Clear()
	cache.test_internalDataStructure(map[string]string{}, t)
}

func TestCache_Race(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("expected to not race but got race instead")
		}
	}()

	cache := NewCache()
	testData := []testKeyValue{
		{"k1", "v1"},
		{"k2", "v2"},
		{"k3", "v3"},
	}
	wg := sync.WaitGroup{}
	fnCacheSet := func(test testKeyValue) {
		defer wg.Done()
		cache.Set(test.Key, test.Val)
	}
	fnCacheGet := func(test testKeyValue) {
		defer wg.Done()
		_, _ = cache.Get(test.Key)
	}
	fnCacheClear := func(test testKeyValue) {
		defer wg.Done()
		cache.Clear()
	}
	for i := 0; i < 100; i++ {
		for _, test := range testData {
			wg.Add(3)
			go fnCacheSet(test)
			go fnCacheGet(test)
			go fnCacheClear(test)
		}
	}

	wg.Wait()
}
