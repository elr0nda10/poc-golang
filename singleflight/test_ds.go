package singleflight

type testKeyValue struct {
	Key string
	Val string
}

type testExpectedOutput struct {
	Val interface{}
	Err error
}
