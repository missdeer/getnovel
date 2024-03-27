package config

import (
	"net/http"

	"github.com/missdeer/getnovel/ebook"
)

type TOCPattern struct {
	Host            string
	BookTitle       string
	BookTitlePos    int
	Item            string
	ArticleTitlePos int
	ArticleURLPos   int
	IsAbsoluteURL   bool
}

type PageContentMarker struct {
	Host  string
	Start []byte
	End   []byte
}

type NovelChapterInfo struct {
	Index int
	Title string
	URL   string
}

type NovelSite struct {
	Title string
	Urls  []string
}

type NovelSiteHandler struct {
	Sites                    []NovelSite
	CanHandle                func(string) bool                                  // (url) -> can handle
	PreprocessChapterListURL func(string) string                                // (original url) -> final url
	ExtractChapterList       func(string, []byte) (string, []*NovelChapterInfo) // (url, raw page content) (title, chapters)
	ExtractChapterContent    func(string, []byte) []byte                        // (raw page content) -> cleanup content
	PreprocessContentLink    func(string) (string, http.Header)                 // (url) -> (final url, headers)
	Download                 func(string, ebook.IBook)
	Begin                    func()
	End                      func()
}
