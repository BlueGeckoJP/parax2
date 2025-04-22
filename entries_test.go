package main

import (
	"flag"
	"testing"
)

var rootFlag = flag.String("root", ".", "Root directory to search")

func BenchmarkSearch(b *testing.B) {
	for b.Loop() {
		search(*rootFlag, 2)
	}
}
