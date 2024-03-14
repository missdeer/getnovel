package handler

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/missdeer/getnovel/config"
	"github.com/missdeer/golib/ic"
)

func preprocess69xinshuChapterListURL(u string) string {
	reg := regexp.MustCompile(`https://www\.69shu\.pro/book/([0-9]+)\.htm`)
	if reg.MatchString(u) {
		ss := reg.FindAllStringSubmatch(u, -1)
		s := ss[0]
		return fmt.Sprintf("https://www.69shu.pro/book/%s/", s[1])
	}
	return u
}

func extract69xinshuChapterList(u string, rawPageContent []byte) (title string, chapters []*config.NovelChapterInfo) {
	c := ic.Convert("gbk", "utf-8", rawPageContent)
	c = bytes.Replace(c, []byte("\r\n"), []byte(""), -1)
	c = bytes.Replace(c, []byte("\r"), []byte(""), -1)
	c = bytes.Replace(c, []byte("\n"), []byte(""), -1)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(c))
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("div.bread a").Each(func(index int, item *goquery.Selection) {
		title = item.Text()
	})
	if strings.HasSuffix(title, `章节列表`) {
		for i := 0; i < 4; i++ {
			_, size := utf8.DecodeLastRuneInString(title)
			title = title[:len(title)-size]
		}
	}
	doc.Find("#catalog li").Each(func(i int, s *goquery.Selection) {
		if a := s.Find("a"); a != nil {
			if href, exists := a.Attr("href"); exists {
				chapters = append(chapters, &config.NovelChapterInfo{
					Index: i + 1,
					Title: a.Text(),
					URL:   href,
				})
			}
		}
	})

	return
}

func extract69xinshuChapterContent(rawPageContent []byte) (c []byte) {
	c = ic.Convert("gbk", "utf-8", rawPageContent)
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(c))
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("h1.hide720").Remove()
	doc.Find("div.txtinfo.hide720").Remove()
	doc.Find("div#txtright").Remove()

	html, err := doc.Find("div.txtnav").Html()
	if err != nil {
		log.Fatal(err)
	}
	c = bytes.Replace([]byte(html), []byte(`&emsp;&emsp;`), []byte("  "), -1)
	c = bytes.Replace(c, []byte("<br /><br />"), []byte("<br/>"), -1)
	c = bytes.Replace(c, []byte("<br/><br/>"), []byte("<br/>"), -1)
	c = bytes.Replace(c, []byte("\xe2\x80\x83"), []byte("&nbsp;"), -1)
	return
}

func init() {
	registerNovelSiteHandler(&config.NovelSiteHandler{
		Sites: []config.NovelSite{
			{
				Title: `69书吧`,
				Urls:  []string{`https://www.69shu.pro/`},
			},
		},
		CanHandle: func(u string) bool {
			patterns := []string{
				`https://www\.69shu\.pro/book/[0-9]+/`,
				`^https://www\.69shu\.pro/book/[0-9]+\.html?$`,
			}
			for _, pattern := range patterns {
				reg := regexp.MustCompile(pattern)
				if reg.MatchString(u) {
					return true
				}
			}
			return false
		},
		PreprocessChapterListURL: preprocess69xinshuChapterListURL,
		ExtractChapterList:       extract69xinshuChapterList,
		ExtractChapterContent:    extract69xinshuChapterContent,
	})
}
