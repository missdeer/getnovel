package main

import (
	"fmt"
	"net/url"
	"os"
)

type Matcher func(string) bool
type Downloader func(string)
type NovelSiteHandler struct {
	Match    Matcher
	Download Downloader
}

var (
	novelSiteHandlers []*NovelSiteHandler
)

func registerNovelSiteHandler(h *NovelSiteHandler) {
	novelSiteHandlers = append(novelSiteHandlers, h)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: getnovel novel-url")
		return
	}
	_, e := url.Parse(os.Args[1])
	if e != nil {
		fmt.Println("invalid novel url input")
		return
	}

	for _, h := range novelSiteHandlers {
		if h.Match(os.Args[1]) {
			h.Download(os.Args[1])
			return
		}
	}
	fmt.Println("Usage: getnovel novel-url")
}
