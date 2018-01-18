package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/dfordsoft/golib/ebook"
	"github.com/satori/go.uuid"
	"golang.org/x/sync/semaphore"
)

type contentUtil struct {
	index   int
	title   string
	link    string
	content string
}
type downloadUtil struct {
	sync.Mutex
	downloader   func(string) []byte
	generator    ebook.IBook
	tempDir      string
	currentPage  int
	maxPage      int
	quit         chan bool
	content      chan contentUtil
	buffer       []contentUtil
	startContent *contentUtil
	endContent   *contentUtil
	ctx          context.Context
	semaphore    *semaphore.Weighted
}

func newDownloadUtil(dl func(string) []byte, generator ebook.IBook) (du *downloadUtil) {
	du = &downloadUtil{
		downloader: dl,
		generator:  generator,
		quit:       make(chan bool),
		ctx:        context.TODO(),
		semaphore:  semaphore.NewWeighted(opts.ParallelCount),
		content:    make(chan contentUtil),
	}
	if opts.FromChapter != 0 {
		du.startContent = &contentUtil{index: opts.FromChapter}
	}
	if opts.FromTitle != "" {
		du.startContent = &contentUtil{title: opts.FromTitle}
	}
	if opts.ToChapter != 0 {
		du.endContent = &contentUtil{index: opts.ToChapter}
	}
	if opts.ToTitle != "" {
		du.endContent = &contentUtil{title: opts.ToTitle}
	}
	var err error
	du.tempDir, err = ioutil.TempDir("", uuid.Must(uuid.NewV4()).String())
	if err != nil {
		log.Fatal("creating temporary directory failed", err)
	}
	return
}

func (du *downloadUtil) wait() {
	<-du.quit
	os.RemoveAll(du.tempDir)
}

func (du *downloadUtil) preprocessURL(index int, title string, link string) (returnImmediately bool) {
	du.Lock()
	defer du.Unlock()
	du.maxPage = index
	if du.startContent != nil {
		if du.startContent.index == index {
			du.startContent.title = title
			du.startContent.link = link
			du.currentPage = index - 1
		}
		if du.startContent.title == title {
			du.startContent.index = index
			du.startContent.link = link
			du.currentPage = index - 1
		}

		if du.startContent.index > index {
			return true
		}
	}
	if du.endContent != nil {
		if du.endContent.index == index {
			du.endContent.title = title
			du.endContent.link = link
		}
		if du.endContent.title == title {
			du.endContent.index = index
			du.endContent.link = link
		}

		du.maxPage = du.endContent.index

		if index > du.endContent.index && du.endContent.index != 0 {
			return true
		}
	}
	return false
}

func (du *downloadUtil) addURL(index int, title string, link string) {
	if du.preprocessURL(index, title, link) {
		return
	}
	// semaphore
	du.semaphore.Acquire(du.ctx, 1)
	go func() {
		filePath := fmt.Sprintf("%s/%d.txt", du.tempDir, index)
		contentFd, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Println("opening file", filePath, "for writing failed ", err)
			return
		}
		contentFd.Write(du.downloader(link))
		contentFd.Close()

		du.content <- contentUtil{
			index:   index,
			title:   title,
			link:    link,
			content: filePath,
		}
		du.semaphore.Release(1)
	}()
}

func (du *downloadUtil) bufferHandler(cu contentUtil) (exit bool) {
	fmt.Println(cu.title, cu.link)
	du.Lock()
	defer du.Unlock()
	// insert into local buffer
	if len(du.buffer) == 0 || du.buffer[0].index > cu.index {
		// push front
		du.buffer = append([]contentUtil{cu}, du.buffer...)
		// check local buffer to pick items to generator
		for ; len(du.buffer) > 0 && du.buffer[0].index == du.currentPage+1; du.buffer = du.buffer[1:] {
			contentFd, err := os.OpenFile(du.buffer[0].content, os.O_RDONLY, 0644)
			if err != nil {
				log.Println("opening file ", du.buffer[0].content, " for reading failed ", err)
				continue
			}

			contentC, err := ioutil.ReadAll(contentFd)
			if err != nil {
				log.Println("reading file ", du.buffer[0].content, " failed ", err)
				continue
			}
			contentFd.Close()
			os.Remove(du.buffer[0].content)

			du.generator.AppendContent(du.buffer[0].title, du.buffer[0].link, string(contentC))
			du.currentPage++
		}

		if du.currentPage == du.maxPage {
			du.quit <- true
			return true
		}
		return false
	}
	if du.buffer[len(du.buffer)-1].index < cu.index {
		// push back
		du.buffer = append(du.buffer, cu)
		return false
	}
	for i := 0; i < len(du.buffer)-1; i++ {
		if du.buffer[i].index < cu.index && du.buffer[i+1].index > cu.index {
			// insert at i+1
			du.buffer = append(du.buffer[:i+1], append([]contentUtil{cu}, du.buffer[i+1:]...)...)
			return false
		}
	}

	return false
}

func (du *downloadUtil) process() {
	go func() {
		for {
			select {
			case cu := <-du.content:
				if du.bufferHandler(cu) {
					return
				}
			}
		}
	}()
}
