package main

import (
	"testing"
)

func TestGenGo(t *testing.T) {
	testGenEqualGolden(t, GenGo, "testdata/go.golden")
}
