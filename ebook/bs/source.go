package bs

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/missdeer/getnovel/legado"
	"github.com/missdeer/golib/httputil"
)

var (
	allBookSources BookSources
)

// LegadoSourceCollection is a thread-safe collection of legado book sources
type LegadoSourceCollection struct {
	sources []*LegadoBookSource
	sync.RWMutex
}

// Add adds a legado source to the collection
func (lsc *LegadoSourceCollection) Add(ls *LegadoBookSource) {
	lsc.Lock()
	lsc.sources = append(lsc.sources, ls)
	lsc.Unlock()
}

// Clear removes all legado sources
func (lsc *LegadoSourceCollection) Clear() {
	lsc.Lock()
	lsc.sources = nil
	lsc.Unlock()
}

// Length returns the number of legado sources
func (lsc *LegadoSourceCollection) Length() int {
	lsc.RLock()
	defer lsc.RUnlock()
	return len(lsc.sources)
}

// Range iterates over all legado sources with a callback
func (lsc *LegadoSourceCollection) Range(fn func(*LegadoBookSource) bool) {
	lsc.RLock()
	defer lsc.RUnlock()
	for _, ls := range lsc.sources {
		if !fn(ls) {
			break
		}
	}
}

// legadoSources is the global thread-safe collection
var legadoSources LegadoSourceCollection

type SearchOutput map[string][]*Book

