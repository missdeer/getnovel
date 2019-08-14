package bs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/missdeer/golib/httputil"
	"github.com/patrickmn/go-cache"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var BSCache *cache.Cache = cache.New(0, 0)

type SearchOutput map[string][]*Book

func SortSearchOutput(so SearchOutput) []string {
	sortedResult := make(map[string]int, len(so))
	// var keys = make([]int, len(so))
	var newKeys = make([]string, len(so))
	// var result = &SearchOutput{}
	for k, v := range so {
		sortedResult[k] = len(v)
	}
	type kv struct {
		Key   string
		Value int
	}
	var ss []kv
	for k, v := range sortedResult {
		ss = append(ss, kv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})
	for _, kv := range ss {
		// fmt.Printf("%s, %d\n", kv.Key, kv.Value)
		newKeys = append(newKeys, kv.Key)
	}
	return newKeys
}

type BookSource struct {
	BookSourceGroup       string `json:"bookSourceGroup"`
	BookSourceName        string `json:"bookSourceName"`
	BookSourceURL         string `json:"bookSourceUrl"`
	CheckURL              string `json:"checkUrl"`
	Enable                bool   `json:"enable"`
	HTTPUserAgent         string `json:"httpUserAgent"`
	RuleBookAuthor        string `json:"ruleBookAuthor"`
	RuleBookContent       string `json:"ruleBookContent"`
	RuleBookName          string `json:"ruleBookName"`
	RuleChapterList       string `json:"ruleChapterList"`
	RuleChapterName       string `json:"ruleChapterName"`
	RuleChapterURL        string `json:"ruleChapterUrl"`
	RuleChapterURLNext    string `json:"ruleChapterUrlNext"`
	RuleContentURL        string `json:"ruleContentUrl"`
	RuleCoverURL          string `json:"ruleCoverUrl"`
	RuleFindURL           string `json:"ruleFindUrl"`
	RuleIntroduce         string `json:"ruleIntroduce"`
	RuleSearchAuthor      string `json:"ruleSearchAuthor"`
	RuleSearchCoverURL    string `json:"ruleSearchCoverUrl"`
	RuleSearchKind        string `json:"ruleSearchKind"`
	RuleSearchLastChapter string `json:"ruleSearchLastChapter"`
	RuleSearchList        string `json:"ruleSearchList"`
	RuleSearchName        string `json:"ruleSearchName"`
	RuleSearchNoteURL     string `json:"ruleSearchNoteUrl"`
	RuleSearchURL         string `json:"ruleSearchUrl"`
	SerialNumber          int    `json:"serialNumber"`
	Weight                int    `json:"weight"`
}

func (bs BookSource) String() string {
	return fmt.Sprintf("%s( %s )", bs.BookSourceName, bs.BookSourceURL)
}

type SearchResult struct {
	BookSourceSite string `json:"source"`
	BookTitle      string `json:"name"`
	Author         string `json:"author"`
	BookURL        string `json:"book_url"`
	CoverURL       string `json:"cover_url"`
	Kind           string `json:"kind"`
	LastChapter    string `json:"last_chapter"`
	NoteURL        string `json:"note_url"`
}

