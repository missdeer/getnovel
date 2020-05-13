package bs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/missdeer/golib/httputil"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// Book Sources:
// https://github.com/idalin/govel/raw/master/models/bs_ok.json
// https://moonbegonia.github.io/Source/
// https://github.com/XIU2/yuedu
// https://gitee.com/vpq/yd
// https://github.com/yeyulingfeng01/yuedu.github.io

var (
	allBookSources BookSources
	bookSourceURLs = []string{
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_176",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_176_1",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_1909tv",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_hy",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_miui",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_qidian",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_tingfree",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/00ksw.net",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/kygso",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/530p",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/9txs",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/biquge5200",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/dmzj",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/jpxs123",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/midu",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/motie",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/quanben5",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/shushu8",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/wenku8",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/xiaobiquge",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/xiaoshuo2016",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/yomou.syosetu.com",
		"https://blackholep.github.io/20190815set1",
		"https://blackholep.github.io/31xsw",
		"https://blackholep.github.io/37shuwu",
		"https://blackholep.github.io/58xsw",
		"https://blackholep.github.io/abcxs",
		"https://blackholep.github.io/abcxsw",
		"https://blackholep.github.io/ayg",
		"https://blackholep.github.io/bjzww",
		"https://blackholep.github.io/bqgb5200",
		"https://blackholep.github.io/bqgbiqubao",
		"https://blackholep.github.io/bqgbiquge",
		"https://blackholep.github.io/bqgbiquwu",
		"https://blackholep.github.io/bqgbqg5",
		"https://blackholep.github.io/bqgibiquge",
		"https://blackholep.github.io/bqgkuxiaoshuo",
		"https://blackholep.github.io/bqgwqge",
		"https://blackholep.github.io/ddxs208xs",
		"https://blackholep.github.io/dyddu1du",
		"https://blackholep.github.io/dydduyidu",
		"https://blackholep.github.io/fqxs",
		"https://blackholep.github.io/gsw",
		"https://blackholep.github.io/hysy",
		"https://blackholep.github.io/mhtxsw",
		"https://blackholep.github.io/psw",
		"https://blackholep.github.io/shlwxw",
		"https://blackholep.github.io/slk",
		"https://blackholep.github.io/uxs",
		"https://blackholep.github.io/wcxsw",
		"https://blackholep.github.io/wlzww",
		"https://blackholep.github.io/wxm",
		"https://blackholep.github.io/xbqgxbaquge",
		"https://blackholep.github.io/xbqgxbiquge6",
		"https://blackholep.github.io/xbyzww",
		"https://blackholep.github.io/xsz",
		"https://blackholep.github.io/xszww",
		"https://blackholep.github.io/ybzw",
		"https://blackholep.github.io/ylgxs",
		"https://blackholep.github.io/ymx",
		"https://blackholep.github.io/yssm",
		"https://blackholep.github.io/ywxs",
		"https://blackholep.github.io/zsw",
		"https://booksources.github.io/",
		"https://booksources.github.io/list/biqudao_com.json",
		"https://booksources.github.io/list/cn3k5_com.json",
		"https://booksources.github.io/list/gzmeal_com.json",
		"https://booksources.github.io/list/novel101_com.json",
		"https://booksources.github.io/list/qinxiaoshuo.com.json",
		"https://booksources.github.io/list/qxs.la.json",
		"https://booksources.github.io/list/x23qb_com.json",
		"https://booksources.github.io/list/x23us.com.json",
		"https://booksources.github.io/list/xiashutxt_com.json",
		"https://booksources.github.io/list/xslou_com.json",
		"https://booksources.github.io/list/zhaishuyuan_com.json",
		"http://cloud.iszoc.com/booksource/booksource.json",
		// "http://cloud.iszoc.com/booksource/booksources.json",
	}
)

// ReadBookSourceFromLocalFileSystem book source is stored in local file, read and parse it
func ReadBookSourceFromLocalFileSystem(fileName string) (bs []BookSource) {
	c, e := ioutil.ReadFile(fileName)

	if e != nil {
		log.Println(e)
		return
	}
	e = json.Unmarshal(c, &bs)
	if e == nil {
		return
	}
	var s BookSource
	e2 := json.Unmarshal(c, &s)
	if e2 != nil {
		log.Println(e, e2)
		return
	}
	bs = append(bs, s)
	for _, b := range bs {
		allBookSources.Add(&b)
	}
	return
}

// ReadBookSourceFromURL book source is stored in a URL, read and parse it
func ReadBookSourceFromURL(u string) (bs []BookSource) {
	c, e := httputil.GetBytes(u,
		http.Header{"User-Agent": []string{"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0"}},
		60*time.Second,
		3)
	if e != nil {
		log.Println(e)
		return
	}
	e = json.Unmarshal(c, &bs)
	if e == nil {
		return
	}
	var s BookSource
	e2 := json.Unmarshal(c, &s)
	if e2 != nil {
		log.Println(e, e2)
		return
	}
	bs = append(bs, s)
	for i := range bs {
		if strings.HasSuffix(bs[i].BookSourceURL, `-By Dark`) {
			bs[i].BookSourceURL = bs[i].BookSourceURL[:len(bs[i].BookSourceURL)-len(`-By Dark`)]
		}
		allBookSources.Add(&bs[i])
	}
	return
}

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

// BookSource book source structure
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

// SearchResult book search result
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
func (bs *BookSource) SearchBook(title string) []*Book {
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

func (bs *BookSource) searchMethod() string {
	if strings.Contains(bs.RuleSearchURL, "@") {
		return "post"
	}
	return "get"
}

func (bs *BookSource) searchPage() int {
	if !strings.Contains(bs.RuleSearchURL, "searchPage") {
		return -1
	}
	if strings.Contains(bs.RuleSearchURL, "searchPage-1") {
		return 0
	}
	return 1
}

func (bs *BookSource) extractSearchResult(doc *goquery.Document) []*Book {
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

// SearchBooks search book from book sources
func SearchBooks(title string) SearchOutput {
	c := make(chan *Book, 5)
	result := make(SearchOutput)

	go func() {
		allBookSources.RLock()
		defer allBookSources.RUnlock()
		for _, bs := range allBookSources.BookSourceCollection {
			searchResult := bs.SearchBook(title)
			if searchResult != nil {
				for _, sr := range searchResult {
					c <- sr
				}
			}
		}
	}()

	for timeout := false; !timeout; {
		select {
		case i, ok := <-c:
			if ok {
				if _, ok = result[i.Name]; !ok {
					result[i.Name] = []*Book{i}
				} else {
					result[i.Name] = append(result[i.Name], i)
				}
			}
		case <-time.After(5 * time.Second):
			log.Printf("Timeout,exiting...\n")
			timeout = true
		}
	}

	for _, key := range SortSearchOutput(result) {
		if key != "" {
			resultJSON, _ := json.MarshalIndent(result[key], "", "    ")
			log.Printf("%s:\n %s\n", key, resultJSON)
		}
	}
	return result
}
