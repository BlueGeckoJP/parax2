package main

import (
	"flag"
	"testing"
)

var rootFlag = flag.String("root", ".", "Root directory to search")

func BenchmarkSearch(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		search(*rootFlag, 2)
	}
}
