package bs

import (
	"fmt"
	"testing"
)

func TestBook(t *testing.T) {
	allBookSources.Clear()
	for _, u := range bookSourceURLs {
		ReadBookSourceFromURL(u)
	}
	book, err := NewBookFromURL("https://www.biquge.cm/9/9434/")
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
}
