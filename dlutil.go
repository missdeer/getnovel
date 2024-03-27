package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/missdeer/getnovel/config"
	"github.com/missdeer/getnovel/ebook"
	"github.com/missdeer/golib/httputil"
	"golang.org/x/sync/semaphore"
)

type ContentInfo struct {
	Index           int
	Title           string
	LinkURL         string
	ContentFilePath string
}
type DownloadUtil struct {
	ContentExtractor        func(string, []byte) []byte
	ContentLinkPreprocessor func(string) (string, http.Header)
	Generator               ebook.IBook
	TempDir                 string
	CurrentPage             int32
	MaxPage                 int32
	Quit                    chan bool
	Content                 chan ContentInfo
	Buffer                  []ContentInfo
	StartContent            *ContentInfo
	EndContent              *ContentInfo
	Ctx                     context.Context
	Semaphore               *semaphore.Weighted
}

func NewDownloadUtil(contentExtractor func(string, []byte) []byte, contentLinkPreprocessor func(string) (string, http.Header), generator ebook.IBook) (dlutil *DownloadUtil) {
	dlutil = &DownloadUtil{
		ContentExtractor:        contentExtractor,
		ContentLinkPreprocessor: contentLinkPreprocessor,
		Generator:               generator,
		Quit:                    make(chan bool),
		Ctx:                     context.TODO(),
		Semaphore:               semaphore.NewWeighted(config.Opts.ParallelCount),
		Content:                 make(chan ContentInfo),
	}
	if config.Opts.FromChapter != 0 {
		dlutil.StartContent = &ContentInfo{Index: config.Opts.FromChapter}
	}
	if config.Opts.FromTitle != "" {
		dlutil.StartContent = &ContentInfo{Title: config.Opts.FromTitle, Index: math.MaxInt32}
	}
	if config.Opts.ToChapter != 0 {
		dlutil.EndContent = &ContentInfo{Index: config.Opts.ToChapter}
	}
	if config.Opts.ToTitle != "" {
		dlutil.EndContent = &ContentInfo{Title: config.Opts.ToTitle}
	}
	var err error
	dlutil.TempDir, err = os.MkdirTemp("", uuid.New().String())
	if err != nil {
		log.Fatal("creating temporary directory failed", err)
	}
	return
}

func (dlutil *DownloadUtil) Wait() {
	<-dlutil.Quit
	os.RemoveAll(dlutil.TempDir)
}

func (dlutil *DownloadUtil) PreprocessURL(index int, title string, link string) (returnImmediately bool, reachEnd bool) {
	atomic.StoreInt32(&dlutil.MaxPage, int32(index))
	if dlutil.StartContent != nil {
		if dlutil.StartContent.Index == index {
			dlutil.StartContent.Title = title
			dlutil.StartContent.LinkURL = link
			atomic.StoreInt32(&dlutil.CurrentPage, int32(index-1))
		}
		if dlutil.StartContent.Title == title {
			dlutil.StartContent.Index = index
			dlutil.StartContent.LinkURL = link
			atomic.StoreInt32(&dlutil.CurrentPage, int32(index-1))
		}

		if dlutil.StartContent.Index > index {
			return true, false
		}
	}
	if dlutil.EndContent != nil {
		if dlutil.EndContent.Index == index {
			dlutil.EndContent.Title = title
			dlutil.EndContent.LinkURL = link
		}
		if dlutil.EndContent.Title == title {
			dlutil.EndContent.Index = index
			dlutil.EndContent.LinkURL = link
		}
		atomic.StoreInt32(&dlutil.MaxPage, int32(dlutil.EndContent.Index))

		if index > dlutil.EndContent.Index && dlutil.EndContent.Index != 0 {
			return true, true
		}
	}
	return false, false
}

