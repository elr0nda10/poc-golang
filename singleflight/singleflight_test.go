package singleflight

import (
	"errors"
	"sync"
	"testing"
	"time"
)

type keyValueTest struct {
	key string
	val string
}

type IOETest struct {
	input          string
	expectedOutput string
	expectedError  error
}

func generateTestData(db *Db) {
	testData := []keyValueTest{
		{"k1", "v1"},
		{"k2", "v2"},
		{"k3", "v3"},
	}
	for _, td := range testData {
		_ = db.Insert(td.key, td.val)
	}
}

func generateTestTrafficData() []IOETest {
	trafficTypes := []IOETest{
		{"k1", "v1", nil},       // normal data 1
		{"k2", "v2", nil},       // normal data 2
		{"k3", "v3", nil},       // normal data 3
		{"k4", "", ErrNotExist}, // non exist data
	}

	var traffics []IOETest
	for i := 0; i < 5; i++ {
		for _, tType := range trafficTypes {
			traffics = append(traffics, tType)
		}
	}

	return traffics
}

func testTrafficAgainstSF(sf *SingleFlightPoC, traffic *IOETest, t *testing.T) {
	res, err := sf.Get(traffic.input)
	if !errors.Is(err, traffic.expectedError) {
		t.Fatalf("for key %v: expecting to get error \"%v\" but get \"%v\" instead",
			traffic.input,
			traffic.expectedError,
			err)
	}
	if err == nil && res != traffic.expectedOutput {
		t.Fatalf("for key %v: expecting to get \"%v\" but get \"%v\" instead",
			traffic.input,
			traffic.expectedOutput,
			res)
	}
}

func testStat(expectedStat, realStat SFStats, t *testing.T) {
	if realStat.cacheHit != expectedStat.cacheHit ||
		realStat.cacheMiss != expectedStat.cacheMiss ||
		realStat.dbHit != expectedStat.dbHit {
		t.Fatalf("expecting (cacheHit,cacheMiss,dbHit) to be (%v,%v,%v) but got (%v,%v,%v) instead",
			expectedStat.cacheHit, expectedStat.cacheMiss, expectedStat.dbHit,
			realStat.cacheHit, realStat.cacheMiss, realStat.dbHit)
	}
}

func TestSingleFlightPoC_Get(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{{
		name: "Data Correctness",
		fn:   getDataCorrectness,
	}, {
		name: "Sequential Traffic",
		fn:   testSequentialTraffic,
	}, {
		name: "Burst Traffic",
		fn:   testBurstTraffic,
	}}

	for _, test := range tests {
		t.Run(test.name, test.fn)
	}
}

func getDataCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		testData []IOETest
	}{{
		name: "Get one data",
		testData: []IOETest{
			{"k1", "v1", nil},
		},
	}, {
		name: "Get one data twice",
		testData: []IOETest{
			{"k1", "v1", nil},
			{"k1", "v1", nil},
		},
	}, {
		name: "Get data that doesn't exist",
		testData: []IOETest{
			{"k4", "", ErrNotExist}, // non exist data
		},
	}, {
		name: "Get data that doesn't exist twice",
		testData: []IOETest{
			{"k4", "", ErrNotExist}, // non exist data
			{"k4", "", ErrNotExist}, // non exist data
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cache := NewCache()
			db := NewDb(0)
			generateTestData(db)
			sf := NewSingleFlightPoC(db, cache)
			for _, testSearch := range test.testData {
				testTrafficAgainstSF(sf, &testSearch, t)
			}
		})
	}
}

func testSequentialTraffic(t *testing.T) {
	cache := NewCache()
	db := NewDb(time.Duration(500))
	generateTestData(db)
	sfTest := NewSingleFlightPoC(db, cache)
	traffics := generateTestTrafficData()

	for _, traffic := range traffics {
		testTrafficAgainstSF(sfTest, &traffic, t)
	}

	testStat(SFStats{
		cacheHit:  4 * 4,
		cacheMiss: 4 * 1,
		dbHit:     4 * 1,
	}, sfTest.getStats(), t)
}

func testBurstTraffic(t *testing.T) {
	cache := NewCache()
	db := NewDb(time.Duration(1000))
	generateTestData(db)
	sfTest := NewSingleFlightPoC(db, cache)
	traffics := generateTestTrafficData()

	wg := sync.WaitGroup{}
	for _, traffic := range traffics {
		wg.Add(1)
		go func(traffic IOETest) {
			defer wg.Done()
			testTrafficAgainstSF(sfTest, &traffic, t)
		}(traffic)
	}

	wg.Wait()

	testStat(SFStats{
		cacheHit:  4 * 0,
		cacheMiss: 4 * 5,
		dbHit:     4 * 1,
	}, sfTest.getStats(), t)
}
