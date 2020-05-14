package bs

import (
	"log"
	"testing"
)

func TestChapter(t *testing.T) {
	setupBookSources()
	c, e := NewChapterFromURL("http://www.b5200.net/46_46254/17700048.html")
	if e != nil {
		t.Error(e)
		return
	}
	log.Println("chapter title:", c.GetTitle())
	log.Println("chapter content:", c.GetContent())
}
