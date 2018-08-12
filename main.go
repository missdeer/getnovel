package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dfordsoft/golib/ebook"
	flags "github.com/jessevdk/go-flags"
)

// Options for all command line options
type Options struct {
	ListenAndServe  string  `long:"httpServe" description:"set http listen and serve address, example: :8080"`
	Format          string  `short:"f" long:"format" description:"set generated file format, candidate values: mobi, epub, pdf"`
	List            bool    `short:"l" long:"list" description:"list supported novel websites"`
	LeftMargin      float64 `long:"leftMargin" description:"set left margin for PDF format"`
	TopMargin       float64 `long:"topMargin" description:"set top margin for PDF format"`
	PageType        string  `short:"p" long:"pageType" description:"set page type for PDF format, candidate values: a0, a1, a2, a3, a4, a5, a6, b0, b1, b2, b3, b4, b5, b6, c0, c1, c2, c3, c4, c5, c6, dxg(=a4), 6inch(90mm x 117mm), 7inch, 10inch(=a4), pc(=a4 & 25.4mm left margin & 31.7mm top margin & 16 point title font size & 12 point content font size)"`
	TitleFontSize   int     `long:"titleFontSize" description:"set title font point size for PDF format"`
	ContentFontSize int     `long:"contentFontSize" description:"set content font point size for PDF format"`
	LineSpacing     float64 `long:"lineSpacing" description:"set line spacing rate for PDF format"`
	PagesPerFile    int     `long:"pagesPerFile" description:"split the big single PDF file to several smaller PDF files, how many pages should be included in a file, 0 means don't split"`
	ChaptersPerFile int     `long:"chaptersPerFile" description:"split the big signle PDF file to several smaller PDF files, how many chapters should be included in a file, 0 means don't split"`
	FontFile        string  `long:"fontFile" description:"set TTF font file path"`
	RetryCount      int     `short:"r" long:"retries" description:"download retry count"`
	Timeout         int     `short:"t" long:"timeout" description:"download timeout seconds"`
	ParallelCount   int64   `long:"parallel" description:"parallel count for downloading"`
	ConfigFile      string  `short:"c" long:"config" description:"read configurations from local file"`
	OutputFile      string  `short:"o" long:"output" description:"output file path"`
	FromChapter     int     `long:"fromChapter" description:"from chapter"`
	FromTitle       string  `long:"fromTitle" description:"from title"`
	ToChapter       int     `long:"toChapter" description:"to chapter"`
	ToTitle         string  `long:"toTitle" description:"to title"`
}

type tocPattern struct {
	host            string
	bookTitle       string
	bookTitlePos    int
	item            string
	articleTitlePos int
	articleURLPos   int
	isAbsoluteURL   bool
}

type pageContentMarker struct {
	host  string
	start []byte
	end   []byte
}

type novelSiteHandler struct {
	Title         string
	MatchPatterns []string
	Download      func(string, ebook.IBook)
}

var (
	novelSiteHandlers []*novelSiteHandler
	opts              Options
	sha1ver           string // sha1 revision used to build the program
	buildTime         string // when the executable was built
)

func registerNovelSiteHandler(h *novelSiteHandler) {
	novelSiteHandlers = append(novelSiteHandlers, h)
}

