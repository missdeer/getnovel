package main

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
)

type novelSiteHandler struct {
	Title         string
	MatchPatterns []string
	Download      func(string)
}

var (
	novelSiteHandlers []*novelSiteHandler
)

func registerNovelSiteHandler(h *novelSiteHandler) {
	novelSiteHandlers = append(novelSiteHandlers, h)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: getnovel novel-toc-url")
		return
	}

	if os.Args[1] == "list" {
		fmt.Println("支持小说网站：")
		for _, h := range novelSiteHandlers {
			fmt.Println(h.Title)
		}
		return
	}

	_, e := url.Parse(os.Args[1])
	if e != nil {
		fmt.Println("invalid novel url input")
		return
	}

	for _, h := range novelSiteHandlers {
		for _, pattern := range h.MatchPatterns {
			r, _ := regexp.Compile(pattern)
			if r.MatchString(os.Args[1]) {
				h.Download(os.Args[1])
				return
			}
		}
	}
	fmt.Println("Usage: getnovel novel-toc-url")
}