/*
例:http://www.gxwztv.com/search.htm?keyword=searchKey&pn=searchPage-1
- ?为get @为post
- searchKey为关键字标识,运行时会替换为搜索关键字,
- searchPage,searchPage-1为搜索页数,从0开始的用searchPage-1,
- page规则还可以写成
{index（第一页）,
indexSecond（第二页）,
indexThird（第三页）,
index-searchPage+1 或 index-searchPage-1 或 index-searchPage}
- 要添加转码编码在最后加 |char=gbk
- |char=escape 会模拟js escape方法进行编码
如果搜索结果可能会跳到简介页请填写简介页url正则
*/
func (b *BookSource) SearchBook(title string) []*Book {
	if b.RuleSearchURL == "" || b.RuleSearchURL == "-" {
		return nil
	}
	searchUrl := b.RuleSearchURL

	// Process encoding transform
	if strings.Contains(searchUrl, "|char") {
		charParam := strings.Split(searchUrl, "|")[1]
		searchUrl = strings.Replace(searchUrl, fmt.Sprintf("|%s", charParam), "", -1)
		charEncoding := strings.Split(charParam, "=")[1]
		charEncoding = strings.ToLower(charEncoding)
		if charEncoding == "gbk" || charEncoding == "gb2312" || charEncoding == "gb18030" {
			data, _ := ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(title)), simplifiedchinese.GBK.NewEncoder()))
			title = string(data)
		}
	}

	var err error
	var p io.Reader
	searchUrl = strings.Replace(searchUrl, "=searchKey", fmt.Sprintf("=%s", url.QueryEscape(title)), -1)
	searchUrl = strings.Replace(searchUrl, "searchPage-1", "0", -1)
	searchUrl = strings.Replace(searchUrl, "searchPage", "1", -1)
	// if searchUrl contains "@", searchKey should be post, not get.
	if b.searchMethod() == "post" {
		data := strings.Split(searchUrl, "@")[1]
		params := strings.Replace(data, "=searchKey", fmt.Sprintf("=%s", url.QueryEscape(title)), -1)
		p, err = httputil.PostPage(strings.Split(searchUrl, "@")[0], params)
	} else {
		log.Println(searchUrl)
		p, err = httputil.GetPage(searchUrl, b.HTTPUserAgent)
	}

	if err != nil {
		log.Printf("searching book error:%s\n", err.Error())
		return nil
	}
	doc, err := goquery.NewDocumentFromReader(p)
	return b.extractSearchResult(doc)

}

func (b *BookSource) searchMethod() string {
	if strings.Contains(b.RuleSearchURL, "@") {
		return "post"
	}
	return "get"
}

func (b *BookSource) searchPage() int {
	if !strings.Contains(b.RuleSearchURL, "searchPage") {
		return -1
	}
	if strings.Contains(b.RuleSearchURL, "searchPage-1") {
		return 0
	}
	return 1
}

func (b *BookSource) extractSearchResult(doc *goquery.Document) []*Book {
	var srList []*Book
	sel, str := ParseRules(doc, b.RuleSearchList)
	if sel != nil {
		sel.Each(func(i int, s *goquery.Selection) {
			_, title := ParseRules(s, b.RuleSearchName)
			if title != "" {
				_, url := ParseRules(s, b.RuleSearchNoteURL)
				_, author := ParseRules(s, b.RuleSearchAuthor)
				_, kind := ParseRules(s, b.RuleSearchKind)
				_, cover := ParseRules(s, b.RuleSearchCoverURL)
				_, lastChapter := ParseRules(s, b.RuleSearchLastChapter)
				_, noteURL := ParseRules(s, b.RuleSearchNoteURL)
				if strings.HasPrefix(url, "/") {
					url = fmt.Sprintf("%s%s", b.BookSourceURL, url)
				}
				if strings.HasPrefix(cover, "/") {
					cover = fmt.Sprintf("%s%s", b.BookSourceURL, cover)
				}
				if strings.HasPrefix(noteURL, "/") {
					noteURL = fmt.Sprintf("%s%s", b.BookSourceURL, noteURL)
				}
				sr := &Book{
					Tag:         b.BookSourceURL,
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

func SearchBooks(title string) SearchOutput {
	c := make(chan *Book, 10)
	result := make(SearchOutput)
	go func() {
		for i, _ := range BSCache.Items() {
			if b, ok := BSCache.Get(i); ok {
				bs, ok := b.(BookSource)
				if ok {
					searchResult := bs.SearchBook(title)
					if searchResult != nil {
						for _, sr := range searchResult {
							c <- sr
						}
					}
				} else {
					log.Println("not book source.")
				}
			}
		}
		close(c)
	}()

	for i := range c {
		if _, ok := result[i.Name]; !ok {
			result[i.Name] = []*Book{i}
			// result[title] = append(result[title], sr)
		} else {
			// fmt.Println("exists, append.")
			result[i.Name] = append(result[i.Name], i)
		}
	}
	for _, key := range SortSearchOutput(result) {
		if key != "" {
			resultJson, _ := json.MarshalIndent(result[key], "", "    ")
			log.Printf("%s:\n %s\n", key, resultJson)
		}
	}
	return result
}
