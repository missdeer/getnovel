package bs

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/missdeer/golib/httputil"
)

var (
	allBookSources BookSources
	bookSourceURLs = []string{
		"https://cdn.jsdelivr.net/gh/yeyulingfeng01/yuedu.github.io/202003.txt",
		"https://gitee.com/vpq/codes/ez5qu1ifx260layps3b7981/raw?blob_name=3.0sy.json",
		"https://xiu2.github.io/yuedu/shuyuan",
		"https://moonbegonia.github.io/Source/yuedu/full.json",
		"https://github.com/idalin/govel/raw/master/models/bs_ok.json",
		"http://alanskycn.gitee.io/vip/assets/import/book_source.json",
		"https://gitee.com/haobai1/bookyuan/raw/master/shuyuan.json",
		"https://gitee.com/zmn1307617161/booksource/raw/master/%E4%B9%A6%E6%BA%90/%E7%B2%BE%E6%8E%923.txt",
		"https://gitee.com/slght/yuedu_booksource/raw/master/%E4%B9%A6%E6%BA%90/API%E4%B9%A6%E6%BA%90_3.0.json",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_176",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_176_1",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_1909tv",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_hy",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_miui",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_qidian",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_tingfree",
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
	}
)

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

// ConvertBookSourceV3ToV2 convert BookSource from v3 to v2
func ConvertBookSourceV3ToV2(bs3 BookSourceV3) BookSourceV2 {
	var header map[string]string
	var userAgent string
	if e := json.Unmarshal([]byte(bs3.Header), &header); e == nil {
		userAgent = header["User-Agent"]
	}
	return BookSourceV2{
		BookSourceGroup:       bs3.BookSourceGroup,
		BookSourceName:        bs3.BookSourceName,
		BookSourceURL:         bs3.BookSourceURL,
		Enable:                bs3.Enable,
		HTTPUserAgent:         userAgent,
		RuleBookAuthor:        bs3.RuleBookInfo.Author,
		RuleBookContent:       bs3.RuleContent.Content,
		RuleBookName:          bs3.RuleBookInfo.Name,
		RuleChapterList:       bs3.RuleTOC.ChapterList,
		RuleChapterName:       bs3.RuleTOC.ChapterName,
		RuleChapterURL:        bs3.RuleTOC.ChapterURL,
		RuleContentURL:        "",
		RuleCoverURL:          bs3.RuleBookInfo.CoverURL,
		RuleIntroduce:         bs3.RuleBookInfo.Intro,
		RuleSearchAuthor:      bs3.RuleSearch.Author,
		RuleSearchCoverURL:    bs3.RuleSearch.CoverURL,
		RuleSearchKind:        bs3.RuleSearch.Kind,
		RuleSearchLastChapter: bs3.RuleSearch.LastChapter,
		RuleSearchList:        bs3.RuleSearch.BookList,
		RuleSearchName:        bs3.RuleSearch.Name,
		RuleSearchNoteURL:     bs3.RuleSearch.BookURL,
		RuleSearchURL:         bs3.RuleSearch.BookURL,
	}
}

func CollectBookSources(bss2 []BookSourceV2) {
	for i := range bss2 {
		lastDotPos := strings.LastIndex(bss2[i].BookSourceURL, `.`)
		lastDashPos := strings.LastIndex(bss2[i].BookSourceURL, `-`)
		if lastDashPos > lastDotPos {
			bss2[i].BookSourceURL = bss2[i].BookSourceURL[:lastDashPos]
		}
		if bss2[i].BookSourceGroup == "" {
			bss2[i].BookSourceGroup = `未分类`
		}

		allBookSources.Add(&bss2[i])
	}
}

// ReadBookSourceFromBytes book source is stored as bytes, read and parse it
func ReadBookSourceFromBytes(c []byte) (bss2 []BookSourceV2) {
	e := json.Unmarshal(c, &bss2)
	if e != nil || len(bss2) == 0 || bss2[0].RuleChapterList == "" {
		// try v3
		var bss3 []BookSourceV3
		e = json.Unmarshal(c, &bss3)
		if e == nil && len(bss3) > 0 && bss3[0].RuleTOC.ChapterList != "" {
			// copy to v2 collection
			bss2 = []BookSourceV2{}
			for _, bs3 := range bss3 {
				bs2 := ConvertBookSourceV3ToV2(bs3)
				bss2 = append(bss2, bs2)
			}
		}
	}

	if len(bss2) == 0 {
		var bs2 BookSourceV2
		e = json.Unmarshal(c, &bs2)
		if e == nil && bs2.RuleChapterList != "" {
			bss2 = append(bss2, bs2)
		} else {
			// try v3
			var bs3 BookSourceV3
			if e = json.Unmarshal(c, &bs3); e == nil {
				bs2 := ConvertBookSourceV3ToV2(bs3)
				bss2 = append(bss2, bs2)
			}
		}
	}
	return
}

// ReadBookSourceFromLocalFileSystem book source is stored in local file, read and parse it
func ReadBookSourceFromLocalFileSystem(fileName string) (bss2 []BookSourceV2) {
	c, e := os.ReadFile(fileName)

	if e != nil {
		log.Println(e)
		return
	}

	bss2 = ReadBookSourceFromBytes(c)

	CollectBookSources(bss2)
	return
}

// ReadBookSourceFromURL book source is stored in a URL, read and parse it
func ReadBookSourceFromURL(u string) (bss2 []BookSourceV2) {
	c, e := httputil.GetBytes(u,
		http.Header{"User-Agent": []string{"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0"}},
		60*time.Second,
		3)

	if e != nil {
		log.Println(u, e)
		return
	}

	bss2 = ReadBookSourceFromBytes(c)

	CollectBookSources(bss2)
	return
}
