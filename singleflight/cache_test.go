package singleflight

import (
	"errors"
	"sync"
	"testing"
)

func TestCache_Set(t *testing.T) {
	tests := []struct {
		name  string
		input []keyValueTest
		data  map[string]string
	}{{
		name: "Insert 1 data",
		input: []keyValueTest{
			{"k1", "v1"},
		},
		data: map[string]string{
			"k1": "v1",
		},
	}, {
		name: "Insert 2 datas",
		input: []keyValueTest{
			{"k1", "v1"},
			{"k2", "v2"},
		},
		data: map[string]string{
			"k1": "v1",
			"k2": "v2",
		},
	}, {
		name: "Insert 2 datas with same key",
		input: []keyValueTest{
			{"k1", "v1"},
			{"k1", "v2"},
		},
		data: map[string]string{
			"k1": "v2",
		},
	}, {
		name: "Insert 2 datas with same value",
		input: []keyValueTest{
			{"k1", "v1"},
			{"k2", "v1"},
		},
		data: map[string]string{
			"k1": "v1",
			"k2": "v1",
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cache := NewCache()
			for _, ti := range test.input {
				cache.Set(ti.key, ti.val)
			}

			contentEq := cache.eqInternalData(test.data)
			if !contentEq {
				t.Fatalf("expecting total data %v but got %v", test.data, cache.data)
			}
		})
	}
}

func TestCache_Get(t *testing.T) {
	cache := NewCache()
	for _, testData := range []keyValueTest{
		{"k1", "v1"},
		{"k2", "v2"},
		{"k1", "v3"},
	} {
		cache.Set(testData.key, testData.val)
	}

	tests := []struct {
		name          string
		input         string
		expected      string
		expectedError error
	}{{
		name:          "element exist",
		input:         "k2",
		expected:      "v2",
		expectedError: nil,
	}, {
		name:          "duplicate key should return latest data",
		input:         "k1",
		expected:      "v3",
		expectedError: nil,
	}, {
		name:          "element not exist should return ErrNotExist",
		input:         "k5",
		expected:      "",
		expectedError: ErrNotExist,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := cache.Get(test.input)
			if !errors.Is(err, test.expectedError) {
				t.Fatalf("expected error to be %v but got %v", test.expectedError, err)
			}
			if err == nil && res != test.expected {
				t.Fatalf("expected value to be %v but got %v", test.expected, res)
			}
		})
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache()
	for _, testData := range []keyValueTest{
		{"k1", "v1"},
		{"k2", "v2"},
		{"k1", "v3"},
	} {
		cache.Set(testData.key, testData.val)
	}

	cache.Clear()
	expectedData := map[string]string{}
	if !cache.eqInternalData(expectedData) {
		t.Fatalf("expecting value to be %v but got %v", expectedData, cache.data)
	}
}

func TestCache_Race(t *testing.T) {
	cache := NewCache()
	testData := []keyValueTest{
		{"k1", "v1"},
		{"k2", "v2"},
		{"k3", "v3"},
	}
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		for _, test := range testData {
			wg.Add(3)
			go func(test keyValueTest) {
				defer wg.Done()
				cache.Set(test.key, test.val)
			}(test)
			go func(test keyValueTest) {
				defer wg.Done()
				_, _ = cache.Get(test.key)
			}(test)
			go func(test keyValueTest) {
				defer wg.Done()
				cache.Clear()
			}(test)
		}
	}

	wg.Wait()
}
