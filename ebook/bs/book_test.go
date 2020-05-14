package bs

import (
	"fmt"
	"log"
	"testing"
)

func setupBookSources() {
	if allBookSources.Length() != 0 {
		log.Println("Already have", allBookSources.Length(), "book sources totally")
		return
	}
	for _, u := range bookSourceURLs {
		bss := ReadBookSourceFromURL(u)
		log.Println("Got", len(bss), "book sources from", u)
		// for _, bs := range bss {
		// 	log.Println(bs.BookSourceGroup, "Book source", bs.BookSourceName, "at", bs.BookSourceURL)
		// }
	}
	log.Println("Got", allBookSources.Length(), "book sources totally")
}

func TestBook(t *testing.T) {
	setupBookSources()
	book, err := NewBookFromURL("http://www.b5200.net/46_46254/")
	if err != nil {
		t.Error(err)
	}
	if book == nil {
		t.Error("no matched book source")
	}
	fmt.Println("===========Book Start===========")
	chapters := book.GetChapterList()
	fmt.Printf("Got %d chapters\n", len(chapters))
	fmt.Printf("Chapter 1 url: %s\n", chapters[0].ChapterURL)
	fmt.Printf("Chapter 1 title: %s\n", chapters[0].GetTitle())
	fmt.Printf("Chapter 1 content: %s\n", chapters[0].GetContent())
	fmt.Printf("Book name: %v\n", book.GetName())
	fmt.Printf("Book introduce: %v\n", book.GetIntroduce())
	fmt.Printf("Book author: %v\n", book.GetAuthor())
	if book.BookSourceInst == nil {
		t.Error("no matched book source")
	}
	fmt.Println("Found book source:", *book.BookSourceInst)
	fmt.Println("===========Book End=============")
}
