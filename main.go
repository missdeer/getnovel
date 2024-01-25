package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/aarzilli/golua/lua"
	"github.com/jessevdk/go-flags"
	"github.com/missdeer/getnovel/ebook"
	"github.com/missdeer/golib/httputil"
)

// Options for all command line options, long name must match field name
type Options struct {
	InsecureSkipVerify         bool    `short:"V" long:"insecureSkipVerify" description:"if true, TLS accepts any certificate"`
	ListenAndServe             string  `short:"s" long:"listenAndServe" description:"set http listen and serve address, example: :8080"`
	Format                     string  `short:"f" long:"format" description:"set generated file format, candidate values: mobi, epub, pdf, html, txt"`
	List                       bool    `short:"l" long:"list" description:"list supported novel websites"`
	LeftMargin                 float64 `long:"leftMargin" description:"set left margin for PDF format"`
	TopMargin                  float64 `long:"topMargin" description:"set top margin for PDF format"`
	PageWidth                  float64 `long:"pageWidth" description:"set page width for PDF format(unit: mm)"`
	PageHeight                 float64 `long:"pageHeight" description:"set page height for PDF format(unit: mm)"`
	PageType                   string  `short:"p" long:"pageType" description:"set page type for PDF format, add suffix to output file name"`
	TitleFontSize              int     `long:"titleFontSize" description:"set title font point size for PDF format"`
	ContentFontSize            int     `long:"contentFontSize" description:"set content font point size for PDF format"`
	LineSpacing                float64 `long:"lineSpacing" description:"set line spacing rate for PDF format"`
	PagesPerFile               int     `long:"pagesPerFile" description:"split the big single PDF file to several smaller PDF files, how many pages should be included in a file, 0 means don't split"`
	ChaptersPerFile            int     `long:"chaptersPerFile" description:"split the big single PDF file to several smaller PDF files, how many chapters should be included in a file, 0 means don't split"`
	FontFile                   string  `long:"fontFile" description:"set TTF font file path"`
	H1FontFamily               string  `long:"h1FontFamily" description:"set H1 font family for mobi/epub/html format"`
	H1FontSize                 string  `long:"h1FontSize" description:"set H1 font size for mobi/epub/html format"`
	H2FontFamily               string  `long:"h2FontFamily" description:"set H2 font family for mobi/epub/html format"`
	H2FontSize                 string  `long:"h2FontSize" description:"set H2 font size for mobi/epub/html format"`
	BodyFontFamily             string  `long:"bodyFontFamily" description:"set body font family for mobi/epub/html format"`
	BodyFontSize               string  `long:"bodyFontSize" description:"set body font size for mobi/epub/html format"`
	ParaFontFamily             string  `long:"paraFontFamily" description:"set paragraph font family for mobi/epub/html format"`
	ParaFontSize               string  `long:"paraFontSize" description:"set paragraph font size for mobi/epub/html format"`
	ParaLineHeight             string  `long:"paraLineHeight" description:"set paragraph line height for mobi/epub/html format"`
	RetryCount                 int     `short:"r" long:"retryCount" description:"download retry count"`
	Timeout                    int     `short:"t" long:"timeout" description:"download timeout seconds"`
	ParallelCount              int64   `long:"parallelCount" description:"parallel count for downloading"`
	ConfigFile                 string  `short:"c" long:"configFile" description:"read configurations from local file"`
	OutputFile                 string  `short:"o" long:"outputFile" description:"output file path"`
	FromChapter                int     `long:"fromChapter" description:"from chapter"`
	FromTitle                  string  `long:"fromTitle" description:"from title"`
	ToChapter                  int     `long:"toChapter" description:"to chapter"`
	ToTitle                    string  `long:"toTitle" description:"to title"`
	Author                     string  `short:"a" long:"author" description:"author"`
	AutoUpdateExternalHandlers bool    `long:"autoUpdateExternalHandlers" description:"auto update external handlers"`
}

var (
	novelSiteHandlers []*NovelSiteHandler
	opts              Options
	sha1ver           string // sha1 revision used to build the program
	buildTime         string // when the executable was built
)

func registerNovelSiteHandler(handler *NovelSiteHandler) {
	novelSiteHandlers = append(novelSiteHandlers, handler)
}

func listCommandHandler() {
	fmt.Println("内建支持小说网站：")
	for _, h := range novelSiteHandlers {
		fmt.Println("\t" + h.Title + ": " + strings.Join(h.Urls, ", "))
	}
}