func SortSearchOutput(so SearchOutput) []string {
	sortedResult := make(map[string]int, len(so))
	// var keys = make([]int, len(so))
	var newKeys = make([]string, 0, len(so)) // Fixed: use 0 length with capacity
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
		defer close(c) // Fixed: close channel when done
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
			if !ok {
				timeout = true // Channel closed, exit loop
				break
			}
			if _, ok = result[i.Name]; !ok {
				result[i.Name] = []*Book{i}
			} else {
				result[i.Name] = append(result[i.Name], i)
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

// ReadLegadoSourceFromBytes parses legado format book sources from bytes
func ReadLegadoSourceFromBytes(c []byte) []*legado.BookSource {
	sources, err := legado.LoadBookSources(c)
	if err != nil {
		log.Printf("Failed to parse legado book sources: %v", err)
		return nil
	}

	// Convert to pointer slice
	result := make([]*legado.BookSource, len(sources))
	for i := range sources {
		result[i] = &sources[i]
	}
	return result
}

// ReadLegadoSourceFromLocalFileSystem reads legado book sources from a local file
func ReadLegadoSourceFromLocalFileSystem(fileName string) []*legado.BookSource {
	c, e := os.ReadFile(fileName)
	if e != nil {
		log.Println(e)
		return nil
	}

	sources := ReadLegadoSourceFromBytes(c)
	CollectLegadoSources(sources)
	return sources
}

// ReadLegadoSourceFromURL reads legado book sources from a URL
func ReadLegadoSourceFromURL(u string) []*legado.BookSource {
	c, e := httputil.GetBytes(u,
		http.Header{"User-Agent": []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36"}},
		60*time.Second,
		3)

	if e != nil {
		log.Println(u, e)
		return nil
	}

	sources := ReadLegadoSourceFromBytes(c)
	CollectLegadoSources(sources)
	return sources
}

// CollectLegadoSources adds legado sources to the global collection
func CollectLegadoSources(sources []*legado.BookSource) {
	for _, source := range sources {
		if source == nil {
			continue
		}
		ls := NewLegadoBookSource(source)
		legadoSources.Add(ls) // Fixed: use thread-safe collection

		// Also add to V2 collection for backward compatibility
		bs2 := ConvertLegadoToV2(source)
		allBookSources.Add(&bs2)
	}
}

// FindLegadoSourceByHost finds a legado book source by host
func FindLegadoSourceByHost(host string) *LegadoBookSource {
	var result *LegadoBookSource
	legadoSources.Range(func(ls *LegadoBookSource) bool {
		if ls.Source == nil {
			return true // continue
		}
		if strings.Contains(ls.Source.BookSourceURL, host) {
			result = ls
			return false // stop
		}
		return true // continue
	})
	return result
}

// FindLegadoSourceByURL finds a legado book source that matches the given URL
func FindLegadoSourceByURL(bookURL string) *LegadoBookSource {
	var result *LegadoBookSource
	legadoSources.Range(func(ls *LegadoBookSource) bool {
		if ls.Source == nil {
			return true // continue
		}
		// Check if the book URL matches the source's URL pattern
		if ls.Source.BookURLPattern != "" {
			// Compile and match the pattern
			re, err := regexp.Compile(ls.Source.BookURLPattern)
			if err == nil && re.MatchString(bookURL) {
				result = ls
				return false // stop
			}
			// Pattern didn't match, try next source
			return true // continue
		}
		// Simple host matching
		if strings.Contains(bookURL, strings.TrimPrefix(strings.TrimPrefix(ls.Source.BookSourceURL, "https://"), "http://")) {
			result = ls
			return false // stop
		}
		return true // continue
	})
	return result
}

// SearchBooksWithLegado searches books using legado sources
func SearchBooksWithLegado(keyword string, page int) SearchOutput {
	c := make(chan *Book, 10)
	result := make(SearchOutput)

	go func() {
		defer close(c)
		legadoSources.Range(func(ls *LegadoBookSource) bool {
			if ls == nil || ls.Source == nil || !ls.Source.Enabled {
				return true // continue
			}
			books, err := ls.LegadoSearchBook(keyword, page)
			if err != nil {
				log.Printf("Search error on %s: %v", ls.Source.BookSourceName, err)
				return true // continue
			}
			for _, book := range books {
				c <- book
			}
			return true // continue
		})
	}()

	for timeout := false; !timeout; {
		select {
		case book, ok := <-c:
			if !ok {
				timeout = true
				break
			}
			if book != nil && book.Name != "" {
				if _, exists := result[book.Name]; !exists {
					result[book.Name] = []*Book{book}
				} else {
					result[book.Name] = append(result[book.Name], book)
				}
			}
		case <-time.After(30 * time.Second):
			log.Printf("Search timeout, exiting...")
			timeout = true
		}
	}

	return result
}

// LoadBookSourcesFromDirectory loads all JSON book source files from a directory
func LoadBookSourcesFromDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".json") {
			continue
		}

		filePath := filepath.Join(dir, name) // Fixed: use filepath.Join
		// Try to load as legado format first
		sources := ReadLegadoSourceFromLocalFileSystem(filePath)
		if len(sources) > 0 {
			log.Printf("Loaded %d legado sources from %s", len(sources), name)
			continue
		}

		// Fall back to V2/V3 format
		bss2 := ReadBookSourceFromLocalFileSystem(filePath)
		if len(bss2) > 0 {
			log.Printf("Loaded %d V2/V3 sources from %s", len(bss2), name)
		}
	}

	return nil
}

// LoadBookSourcesFromURLs loads book sources from a list of URLs
func LoadBookSourcesFromURLs(urls []string) {
	for _, u := range urls {
		// Try to load as legado format first
		sources := ReadLegadoSourceFromURL(u)
		if len(sources) > 0 {
			log.Printf("Loaded %d legado sources from %s", len(sources), u)
			continue
		}

		// Fall back to V2/V3 format
		bss2 := ReadBookSourceFromURL(u)
		if len(bss2) > 0 {
			log.Printf("Loaded %d V2/V3 sources from %s", len(bss2), u)
		}
	}
}

// GetBookSourceCount returns the total number of loaded book sources
func GetBookSourceCount() int {
	allBookSources.RLock()
	defer allBookSources.RUnlock()
	return len(allBookSources.BookSourceCollection)
}

// GetLegadoSourceCount returns the number of loaded legado sources
func GetLegadoSourceCount() int {
	return legadoSources.Length()
}

// ClearAllSources clears all loaded book sources
func ClearAllSources() {
	allBookSources.Lock()
	allBookSources.BookSourceCollection = nil
	allBookSources.Unlock()
	legadoSources.Clear()
}
