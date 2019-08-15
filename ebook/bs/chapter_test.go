package bs

import (
	"fmt"
	"testing"
)

func TestChapter(t *testing.T) {
	allBookSources.Clear()
	for _, u := range bookSourceURLs {
		ReadBookSourceFromURL(u)
	}
	c, e := NewChapterFromURL("https://www.biquge.cm/9/9434/7424413.html")
	if e != nil {
		t.Error(e)
		return
	}
	fmt.Println(c.GetContent())
}
