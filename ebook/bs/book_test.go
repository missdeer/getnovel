package bs

import (
	"fmt"
	"log"
	"testing"
)

func TestBook(t *testing.T) {
	allBookSources.Clear()
	for _, u := range bookSourceURLs {
		bss := ReadBookSourceFromURL(u)
		log.Println("Got", len(bss), "book sources from", u)
		// for _, bs := range bss {
		// 	log.Println(bs.BookSourceGroup, "Book source", bs.BookSourceName, "at", bs.BookSourceURL)
		// }
	}

	log.Println("Got", allBookSources.Length(), "book sources totally")
	book, err := NewBookFromURL("http://www.b5200.net/46_46254/")
	if err != nil {
		t.Error(err)
	}
	if book == nil {
		t.Error("no matched book source")
	}
	fmt.Println("===========Book Start===========")
	chapters := book.GetChapterList()
	fmt.Printf("chapter list: %v\n", chapters)
	fmt.Printf("chapter 1 content: %s\n", chapters[0].GetContent())
	fmt.Printf("name: %v\n", book.GetName())
	fmt.Printf("introduce: %v\n", book.GetIntroduce())
	fmt.Printf("author: %v\n", book.GetAuthor())
	fmt.Println("===========Book End=============")

	if book.BookSourceInst == nil {
		t.Error("no matched book source")
	}
	fmt.Println("found book source:", *book.BookSourceInst)
}
