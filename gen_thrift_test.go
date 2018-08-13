package main

import (
	"testing"
)

func TestGenThrift(t *testing.T) {
	testGenEqualGolden(t, GenThrift, "testdata/thrift.golden")
}
