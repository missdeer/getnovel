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
		for _, bs := range bss {
			log.Println(bs.BookSourceGroup, "Book source", bs.BookSourceName, "at", bs.BookSourceURL)
		}
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
	fmt.Printf("%v\n", book.GetChapterList())
	fmt.Printf("%v\n", book.GetName())
	fmt.Printf("%v\n", book.GetIntroduce())
	fmt.Printf("%v\n", book.GetAuthor())
	fmt.Println("===========Book End=============")

	bs := book.findBookSourceForBook()
	if bs == nil {
		t.Error("no matched book source")
	}
	fmt.Println(bs.BookSourceGroup, bs.BookSourceName, bs.BookSourceURL)
}
