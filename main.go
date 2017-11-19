package main

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
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

func listCommandHandler() {
	fmt.Println("支持小说网站：")
	for _, h := range novelSiteHandlers {
		urlMap := make(map[string]struct{})
		for _, p := range h.MatchPatterns {
			u := strings.Replace(p, `\`, ``, -1)
			idxStart := strings.Index(u, `www.`)
			idxEnd := strings.Index(u[idxStart:], `/`)
			u = u[:idxStart+idxEnd]
			urlMap[u] = struct{}{}
		}
		var urls []string
		for u := range urlMap {
			urls = append(urls, u)
		}
		fmt.Println(h.Title + ": " + strings.Join(urls, ", "))
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: getnovel novel-toc-url")
		return
	}

	if os.Args[1] == "list" {
		listCommandHandler()
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
