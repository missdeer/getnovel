package handler

import (
	"fmt"
	"strings"

	"github.com/missdeer/getnovel/config"
)

var (
	novelSiteHandlers []*config.NovelSiteHandler
)

func ListHandlers() {
	fmt.Println("当前支持小说网站：")
	for _, h := range novelSiteHandlers {
		for _, site := range h.Sites {
			fmt.Println("\t" + site.Title + ": " + strings.Join(site.Urls, ", "))
		}
	}
}

func RunHandler(runHandler func(*config.NovelSiteHandler) bool) bool {
	for _, handler := range novelSiteHandlers {
		if runHandler(handler) {
			return true
		}
	}

	return false
}

func registerNovelSiteHandler(handler *config.NovelSiteHandler) {
	novelSiteHandlers = append(novelSiteHandlers, handler)
}
