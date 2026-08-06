package parse

import (
	"sync"
	"testing"
)

// FuzzParse checks that malformed input never panics or hangs the parser and
// that extracted symbols always carry valid ranges.
func FuzzParse(f *testing.F) {
	paths := []string{"a.go", "a.ts", "a.js", "a.py", "a.rs", "a.java", "a.c", "a.cpp"}
	seeds := []string{
		"",
		"package main\nfunc main() {}",
		"class A { }",
		"\x00\x01\x02",
	}
	for _, path := range paths {
		for _, seed := range seeds {
			f.Add(path, seed)
		}
	}
	f.Fuzz(func(t *testing.T, path string, source string) {
		lang, ok := Detect(path)
		if !ok {
			t.Skip()
		}
		result := lang.Parse([]byte(source))
		for _, symbol := range result.Symbols {
			if symbol.EndByte <= symbol.StartByte {
				t.Errorf("empty range for %s: %d-%d", symbol.Name, symbol.StartByte, symbol.EndByte)
			}
			if symbol.StartLine == 0 || symbol.EndLine < symbol.StartLine {
				t.Errorf("invalid lines for %s: %d-%d", symbol.Name, symbol.StartLine, symbol.EndLine)
			}
		}
	})
}

// TestConcurrentExtraction verifies that one compiled query can be shared by
// parsers running concurrently without corruption or data races.
func TestConcurrentExtraction(t *testing.T) {
	lang, ok := Detect("main.go")
	if !ok {
		t.Fatal("no go language")
	}
	code := []byte("package main\n\nfunc Foo() {}\nfunc Bar() {}\ntype T struct{}\n")
	var wg sync.WaitGroup
	for range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := lang.Parse(code)
			if len(result.Symbols) != 3 {
				t.Errorf("got %d symbols, want 3", len(result.Symbols))
			}
		}()
	}
	wg.Wait()
}
