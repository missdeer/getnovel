package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/dfordsoft/golib/ebook"
	"github.com/dfordsoft/golib/httputil"
	"github.com/dfordsoft/golib/ic"
)

// NovelSiteConfig defines novel site configuration information
type NovelSiteConfig struct {
	Sites []struct {
		Host          string `json:"host"`
		Name          string `json:"name"`
		TOCURLPattern string `json:"tocURLPattern"`
	} `json:"sites"`
	Title              string `json:"title"`
	BookTitlePattern   string `json:"bookTitlePattern"`
	BookTitlePos       int    `json:"bookTitlePos"`
	ArticlePattern     string `json:"articlePattern"`
	ArticleTitlePos    int    `json:"articleTitlePos"`
	ArticleURLPos      int    `json:"articleURLPos"`
	IsAbsoluteURL      bool   `json:"isAbsoluteURL"`
	Encoding           string `json:"encoding"`
	TOCStyle           string `json:"tocStyle"`
	UserAgent          string `json:"userAgent"`
	PageContentMarkers []struct {
		Host  string `json:"host"`
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"pageContentMarkers"`
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

// Download download book content from novel URL and generate a ebook
func (nsc *NovelSiteConfig) Download(u string, gen ebook.IBook) {
	theURL, _ := url.Parse(u)
	headers := http.Header{
		"Referer":                   []string{fmt.Sprintf("%s://%s", theURL.Scheme, theURL.Host)},
		"User-Agent":                []string{nsc.UserAgent},
		"Accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"},
		"Accept-Language":           []string{`en-US,en;q=0.8`},
		"Upgrade-Insecure-Requests": []string{"1"},
	}

	dlPage := func(u string) (c []byte) {
		var err error
		theURL, _ := url.Parse(u)
		c, err = httputil.GetBytes(u, headers, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
		if err != nil {
			return
		}

		if bytes.Index(c, []byte("charset="+nsc.Encoding)) > 0 {
			c = ic.Convert(nsc.Encoding, "utf-8", c)
		}
		c = bytes.Replace(c, []byte("\r\n"), []byte(""), -1)
		c = bytes.Replace(c, []byte("\r"), []byte(""), -1)
		c = bytes.Replace(c, []byte("\n"), []byte(""), -1)
		for _, m := range nsc.PageContentMarkers {
			if theURL.Host == m.Host {
				idx := bytes.Index(c, []byte(m.Start))
				if idx > 1 {
					//fmt.Println("found start")
					c = c[idx+len(m.Start):]
				}
				idx = bytes.Index(c, []byte(m.End))
				if idx > 1 {
					//fmt.Println("found end")
					c = c[:idx]
				}
				break
			}
		}

		c = bytes.Replace(c, []byte("<br /><br />&nbsp;&nbsp;&nbsp;&nbsp;"), []byte("</p><p>"), -1)
		c = bytes.Replace(c, []byte("<br />&nbsp;&nbsp;&nbsp;&nbsp;"), []byte("</p><p>"), -1)
		c = bytes.Replace(c, []byte("<br/><br/>"), []byte("</p><p>"), -1)
		c = bytes.Replace(c, []byte(`　　`), []byte(""), -1)
		return
	}

	b, err := httputil.GetBytes(u, headers, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
	if err != nil {
		return
	}

	b = bytes.Replace(b, []byte("<dd>"), []byte("\n<dd>"), -1)
	b = bytes.Replace(b, []byte("</dd>"), []byte("</dd>\n"), -1)

	if bytes.Index(b, []byte("charset="+nsc.Encoding)) > 0 {
		b = ic.Convert(nsc.Encoding, "utf-8", b)
	}
	gen.Begin()

	dlutil := newDownloadUtil(dlPage, gen)
	dlutil.process()

	var title string
	var lines []string

	r, _ := regexp.Compile(nsc.ArticlePattern)
	re, _ := regexp.Compile(nsc.BookTitlePattern)
	scanner := bufio.NewScanner(bytes.NewReader(b))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if title == "" {
			ss := re.FindAllStringSubmatch(line, -1)
			if len(ss) > 0 && len(ss[0]) > 0 {
				s := ss[0]
				title = s[nsc.BookTitlePos]
				gen.SetTitle(title)
				continue
			}
		}
		if r.MatchString(line) {
			lines = append(lines, line)
		}
	}

	for i := len(lines) - 1; i >= 0 && i < len(lines) && lines[0] == lines[i]; i -= 2 {
		lines = lines[1:]
	}

	for index, line := range lines {
		ss := r.FindAllStringSubmatch(line, -1)
		s := ss[0]
		articleURL := s[nsc.ArticleURLPos]
		finalURL := fmt.Sprintf("%s://%s%s", theURL.Scheme, theURL.Host, articleURL)
		if articleURL[0] != '/' {
			finalURL = fmt.Sprintf("%s%s", u, articleURL)
		}
		if strings.HasPrefix(articleURL, "http") {
			finalURL = articleURL
		}

		if dlutil.addURL(index+1, s[nsc.ArticleTitlePos], finalURL) {
			break
		}
	}
	dlutil.wait()
	gen.End()
}
