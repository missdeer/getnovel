package bs

import (
	"fmt"
	"testing"
)

func setupBookSources() {
	// Book sources are now loaded from command line options or config
	// For tests, we check if any sources are already loaded
	if allBookSources.Length() == 0 {
		fmt.Println("No book sources loaded. Use --bookSourceURL, --bookSourceDir, or --bookSourceFile to load sources.")
	}
}

func TestBook(t *testing.T) {
	setupBookSources()
	if allBookSources.Length() == 0 {
		t.Skip("No book sources loaded - skipping integration test")
	}
	book, err := NewBookFromURL("https://www.mangg.net/id68990/")
	if err != nil {
		t.Error(err)
		return
	}
	if book == nil {
		t.Error("no matched book source")
		return
	}
	fmt.Println("===========Book Start===========")
	chapters := book.GetChapterList()
	if len(chapters) == 0 {
		t.Error("no chapters found")
		return
	}
	fmt.Printf("Got %d chapters\n", len(chapters))
	fmt.Printf("Chapter 1 url: %s\n", chapters[0].ChapterURL)
	fmt.Printf("Chapter 1 title: %s\n", chapters[0].GetTitle())
	fmt.Printf("Chapter 1 content: %s\n", chapters[0].GetContent())
	fmt.Printf("Book name: %v\n", book.GetName())
	fmt.Printf("Book introduce: %v\n", book.GetIntroduce())
	fmt.Printf("Book author: %v\n", book.GetAuthor())
	if book.BookSourceInst == nil {
		t.Error("no matched book source")
		return
	}
	fmt.Println("Found book source:", *book.BookSourceInst)
	fmt.Println("===========Book End=============")
}
