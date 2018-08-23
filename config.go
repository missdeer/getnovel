package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/dfordsoft/golib/ebook"
)

// NovelSiteConfig defines novel site configuration information
type NovelSiteConfig struct {
	Sites []struct {
		Host          string `json:"host"`
		Name          string `json:"name"`
		TOCURLPattern string `json:"tocURLPattern"`
	} `json:"sites"`
	BookTitlePattern  string `json:"bookTitlePattern"`
	BookTitlePos      int    `json:"bookTitlePos"`
	ArticlePattern    string `json:"articlePattern"`
	ArticleTitlePos   int    `json:"articleTitlePos"`
	ArticleURLPos     int    `json:"articleURLPos"`
	IsAbsoluteURL     bool   `json:"isAbsoluteURL"`
	Encoding          string `json:"encoding"`
	TOCStyle          string `json:"tocStyle"`
	UserAgent         string `json:"userAgent"`
	PageContentMarker struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"pageContentMarker"`
}

var (
	novelSiteConfigurations []NovelSiteConfig
)

func readNovelSiteConfigurations() {
	matches, err := filepath.Glob("config/*.cfg")
	if err != nil {
		panic(err)
	}

	for _, configFile := range matches {
		contentFd, err := os.OpenFile(configFile, os.O_RDONLY, 0644)
		if err != nil {
			log.Println("opening config file ", configFile, " for reading failed ", err)
			continue
		}

		contentC, err := ioutil.ReadAll(contentFd)
		contentFd.Close()
		if err != nil {
			log.Println("reading config file ", configFile, " failed ", err)
			continue
		}

		config := []NovelSiteConfig{}
		if err = json.Unmarshal(contentC, config); err != nil {
			log.Println("unmarshall configurations failed", err)
			continue
		}

		novelSiteConfigurations = append(novelSiteConfigurations, config...)
	}
}

// Download download book content from novelURL and generate a ebook via gen
func (nsc *NovelSiteConfig) Download(novelURL string, gen ebook.IBook) {

}
