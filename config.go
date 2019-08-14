package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/missdeer/getnovel/ebook"
	"github.com/missdeer/golib/httputil"
	"github.com/missdeer/golib/ic"
)

// PageProcessor defines page processor, replace From with To
type PageProcessor struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// NovelSiteConfig defines novel site configuration information
type NovelSiteConfig struct {
	Host                  string          `json:"host"`
	SiteName              string          `json:"siteName"`
	BookTitlePattern      string          `json:"bookTitlePattern"`
	ArticleListPattern    string          `json:"articleListPattern"`
	ArticleTitlePattern   string          `json:"articleTitlePattern"`
	ArticleContentPattern string          `json:"articleContentPattern"`
	ArticleURLPattern     string          `json:"articleURLPattern"`
	IsAbsoluteURL         bool            `json:"isAbsoluteURL"`
	Encoding              string          `json:"encoding"`
	TOCStyle              string          `json:"tocStyle"`
	Cookies               string          `json:"cookies"`
	UserAgent             string          `json:"userAgent"`
	PagePreprocessor      []PageProcessor `json:"pagePreprocessor"`
	PagePostprocessor     []PageProcessor `json:"pagePostprocessor"`
}

// ArticleInfo simple article info
type ArticleInfo struct {
	Title string
	URL   string
}

var (
	novelSiteConfigurations []NovelSiteConfig
)

func readNovelSiteConfigurations() {
	matches, err := filepath.Glob("config/*.json")
	if err != nil {
		panic(err)
	}

	for _, configFile := range matches {
		fd, err := os.OpenFile(configFile, os.O_RDONLY, 0644)
		if err != nil {
			log.Println("opening config file ", configFile, " for reading failed ", err)
			continue
		}

		c, err := ioutil.ReadAll(fd)
		fd.Close()
		if err != nil {
			log.Println("reading config file ", configFile, " failed ", err)
			continue
		}

		config := []NovelSiteConfig{}
		if err = json.Unmarshal(c, &config); err != nil {
			log.Println("unmarshall configurations failed", err)
			continue
		}

		novelSiteConfigurations = append(novelSiteConfigurations, config...)
	}
}

func match(doc *goquery.Document, pattern string) (res string) {
	return
}

func matchSelection(doc *goquery.Document, pattern string) *goquery.Selection {
	return nil
}

func matchArray(sel *goquery.Selection, pattern string) (res []string) {
	return
}

// Download download book content from novel URL and generate a ebook
func (nsc *NovelSiteConfig) Download(u string, gen ebook.IBook) {
	theURL, _ := url.Parse(u)
	headers := http.Header{
		"Referer":                   []string{fmt.Sprintf("%s://%s", theURL.Scheme, theURL.Host)},
		"User-Agent":                []string{"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0"},
		"Accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"},
		"Accept-Language":           []string{`en-US,en;q=0.8`},
		"Upgrade-Insecure-Requests": []string{"1"},
	}
	if nsc.UserAgent != "" {
		headers["User-Agent"] = []string{nsc.UserAgent}
	}
	if nsc.Cookies != "" {
		headers["Cookie"] = []string{nsc.Cookies}
	}

	dlPage := func(u string) (c []byte) {
		var err error
		c, err = httputil.GetBytes(u, headers, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
		if err != nil {
			return
		}

		// encoding convert
		if nsc.Encoding != "utf-8" {
			c = ic.Convert(nsc.Encoding, "utf-8", c)
		}

		// preprocess
		for _, m := range nsc.PagePreprocessor {
			c = bytes.Replace(c, []byte(m.From), []byte(m.To), -1)
		}

		// find the main content
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(c))
		if err != nil {
			log.Println(err)
			return
		}
		c = []byte(match(doc, nsc.ArticleContentPattern))

		// post process
		for _, m := range nsc.PagePostprocessor {
			c = bytes.Replace(c, []byte(m.From), []byte(m.To), -1)
		}
		return
	}

	b, err := httputil.GetBytes(u, headers, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
	if err != nil {
		return
	}

	if nsc.Encoding != "utf-8" {
		b = ic.Convert(nsc.Encoding, "utf-8", b)
	}
	gen.Begin()

	dlutil := newDownloadUtil(dlPage, gen)
	dlutil.process()

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(b))
	if err != nil {
		log.Println(err)
		return
	}

	// extract book title
	title := match(doc, nsc.BookTitlePattern)
	gen.SetTitle(title)

	// extract book articles title and URL
	selection := matchSelection(doc, nsc.ArticleListPattern)

	titles := matchArray(selection, nsc.ArticleTitlePattern)
	urls := matchArray(selection, nsc.ArticleURLPattern)

	if len(titles) != len(urls) {
		log.Fatalln("title count and URL count not match", len(titles), len(urls))
		return
	}

	var articles []ArticleInfo
	for i := 0; i < len(titles); i++ {
		article := ArticleInfo{
			Title: titles[i],
			URL:   urls[i],
		}
		articles = append(articles, article)
	}

	// clean & sort articles
	switch nsc.TOCStyle {
	case "from-begin-to-end":
	case "from-end-to-begin":
		for i := len(articles)/2 - 1; i >= 0; i-- {
			opp := len(articles) - 1 - i
			articles[i], articles[opp] = articles[opp], articles[i]
		}
	case "recent-at-begin":
		for i := len(articles) - 1; i >= 0 && i < len(articles) && articles[0].URL == articles[i].URL; i -= 2 {
			articles = articles[1:]
		}
	}

	// download article content
	for index, article := range articles {
		finalURL := article.URL
		if !nsc.IsAbsoluteURL {
			finalURL = fmt.Sprintf("%s://%s%s", theURL.Scheme, theURL.Host, article.URL)
		}
		if dlutil.addURL(index+1, article.Title, finalURL) {
			break
		}
	}

	dlutil.wait()
	gen.End()
}
