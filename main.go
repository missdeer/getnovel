package main

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/dfordsoft/golib/ebook"
	flags "github.com/jessevdk/go-flags"
)

// Options for all command line options
type Options struct {
	Format          string  `short:"f" long:"format" description:"set generated file format, candidate values: mobi, epub, pdf"`
	List            bool    `short:"l" long:"list" description:"list supported novel websites"`
	LeftMargin      float64 `long:"leftMargin" description:"set left margin for PDF format"`
	TopMargin       float64 `long:"topMargin" description:"set top margin for PDF format"`
	PageType        string  `long:"pageType" description:"set page type for PDF format, candidate values: a0, a1, a2, a3, a4, a5, a6, b0, b1, b2, b3, b4, b5, b6, c0, c1, c2, c3, c4, c5, c6, dxg(=a4), 6inch(90mm x 117mm), 7inch, 10inch(=a4)"`
	TitleFontSize   int     `long:"titleFontSize" description:"set title font point size for PDF format"`
	ContentFontSize int     `long:"contentFontSize" description:"set content font point size for PDF format"`
	LineSpacing     float64 `long:"lineSpacing" description:"set line spacing rate for PDF format"`
	FontFamily      string  `long:"fontFamily" description:"set font family name"`
	FontFile        string  `long:"fontFile" description:"set TTF font file path"`
	RetryCount      int     `short:"r" long:"retries" description:"download retry count"`
	Timeout         int     `short:"t" long:"timeout" description:"download timeout seconds"`
	ParallelCount   int64   `short:"p" long:"parallel" description:"parallel count for downloading"`
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
	Download      func(string)
}

var (
	novelSiteHandlers []*novelSiteHandler
	gen               ebook.IBook
	opts              Options
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

func main() {
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
		FontFamily:      "CustomFont",
		FontFile:        "fonts/CustomFont.ttf",
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

	downloaded := false
	for _, novelURL := range args {
		_, e := url.Parse(novelURL)
		if e != nil {
			fmt.Println("invalid URL", novelURL)
			continue
		}
		for _, h := range novelSiteHandlers {
			for _, pattern := range h.MatchPatterns {
				r, _ := regexp.Compile(pattern)
				if r.MatchString(novelURL) {
					gen = ebook.NewBook(opts.Format)
					gen.SetFontSize(opts.TitleFontSize, opts.ContentFontSize)
					gen.SetLineSpacing(opts.LineSpacing)
					gen.SetMargins(opts.LeftMargin, opts.TopMargin)
					gen.SetPageType(opts.PageType)
					gen.SetFontFamily(opts.FontFamily)
					gen.SetFontFile(opts.FontFile)
					gen.Info()
					h.Download(novelURL)
					downloaded = true
				}
			}
		}
	}
	if !downloaded {
		fmt.Println("使用方法：\n\tgetnovel 小说目录网址")
		listCommandHandler()
	}
}
