package bs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/missdeer/golib/httputil"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

// BookSourceV2 book source structure
type BookSourceV2 struct {
	BookSourceGroup       string `json:"bookSourceGroup"`
	BookSourceName        string `json:"bookSourceName"`
	BookSourceURL         string `json:"bookSourceUrl"`
	Enable                bool   `json:"enable"`
	HTTPUserAgent         string `json:"httpUserAgent"`
	RuleBookAuthor        string `json:"ruleBookAuthor,omitempty"`
	RuleBookContent       string `json:"ruleBookContent"`
	RuleBookName          string `json:"ruleBookName,omitempty"`
	RuleChapterList       string `json:"ruleChapterList"`
	RuleChapterName       string `json:"ruleChapterName"`
	RuleChapterURL        string `json:"ruleChapterUrl"`
	RuleContentURL        string `json:"ruleContentUrl,omitempty"`
	RuleCoverURL          string `json:"ruleCoverUrl,omitempty"`
	RuleIntroduce         string `json:"ruleIntroduce"`
	RuleSearchAuthor      string `json:"ruleSearchAuthor"`
	RuleSearchCoverURL    string `json:"ruleSearchCoverUrl"`
	RuleSearchKind        string `json:"ruleSearchKind"`
	RuleSearchLastChapter string `json:"ruleSearchLastChapter"`
	RuleSearchList        string `json:"ruleSearchList"`
	RuleSearchName        string `json:"ruleSearchName"`
	RuleSearchNoteURL     string `json:"ruleSearchNoteUrl"`
	RuleSearchURL         string `json:"ruleSearchUrl"`
}

func (bs BookSourceV2) String() string {
	return fmt.Sprintf("%s( %s )", bs.BookSourceName, bs.BookSourceURL)
}

// SearchBook search book on the book source
// 例:http://www.gxwztv.com/search.htm?keyword=searchKey&pn=searchPage-1
// - ?为get @为post
// - searchKey为关键字标识,运行时会替换为搜索关键字,
// - searchPage,searchPage-1为搜索页数,从0开始的用searchPage-1,
// - page规则还可以写成
// {index（第一页）,
// indexSecond（第二页）,
// indexThird（第三页）,
// index-searchPage+1 或 index-searchPage-1 或 index-searchPage}
// - 要添加转码编码在最后加 |char=gbk
// - |char=escape 会模拟js escape方法进行编码
// 如果搜索结果可能会跳到简介页请填写简介页url正则
func (bs *BookSourceV2) SearchBook(title string) []*Book {
	if bs.RuleSearchURL == "" || bs.RuleSearchURL == "-" {
		return nil
	}
	searchURL := bs.RuleSearchURL

	// Process encoding transform
	if strings.Contains(searchURL, "|char") {
		charParam := strings.Split(searchURL, "|")[1]
		searchURL = strings.Replace(searchURL, fmt.Sprintf("|%s", charParam), "", -1)
		charEncoding := strings.Split(charParam, "=")[1]
		charEncoding = strings.ToLower(charEncoding)
		switch charEncoding {
		case "gbk":
			data, _ := ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(title)), simplifiedchinese.GBK.NewEncoder()))
			title = string(data)
		case "gb2312":
			data, _ := ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(title)), simplifiedchinese.HZGB2312.NewEncoder()))
			title = string(data)
		case "gb18030":
			data, _ := ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(title)), simplifiedchinese.GB18030.NewEncoder()))
			title = string(data)
		case "big5", "big-5":
			data, _ := ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(title)), traditionalchinese.Big5.NewEncoder()))
			title = string(data)
		}
	}

	var err error
	var p io.Reader
	searchURL = strings.Replace(searchURL, "=searchKey", fmt.Sprintf("=%s", url.QueryEscape(title)), -1)
	searchURL = strings.Replace(searchURL, "searchPage-1", "0", -1)
	searchURL = strings.Replace(searchURL, "searchPage", "1", -1)
	// if searchUrl contains "@", searchKey should be post, not get.
	if bs.searchMethod() == "post" {
		data := strings.Split(searchURL, "@")[1]
		params := strings.Replace(data, "=searchKey", fmt.Sprintf("=%s", url.QueryEscape(title)), -1)
		p, err = httputil.PostPage(strings.Split(searchURL, "@")[0], params)
	} else {
		log.Println(searchURL)
		p, err = httputil.GetPage(searchURL, bs.HTTPUserAgent)
	}

	if err != nil {
		log.Printf("searching book error:%s\n", err.Error())
		return nil
	}
	doc, err := goquery.NewDocumentFromReader(p)
	if err != nil {
		log.Printf("searching book error:%s\n", err.Error())
		return nil
	}
	if doc == nil {
		log.Printf("doc is nil.")
		return nil
	}
	return bs.extractSearchResult(doc)
}

func (bs *BookSourceV2) searchMethod() string {
	if strings.Contains(bs.RuleSearchURL, "@") {
		return "post"
	}
	return "get"
}

func (bs *BookSourceV2) searchPage() int {
	if !strings.Contains(bs.RuleSearchURL, "searchPage") {
		return -1
	}
	if strings.Contains(bs.RuleSearchURL, "searchPage-1") {
		return 0
	}
	return 1
}

func (bs *BookSourceV2) extractSearchResult(doc *goquery.Document) []*Book {
	var srList []*Book
	sel, str := ParseRules(doc, bs.RuleSearchList)
	if sel != nil {
		sel.Each(func(i int, s *goquery.Selection) {
			_, title := ParseRules(s, bs.RuleSearchName)
			if title != "" {
				_, url := ParseRules(s, bs.RuleSearchNoteURL)
				_, author := ParseRules(s, bs.RuleSearchAuthor)
				_, kind := ParseRules(s, bs.RuleSearchKind)
				_, cover := ParseRules(s, bs.RuleSearchCoverURL)
				_, lastChapter := ParseRules(s, bs.RuleSearchLastChapter)
				_, noteURL := ParseRules(s, bs.RuleSearchNoteURL)
				if strings.HasPrefix(url, "/") {
					url = fmt.Sprintf("%s%s", bs.BookSourceURL, url)
				}
				if strings.HasPrefix(cover, "/") {
					cover = fmt.Sprintf("%s%s", bs.BookSourceURL, cover)
				}
				if strings.HasPrefix(noteURL, "/") {
					noteURL = fmt.Sprintf("%s%s", bs.BookSourceURL, noteURL)
				}
				sr := &Book{
					Tag:         bs.BookSourceURL,
					Name:        title,
					Author:      author,
					Kind:        kind,
					CoverURL:    cover,
					LastChapter: lastChapter,
					NoteURL:     noteURL,
				}
				srList = append(srList, sr)

			}
		})

	} else {
		log.Printf("No search result found. string:%s\n", str)
	}

	return srList
}