func listCommandHandler() {
	fmt.Println("支持小说网站：")
	for _, h := range novelSiteHandlers {
		urlMap := make(map[string]struct{})
		for _, p := range h.MatchPatterns {
			u := strings.Replace(p, `\`, ``, -1)
			idxStart := strings.Index(u, `www.`)
			idxEnd := strings.Index(u[idxStart:], `/`)
			u = u[:idxStart+idxEnd]
			urlMap[u] = struct{}{}
		}
		var urls []string
		for u := range urlMap {
			urls = append(urls, u)
		}
		fmt.Println("\t" + h.Title + ": " + strings.Join(urls, ", "))
	}
}

func readConfigFile(opts *Options) bool {
	if opts.ConfigFile != "" {
		contentFd, err := os.OpenFile(opts.ConfigFile, os.O_RDONLY, 0644)
		if err != nil {
			log.Println("opening config file ", opts.ConfigFile, " for reading failed ", err)
			return false
		}

		contentC, err := ioutil.ReadAll(contentFd)
		contentFd.Close()
		if err != nil {
			log.Println("reading config file ", opts.ConfigFile, " failed ", err)
			return false
		}

		var options map[string]interface{}
		if err = json.Unmarshal(contentC, &options); err != nil {
			log.Println("unmarshall configurations failed", err)
			return false
		}

		if f, ok := options["format"]; ok {
			if v := f.(string); len(v) > 0 {
				opts.Format = v
			}
		}
		if f, ok := options["pageType"]; ok {
			if v := f.(string); len(v) > 0 {
				opts.PageType = v
			}
		}
		if f, ok := options["fontFile"]; ok {
			if v := f.(string); len(v) > 0 {
				opts.FontFile = v
			}
		}
		if f, ok := options["fromChapter"]; ok {
			if v := f.(int); v > 0 {
				opts.FromChapter = v
			}
		}
		if f, ok := options["fromTitle"]; ok {
			if v := f.(string); len(v) > 0 {
				opts.FromTitle = v
			}
		}
		if f, ok := options["toChapter"]; ok {
			if v := f.(int); v > 0 {
				opts.ToChapter = v
			}
		}
		if f, ok := options["toTitle"]; ok {
			if v := f.(string); len(v) > 0 {
				opts.ToTitle = v
			}
		}

		if f, ok := options["leftMargin"]; ok {
			if v := f.(float64); v > 0 {
				opts.LeftMargin = v
			}
		}
		if f, ok := options["topMargin"]; ok {
			if v := f.(float64); v > 0 {
				opts.TopMargin = v
			}
		}
		if f, ok := options["lineSpacing"]; ok {
			if v := f.(float64); v > 0 {
				opts.LineSpacing = v
			}
		}
		if f, ok := options["titleFontSize"]; ok {
			if v := f.(int); v > 0 {
				opts.TitleFontSize = v
			}
		}
		if f, ok := options["contentFontSize"]; ok {
			if v := f.(int); v > 0 {
				opts.ContentFontSize = v
			}
		}
		if f, ok := options["pagesPerFile"]; ok {
			if v := f.(int); v > 0 {
				opts.PagesPerFile = v
			}
		}
		if f, ok := options["chaptersPerFile"]; ok {
			if v := f.(int); v > 0 {
				opts.ChaptersPerFile = v
			}
		}
		if f, ok := options["retries"]; ok {
			if v := f.(int); v > 0 {
				opts.RetryCount = v
			}
		}
		if f, ok := options["timeout"]; ok {
			if v := f.(int); v > 0 {
				opts.Timeout = v
			}
		}
		if f, ok := options["parallel"]; ok {
			if v := f.(int64); v > 0 {
				opts.ParallelCount = v
			}
		}
	}
	return true
}

func downloadBook(novelURL string, ch chan bool) {
	for _, h := range novelSiteHandlers {
		for _, pattern := range h.MatchPatterns {
			r, _ := regexp.Compile(pattern)
			if r.MatchString(novelURL) {
				gen := ebook.NewBook(opts.Format)
				gen.SetFontSize(opts.TitleFontSize, opts.ContentFontSize)
				gen.SetLineSpacing(opts.LineSpacing)
				gen.PagesPerFile(opts.PagesPerFile)
				gen.ChaptersPerFile(opts.ChaptersPerFile)
				gen.SetMargins(opts.LeftMargin, opts.TopMargin)
				gen.SetPageType(opts.PageType)
				gen.SetFontFile(opts.FontFile)
				gen.Output(opts.OutputFile)
				gen.Info()
				h.Download(novelURL, gen)
				fmt.Println("downloaded", novelURL)
				ch <- true
				return
			}
		}
	}
	fmt.Println("not downloaded", novelURL)
	ch <- false
}

func main() {
	fmt.Println("getnovel SHA1:", sha1ver, "build at", buildTime)
	if len(os.Args) < 2 {
		fmt.Println("使用方法：\n\tgetnovel 小说目录网址")
		listCommandHandler()
		return
	}

	opts = Options{
		Format:          "mobi",
		List:            false,
		LeftMargin:      10,
		TopMargin:       10,
		PageType:        "a4",
		TitleFontSize:   24,
		ContentFontSize: 18,
		LineSpacing:     1.2,
		PagesPerFile:    0,
		ChaptersPerFile: 0,
		FontFile:        filepath.Join("fonts", "CustomFont.ttf"),
		RetryCount:      3,
		Timeout:         60,
		ParallelCount:   10,
	}

	args, err := flags.Parse(&opts)
	if err != nil {
		return
	}

	if opts.List {
		listCommandHandler()
		return
	}

	if opts.ListenAndServe != "" {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}
		ifaces, err := net.Interfaces()
		var ips []string
		for _, i := range ifaces {

			addrs, err := i.Addrs()
			if err != nil {
				fmt.Println(err)
				continue
			}
			for _, addr := range addrs {
				switch v := addr.(type) {
				case *net.IPNet:
					if v.IP.IsLoopback() || v.IP.IsLinkLocalMulticast() || v.IP.IsLinkLocalUnicast() {
						break
					}
					ips = append(ips, "\t"+v.IP.String())
				case *net.IPAddr:
					if v.IP.IsLoopback() || v.IP.IsLinkLocalMulticast() || v.IP.IsLinkLocalUnicast() {
						break
					}
					ips = append(ips, "\t"+v.IP.String())
				}
			}
		}
		fmt.Println("Local IP:")
		fmt.Println(strings.Join(ips, "\n"))
		fmt.Println("starting http server on", opts.ListenAndServe)
		log.Fatal(http.ListenAndServe(opts.ListenAndServe, http.FileServer(http.Dir(dir))))
		return
	}

	if !readConfigFile(&opts) {
		return
	}

	downloadedChannel := make(chan bool)
	donwloadCount := 0
	for _, novelURL := range args {
		_, e := url.Parse(novelURL)
		if e != nil {
			fmt.Println("invalid URL", novelURL)
			continue
		}
		donwloadCount++
		go downloadBook(novelURL, downloadedChannel)
	}

	downloaded := false
	for i := 0; i < donwloadCount; i++ {
		ch := <-downloadedChannel
		downloaded = (downloaded || ch)
	}

	if !downloaded {
		fmt.Println("使用方法：\n\tgetnovel 小说目录网址")
		listCommandHandler()
	}
}
