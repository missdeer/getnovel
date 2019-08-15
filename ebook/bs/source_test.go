package bs

import (
	"fmt"
	"testing"
)

func TestReadBookSourceFromLocalFileSystem(t *testing.T) {

}

func TestReadBookSourceFromURL(t *testing.T) {
	for _, u := range bookSourceURLs {
		bs := ReadBookSourceFromURL(u)
		for _, b := range bs {
			fmt.Println(b.BookSourceGroup, b.BookSourceName, b.BookSourceURL, b.Enable)
		}
	}
}
