package bs

import (
	"log"
	"net/url"
	"sync"
)

// BookSourceCollection is an array that the element type is a BookSourceV2
type BookSourceCollection []*BookSourceV2

// BookSources wrapper for operating array in concurrent environment
type BookSources struct {
	BookSourceCollection
	sync.RWMutex
}

type ByBookSourceURL []*BookSourceV2

func (bss ByBookSourceURL) Len() int {
	return len(bss)
}

func (bss ByBookSourceURL) Less(i, j int) bool {
	return bss[i].BookSourceURL < bss[j].BookSourceURL
}

func (bss ByBookSourceURL) Swap(i, j int) {
	bss[i], bss[j] = bss[j], bss[i]
}

// Add add a book source to array
func (bss *BookSources) Add(bs *BookSourceV2) {
	bss.Lock()
	bss.BookSourceCollection = append(bss.BookSourceCollection, bs)
	bss.Unlock()
}

// Clear remove all elements
func (bss *BookSources) Clear() {
	bss.Lock()
	bss.BookSourceCollection = []*BookSourceV2{}
	bss.Unlock()
}

// Length returns count of book sources
func (bss *BookSources) Length() int {
	bss.RLock()
	defer bss.RUnlock()
	return len(bss.BookSourceCollection)
}

// FindBookSourceByHost find the first matched book source
func (bss *BookSources) FindBookSourceByHost(host string) *BookSourceV2 {
	u, e := url.Parse(host)
	if e != nil {
		log.Println(e)
		return nil
	}
	bss.RLock()
	defer bss.RUnlock()
	for _, v := range bss.BookSourceCollection {
		if v.BookSourceURL == host {
			return v
		}
		bsu, e := url.Parse(v.BookSourceURL)
		if e != nil {
			continue
		}
		if bsu.Host == u.Host {
			return v
		}
	}
	return nil
}

// FindBookSourcesByHost find all the matched book sources
func (bss *BookSources) FindBookSourcesByHost(host string) (res BookSources) {
	bss.RLock()
	defer bss.RUnlock()
	for _, v := range bss.BookSourceCollection {
		if v.BookSourceURL == host {
			res.BookSourceCollection = append(res.BookSourceCollection, v)
		}
	}
	return
}
