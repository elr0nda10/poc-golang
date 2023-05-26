package singleflight

import (
	"errors"
	"sync"
	"testing"
	"time"
)

const windowsInaccuracy = 17 * time.Millisecond

type sleepTest struct {
	expectedDuration time.Duration
	realDuration     time.Duration
	startTime        time.Time
}

func newSleepTest(expectedDuration time.Duration) *sleepTest {
	return &sleepTest{
		expectedDuration: expectedDuration,
	}
}

func (s *sleepTest) start() {
	s.startTime = time.Now()
}

func (s *sleepTest) finish() {
	s.realDuration = time.Now().Sub(s.startTime)
}

func (s *sleepTest) isFinishCorrectly() bool {
	lowerBound := s.expectedDuration - windowsInaccuracy
	upperBound := s.expectedDuration + windowsInaccuracy

	return s.realDuration >= lowerBound && s.realDuration <= upperBound
}

func TestDb_InsertDuration(t *testing.T) {
	t.Run("Insert should be run in 1000 ms", func(t *testing.T) {
		delayInMs := time.Duration(1000)
		db := NewDb(delayInMs)
		timerTest := newSleepTest(delayInMs * time.Millisecond)
		timerTest.start()
		_ = db.Insert("k1", "v1")
		timerTest.finish()
		if !timerTest.isFinishCorrectly() {
			t.Fatalf("expected sleep for %v but got %v instead",
				timerTest.expectedDuration,
				timerTest.realDuration)
		}
	})
}

func TestDb_Insert(t *testing.T) {
	tests := []struct {
		name          string
		data          []keyValueTest
		expectedData  map[string]string
		expectedError error
	}{{
		name: "Insert one data",
		data: []keyValueTest{
			{"k1", "v1"},
		},
		expectedData: map[string]string{
			"k1": "v1",
		},
		expectedError: nil,
	}, {
		name: "Insert two data",
		data: []keyValueTest{
			{"k1", "v1"},
			{"k2", "v2"},
		},
		expectedData: map[string]string{
			"k1": "v1",
			"k2": "v2",
		},
		expectedError: nil,
	}, {
		name: "Insert duplicate data",
		data: []keyValueTest{
			{"k1", "v1"},
			{"k1", "v2"},
		},
		expectedData: map[string]string{
			"k1": "v1",
		},
		expectedError: ErrExist,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := NewDb(0)
			var err error = nil
			for _, input := range test.data {
				err = db.Insert(input.key, input.val)
			}
			if !errors.Is(err, test.expectedError) {
				t.Fatalf("expecting error \"%v\" but got \"%v\" instead", test.expectedError, err)
			}
			if !db.eqInternalData(test.expectedData) {
				t.Fatalf("expecting internal data to be %v but got %v instead", test.expectedData, db.data)
			}
		})
	}
}

func TestDb_SelectDuration(t *testing.T) {
	t.Run("Select should be done in 1000 ms", func(t *testing.T) {
		delayInMs := time.Duration(1000)
		db := NewDb(delayInMs)
		timerTest := newSleepTest(delayInMs * time.Millisecond)
		timerTest.start()
		_, _ = db.Select("k1")
		timerTest.finish()
		if !timerTest.isFinishCorrectly() {
			t.Fatalf("expected sleep for %v but got %v instead",
				timerTest.expectedDuration,
				timerTest.realDuration)
		}
	})
}

func TestDb_Select(t *testing.T) {
	testData := []keyValueTest{
		{"k1", "v1"},
		{"k2", "v2"},
		{"k1", "v3"},
	}
	db := NewDb(0)
	for _, data := range testData {
		_ = db.Insert(data.key, data.val)
	}

	tests := []struct {
		name           string
		input          string
		expectedOutput string
		expectedError  error
	}{{
		name:           "Get single data",
		input:          "k2",
		expectedOutput: "v2",
		expectedError:  nil,
	}, {
		name:           "Get duplicate data",
		input:          "k1",
		expectedOutput: "v1",
		expectedError:  nil,
	}, {
		name:           "Get non exist data",
		input:          "k3",
		expectedOutput: "",
		expectedError:  ErrNotExist,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := db.Select(test.input)
			if !errors.Is(err, test.expectedError) {
				t.Fatalf("expecting error \"%v\" but got \"%v\" instead", test.expectedError, err)
			}
			if res != test.expectedOutput {
				t.Fatalf("expecting data \"%v\" but got \"%v\" instead", test.expectedOutput, res)
			}
		})
	}
}

func TestDb_Race(t *testing.T) {
	db := NewDb(0)
	testData := []keyValueTest{
		{"k1", "v1"},
		{"k2", "v2"},
		{"k3", "v3"},
	}
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		for _, test := range testData {
			wg.Add(2)
			go func(test keyValueTest) {
				defer wg.Done()
				_ = db.Insert(test.key, test.val)
			}(test)
			go func(test keyValueTest) {
				defer wg.Done()
				_, _ = db.Select(test.key)
			}(test)
		}
	}

	wg.Wait()
}
