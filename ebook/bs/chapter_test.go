package bs

import (
	"fmt"
	"testing"
)

func TestChapter(t *testing.T) {
	c, e := NewBookFromURL("https://www.zwdu.com/book/32642/16771698.html")
	if e != nil {
		t.Error(e)
		return
	}
	fmt.Println(c.GetContent())
}