func (dlutil *DownloadUtil) AddURL(index int, title string, link string) (reachEnd bool) {
	if r, e := dlutil.PreprocessURL(index, title, link); r == true {
		return e
	}
	// semaphore
	dlutil.Semaphore.Acquire(dlutil.Ctx, 1)
	go func() {
		filePath := fmt.Sprintf("%s/%d.txt", dlutil.TempDir, index)
		contentFd, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Println("opening file", filePath, "for writing failed ", err)
			return
		}

		theURL, _ := url.Parse(link)
		headers := http.Header{
			"Referer":                   []string{fmt.Sprintf("%s://%s", theURL.Scheme, theURL.Host)},
			"User-Agent":                []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0"},
			"Accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"},
			"Accept-Language":           []string{`en-US,en;q=0.8`},
			"Upgrade-Insecure-Requests": []string{"1"},
		}
		if dlutil.ContentLinkPreprocessor != nil {
			link, headers = dlutil.ContentLinkPreprocessor(link)
		}
		rawPageContent, err := httputil.GetBytes(link, headers, time.Duration(config.Opts.Timeout)*time.Second, config.Opts.RetryCount)
		if err != nil {
			log.Println("getting chapter content from", link, "failed ", err)
			return
		}
		contentFd.Write(dlutil.ContentExtractor(link, rawPageContent))
		contentFd.Close()

		dlutil.Content <- ContentInfo{
			Index:           index,
			Title:           title,
			LinkURL:         link,
			ContentFilePath: filePath,
		}
		dlutil.Semaphore.Release(1)
	}()
	return false
}

func (dlutil *DownloadUtil) BufferHandler(contentInfo ContentInfo) (exit bool) {
	fmt.Println(contentInfo.Title, contentInfo.LinkURL)
	// insert into local buffer
	if len(dlutil.Buffer) == 0 || dlutil.Buffer[0].Index > contentInfo.Index {
		// push front
		dlutil.Buffer = append([]ContentInfo{contentInfo}, dlutil.Buffer...)
		// check local buffer to pick items to generator
		for ; len(dlutil.Buffer) > 0 && int32(dlutil.Buffer[0].Index) == atomic.LoadInt32(&dlutil.CurrentPage)+1; dlutil.Buffer = dlutil.Buffer[1:] {
			contentFd, err := os.OpenFile(dlutil.Buffer[0].ContentFilePath, os.O_RDONLY, 0644)
			if err != nil {
				log.Println("opening file ", dlutil.Buffer[0].ContentFilePath, " for reading failed ", err)
				continue
			}

			contentC, err := io.ReadAll(contentFd)
			contentFd.Close()
			if err != nil {
				log.Println("reading file ", dlutil.Buffer[0].ContentFilePath, " failed ", err)
				continue
			}
			os.Remove(dlutil.Buffer[0].ContentFilePath)

			dlutil.Generator.AppendContent(dlutil.Buffer[0].Title, dlutil.Buffer[0].LinkURL, string(contentC))
			atomic.AddInt32(&dlutil.CurrentPage, 1)
		}

		if atomic.LoadInt32(&dlutil.CurrentPage) == atomic.LoadInt32(&dlutil.MaxPage) {
			dlutil.Quit <- true
			return true
		}
		return false
	}
	if dlutil.Buffer[len(dlutil.Buffer)-1].Index < contentInfo.Index {
		// push back
		dlutil.Buffer = append(dlutil.Buffer, contentInfo)
		return false
	}
	for i := 0; i < len(dlutil.Buffer)-1; i++ {
		if dlutil.Buffer[i].Index < contentInfo.Index && dlutil.Buffer[i+1].Index > contentInfo.Index {
			// insert at i+1
			dlutil.Buffer = append(dlutil.Buffer[:i+1], append([]ContentInfo{contentInfo}, dlutil.Buffer[i+1:]...)...)
			return false
		}
	}

	return false
}

func (dlutil *DownloadUtil) Process() {
	go func() {
		for contentInfo := range dlutil.Content {
			if dlutil.BufferHandler(contentInfo) {
				return
			}
		}
	}()
}