func runHandler(handler *NovelSiteHandler, novelURL string, ch chan bool) bool {
	if handler.Begin != nil {
		handler.Begin()
	}
	defer func() {
		if handler.End != nil {
			handler.End()
		}
	}()
	if !handler.CanHandle(novelURL) {
		return false
	}
	gen := ebook.NewBook(opts.Format)
	gen.SetPDFFontSize(opts.TitleFontSize, opts.ContentFontSize)
	gen.SetHTMLBodyFont(opts.BodyFontFamily, opts.BodyFontSize)
	gen.SetHTMLH1Font(opts.H1FontFamily, opts.H1FontSize)
	gen.SetHTMLH2Font(opts.H2FontFamily, opts.H2FontSize)
	gen.SetHTMLParaFont(opts.ParaFontFamily, opts.ParaFontSize, opts.ParaLineHeight)
	gen.SetLineSpacing(opts.LineSpacing)
	gen.PagesPerFile(opts.PagesPerFile)
	gen.ChaptersPerFile(opts.ChaptersPerFile)
	gen.SetMargins(opts.LeftMargin, opts.TopMargin)
	gen.SetPageType(opts.PageType)
	gen.SetPageSize(opts.PageWidth, opts.PageHeight)
	gen.SetFontFile(opts.FontFile)
	gen.Output(opts.OutputFile)
	if handler.PreprocessChapterListURL != nil {
		novelURL = handler.PreprocessChapterListURL(novelURL)
	}
	theURL, _ := url.Parse(novelURL)
	headers := http.Header{
		"Referer":                   []string{fmt.Sprintf("%s://%s", theURL.Scheme, theURL.Host)},
		"User-Agent":                []string{"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0"},
		"Accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"},
		"Accept-Language":           []string{`en-US,en;q=0.8`},
		"Upgrade-Insecure-Requests": []string{"1"},
	}
	rawPageContent, err := httputil.GetBytes(novelURL, headers, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
	if err != nil {
		ch <- false
		return true
	}
	title, chapters := handler.ExtractChapterList(novelURL, rawPageContent)
	if len(chapters) == 0 {
		ch <- false
		return true
	}

	gen.Info()
	gen.Begin()

	gen.SetTitle(title)
	gen.SetAuthor(opts.Author)
	dlutil := NewDownloadUtil(handler.ExtractChapterContent, gen)
	dlutil.Process()
	for _, chapter := range chapters {
		if dlutil.AddURL(chapter.Index, chapter.Title, chapter.URL) {
			break
		}
	}
	dlutil.Wait()
	gen.End()

	//handler.Download(novelURL, gen)
	fmt.Println("downloaded", novelURL)
	ch <- true
	return true
}

func downloadBook(novelURL string, ch chan bool) {
	for _, handler := range novelSiteHandlers {
		if runHandler(handler, novelURL, ch) {
			return
		}
	}

	fmt.Println("not downloaded", novelURL)
	ch <- false
}

func main() {
	luaVersion := lua.GetLuaRelease()
	luajitVersion := lua.GetLuaJITVersion()
	if luajitVersion != "" {
		luaVersion = luajitVersion
	}
	fmt.Printf("GetNovel with %s\nCommit Id: %s\nBuilt at %s\n", luaVersion, sha1ver, buildTime)
	if len(os.Args) < 2 {
		fmt.Println("使用方法：\n\tgetnovel 小说目录网址")
		listCommandHandler()
		return
	}

	opts = Options{
		InsecureSkipVerify:         false,
		Format:                     "mobi",
		List:                       false,
		LeftMargin:                 10,
		TopMargin:                  10,
		PageHeight:                 841.89,
		PageWidth:                  595.28,
		TitleFontSize:              24,
		ContentFontSize:            18,
		H1FontFamily:               "CustomFont",
		H2FontFamily:               "CustomFont",
		BodyFontFamily:             "CustomFont",
		ParaFontFamily:             "CustomFont",
		H1FontSize:                 "4em",
		H2FontSize:                 "1.2em",
		BodyFontSize:               "1.2em",
		ParaFontSize:               "1.0em",
		ParaLineHeight:             "1.0em",
		LineSpacing:                1.2,
		PagesPerFile:               0,
		ChaptersPerFile:            0,
		FontFile:                   filepath.Join("fonts", "CustomFont.ttf"),
		RetryCount:                 3,
		Timeout:                    60,
		ParallelCount:              int64(runtime.NumCPU()) * 2, // get cpu logical core number
		Author:                     "GetNovel用户",
		AutoUpdateExternalHandlers: false,
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
		if err != nil {
			log.Fatal(err)
		}
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

	httputil.SetInsecureSkipVerify(opts.InsecureSkipVerify)

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
