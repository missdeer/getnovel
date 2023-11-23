package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/missdeer/getnovel/ebook"
	"github.com/missdeer/golib/httputil"
	"golang.org/x/sync/semaphore"
)

type ContentUtil struct {
	Index   int
	Title   string
	LinkURL string
	Content string
}
type DownloadUtil struct {
	ContentExtractor func([]byte) []byte
	Generator        ebook.IBook
	TempDir          string
	CurrentPage      int32
	MaxPage          int32
	Quit             chan bool
	Content          chan ContentUtil
	Buffer           []ContentUtil
	StartContent     *ContentUtil
	EndContent       *ContentUtil
	Ctx              context.Context
	Semaphore        *semaphore.Weighted
}

func NewDownloadUtil(extractor func([]byte) []byte, generator ebook.IBook) (du *DownloadUtil) {
	du = &DownloadUtil{
		ContentExtractor: extractor,
		Generator:        generator,
		Quit:             make(chan bool),
		Ctx:              context.TODO(),
		Semaphore:        semaphore.NewWeighted(opts.ParallelCount),
		Content:          make(chan ContentUtil),
	}
	if opts.FromChapter != 0 {
		du.StartContent = &ContentUtil{Index: opts.FromChapter}
	}
	if opts.FromTitle != "" {
		du.StartContent = &ContentUtil{Title: opts.FromTitle, Index: math.MaxInt32}
	}
	if opts.ToChapter != 0 {
		du.EndContent = &ContentUtil{Index: opts.ToChapter}
	}
	if opts.ToTitle != "" {
		du.EndContent = &ContentUtil{Title: opts.ToTitle}
	}
	var err error
	du.TempDir, err = ioutil.TempDir("", uuid.New().String())
	if err != nil {
		log.Fatal("creating temporary directory failed", err)
	}
	return
}

func (du *DownloadUtil) Wait() {
	<-du.Quit
	os.RemoveAll(du.TempDir)
}

func (du *DownloadUtil) PreprocessURL(index int, title string, link string) (returnImmediately bool, reachEnd bool) {
	atomic.StoreInt32(&du.MaxPage, int32(index))
	if du.StartContent != nil {
		if du.StartContent.Index == index {
			du.StartContent.Title = title
			du.StartContent.LinkURL = link
			atomic.StoreInt32(&du.CurrentPage, int32(index-1))
		}
		if du.StartContent.Title == title {
			du.StartContent.Index = index
			du.StartContent.LinkURL = link
			atomic.StoreInt32(&du.CurrentPage, int32(index-1))
		}

		if du.StartContent.Index > index {
			return true, false
		}
	}
	if du.EndContent != nil {
		if du.EndContent.Index == index {
			du.EndContent.Title = title
			du.EndContent.LinkURL = link
		}
		if du.EndContent.Title == title {
			du.EndContent.Index = index
			du.EndContent.LinkURL = link
		}
		atomic.StoreInt32(&du.MaxPage, int32(du.EndContent.Index))

		if index > du.EndContent.Index && du.EndContent.Index != 0 {
			return true, true
		}
	}
	return false, false
}

func (du *DownloadUtil) AddURL(index int, title string, link string) (reachEnd bool) {
	if r, e := du.PreprocessURL(index, title, link); r == true {
		return e
	}
	// semaphore
	du.Semaphore.Acquire(du.Ctx, 1)
	go func() {
		filePath := fmt.Sprintf("%s/%d.txt", du.TempDir, index)
		contentFd, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Println("opening file", filePath, "for writing failed ", err)
			return
		}

		theURL, _ := url.Parse(link)
		headers := http.Header{
			"Referer":                   []string{fmt.Sprintf("%s://%s", theURL.Scheme, theURL.Host)},
			"User-Agent":                []string{"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0"},
			"Accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"},
			"Accept-Language":           []string{`en-US,en;q=0.8`},
			"Upgrade-Insecure-Requests": []string{"1"},
		}
		rawPageContent, err := httputil.GetBytes(link, headers, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
		if err != nil {
			log.Println("getting chapter content from", link, "failed ", err)
			return
		}
		contentFd.Write(du.ContentExtractor(rawPageContent))
		contentFd.Close()

		du.Content <- ContentUtil{
			Index:   index,
			Title:   title,
			LinkURL: link,
			Content: filePath,
		}
		du.Semaphore.Release(1)
	}()
	return false
}

func (du *DownloadUtil) BufferHandler(cu ContentUtil) (exit bool) {
	fmt.Println(cu.Title, cu.LinkURL)
	// insert into local buffer
	if len(du.Buffer) == 0 || du.Buffer[0].Index > cu.Index {
		// push front
		du.Buffer = append([]ContentUtil{cu}, du.Buffer...)
		// check local buffer to pick items to generator
		for ; len(du.Buffer) > 0 && int32(du.Buffer[0].Index) == atomic.LoadInt32(&du.CurrentPage)+1; du.Buffer = du.Buffer[1:] {
			contentFd, err := os.OpenFile(du.Buffer[0].Content, os.O_RDONLY, 0644)
			if err != nil {
				log.Println("opening file ", du.Buffer[0].Content, " for reading failed ", err)
				continue
			}

			contentC, err := io.ReadAll(contentFd)
			contentFd.Close()
			if err != nil {
				log.Println("reading file ", du.Buffer[0].Content, " failed ", err)
				continue
			}
			os.Remove(du.Buffer[0].Content)

			du.Generator.AppendContent(du.Buffer[0].Title, du.Buffer[0].LinkURL, string(contentC))
			atomic.AddInt32(&du.CurrentPage, 1)
		}

		if atomic.LoadInt32(&du.CurrentPage) == atomic.LoadInt32(&du.MaxPage) {
			du.Quit <- true
			return true
		}
		return false
	}
	if du.Buffer[len(du.Buffer)-1].Index < cu.Index {
		// push back
		du.Buffer = append(du.Buffer, cu)
		return false
	}
	for i := 0; i < len(du.Buffer)-1; i++ {
		if du.Buffer[i].Index < cu.Index && du.Buffer[i+1].Index > cu.Index {
			// insert at i+1
			du.Buffer = append(du.Buffer[:i+1], append([]ContentUtil{cu}, du.Buffer[i+1:]...)...)
			return false
		}
	}

	return false
}

func (du *DownloadUtil) Process() {
	go func() {
		for {
			select {
			case cu := <-du.Content:
				if du.BufferHandler(cu) {
					return
				}
			}
		}
	}()
}
