package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/missdeer/getnovel/ebook"
)

// Options for all command line options, long name must match field name
type Options struct {
	ListenAndServe  string  `long:"listenAndServe" description:"set http listen and serve address, example: :8080"`
	Format          string  `short:"f" long:"format" description:"set generated file format, candidate values: mobi, epub, pdf"`
	List            bool    `short:"l" long:"list" description:"list supported novel websites"`
	LeftMargin      float64 `long:"leftMargin" description:"set left margin for PDF format"`
	TopMargin       float64 `long:"topMargin" description:"set top margin for PDF format"`
	PageWidth       float64 `long:"pageWidth" description:"set page width for PDF format(unit: mm)"`
	PageHeight      float64 `long:"pageHeight" description:"set page height for PDF format(unit: mm)"`
	PageType        string  `short:"p" long:"pageType" description:"set page type for PDF format, add suffix to output file name"`
	TitleFontSize   int     `long:"titleFontSize" description:"set title font point size for PDF format"`
	ContentFontSize int     `long:"contentFontSize" description:"set content font point size for PDF format"`
	LineSpacing     float64 `long:"lineSpacing" description:"set line spacing rate for PDF format"`
	PagesPerFile    int     `long:"pagesPerFile" description:"split the big single PDF file to several smaller PDF files, how many pages should be included in a file, 0 means don't split"`
	ChaptersPerFile int     `long:"chaptersPerFile" description:"split the big single PDF file to several smaller PDF files, how many chapters should be included in a file, 0 means don't split"`
	FontFile        string  `long:"fontFile" description:"set TTF font file path"`
	RetryCount      int     `short:"r" long:"retryCount" description:"download retry count"`
	Timeout         int     `short:"t" long:"timeout" description:"download timeout seconds"`
	ParallelCount   int64   `long:"parallelCount" description:"parallel count for downloading"`
	ConfigFile      string  `short:"c" long:"configFile" description:"read configurations from local file"`
	OutputFile      string  `short:"o" long:"outputFile" description:"output file path"`
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
	fmt.Println("内建支持小说网站：")
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
				gen.SetPageSize(opts.PageWidth, opts.PageHeight)
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
		PageHeight:      841.89,
		PageWidth:       595.28,
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
		log.Fatalln("parsing flags failed", err)
		return
	}

	readLocalBookSource()

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

	if opts.ConfigFile != "" {
		if !readLocalConfigFile(&opts) && !readRemotePreset(&opts) {
			return
		}
	}

	downloadedChannel := make(chan bool)
	downloadCount := 0
	for _, novelURL := range args {
		_, e := url.Parse(novelURL)
		if e != nil {
			fmt.Println("invalid URL", novelURL)
			continue
		}
		downloadCount++
		go downloadBook(novelURL, downloadedChannel)
	}

	downloaded := false
	for i := 0; i < downloadCount; i++ {
		ch := <-downloadedChannel
		downloaded = downloaded || ch
	}

	if !downloaded {
		fmt.Println("使用方法：\n\tgetnovel 小说目录网址")
		listCommandHandler()
	}
}
