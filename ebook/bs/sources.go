package bs

import (
	"sync"
)

// BookSourceCollection is an array that the element type is a BookSource
type BookSourceCollection []*BookSource

// BookSources wrapper for operating array in concurrent environment
type BookSources struct {
	BookSourceCollection
	sync.RWMutex
}

// Add add a book source to array
func (bss *BookSources) Add(bs *BookSource) {
	bss.Lock()
	bss.BookSourceCollection = append(bss.BookSourceCollection, bs)
}

// FindBookSourceByHost find the first matched book source
func (bss *BookSources) FindBookSourceByHost(host string) (bs *BookSource) {
	bss.RLock()
	defer bss.RUnlock()
	for _, v := range bss.BookSourceCollection {
		if v.BookSourceURL == host {
			return bs
		}
	}
	return
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
