package bs

import (
	"fmt"
	"testing"
)

func TestReadBookSourceFromLocalFileSystem(t *testing.T) {

}

func TestReadBookSourceFromURL(t *testing.T) {
	setupBookSources()
	for _, b := range allBookSources.BookSourceCollection {
		fmt.Println(b.BookSourceGroup, b.BookSourceName, b.BookSourceURL, b.Enable)
	}
	if allBookSources.Length() == 0 {
		t.Error("no book sources read")
	}
}

func TestSearchBooks(t *testing.T) {
	setupBookSources()
	sr:= SearchBooks(`斗破苍穹`)
	if sr == nil {
		t.Error("can't find 斗破苍穹")
	}
	fmt.Println(sr)
}