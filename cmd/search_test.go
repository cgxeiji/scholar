package cmd

import "testing"

func BenchmarkSearchContains(b *testing.B) {
	Execute()
	for i := 0; i < b.N; i++ {
		guiSearch("test", entryList(), searcher)
	}
}

func BenchmarkSearchFuzzy(b *testing.B) {
	Execute()
	for i := 0; i < b.N; i++ {
		guiSearchFuzzy("test", entryList())
	}
}
