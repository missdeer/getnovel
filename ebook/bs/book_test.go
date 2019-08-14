package bs

import (
	"fmt"
	"testing"
)

func TestBook(t *testing.T) {
	book := Book{}
	fmt.Println("===========Book Start===========")
	book.FromURL("https://www.zwdu.com/book/32642/")
	fmt.Printf("%v\n", book.GetChapterList())
	fmt.Printf("%v\n", book.GetName())
	fmt.Printf("%v\n", book.GetIntroduce())
	fmt.Printf("%v\n", book.GetAuthor())
	fmt.Println("===========Book End=============")
}
