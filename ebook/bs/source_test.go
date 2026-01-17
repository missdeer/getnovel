package bs

import (
	"fmt"
	"testing"
)

func TestReadBookSourceFromLocalFileSystem(t *testing.T) {
	// This test requires a local book source file
	t.Skip("No local book source file configured - skipping test")
}

func TestReadBookSourceFromURL(t *testing.T) {
	setupBookSources()
	if allBookSources.Length() == 0 {
		t.Skip("No book sources loaded - skipping integration test")
	}
	for _, b := range allBookSources.BookSourceCollection {
		fmt.Println(b.BookSourceGroup, b.BookSourceName, b.BookSourceURL, b.Enable)
	}
}

func TestSearchBooks(t *testing.T) {
	setupBookSources()
	if allBookSources.Length() == 0 {
		t.Skip("No book sources loaded - skipping integration test")
	}
	sr := SearchBooks(`斗破苍穹`)
	if sr == nil {
		t.Error("can't find 斗破苍穹")
	}
	fmt.Println(sr)
}