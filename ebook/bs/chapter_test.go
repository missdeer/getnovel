package bs

import (
	"fmt"
	"testing"
)

func TestChapter(t *testing.T) {
	c := Chapter{}
	c.FromURL("https://www.zwdu.com/book/32642/16771698.html")
	fmt.Println(c.GetContent())
}
