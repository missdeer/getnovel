package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aarzilli/golua/lua"
	"github.com/jessevdk/go-flags"
	"github.com/missdeer/getnovel/config"
	"github.com/missdeer/getnovel/ebook"
	"github.com/missdeer/getnovel/handler"
	"github.com/missdeer/golib/httputil"
)

var (
	sha1ver   string // sha1 revision used to build the program
	buildTime string // when the executable was built
)

func runHandler(handler *config.NovelSiteHandler, novelURL string, ch chan bool) bool {
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
	gen := ebook.NewBook(config.Opts.Format)
	gen.SetPDFFontSize(config.Opts.TitleFontSize, config.Opts.ContentFontSize)
	gen.SetHTMLBodyFont(config.Opts.BodyFontFamily, config.Opts.BodyFontSize)
	gen.SetHTMLH1Font(config.Opts.H1FontFamily, config.Opts.H1FontSize)
	gen.SetHTMLH2Font(config.Opts.H2FontFamily, config.Opts.H2FontSize)
	gen.SetHTMLParaFont(config.Opts.ParaFontFamily, config.Opts.ParaFontSize, config.Opts.ParaLineHeight)
	gen.SetLineSpacing(config.Opts.LineSpacing)
	gen.PagesPerFile(config.Opts.PagesPerFile)
	gen.ChaptersPerFile(config.Opts.ChaptersPerFile)
	gen.SetMargins(config.Opts.LeftMargin, config.Opts.TopMargin)
	gen.SetPageType(config.Opts.PageType)
	gen.SetPageSize(config.Opts.PageWidth, config.Opts.PageHeight)
	gen.SetFontFile(config.Opts.FontFile)
	gen.Output(config.Opts.OutputFile)
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
	rawPageContent, err := httputil.GetBytes(novelURL, headers, time.Duration(config.Opts.Timeout)*time.Second, config.Opts.RetryCount)
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
	gen.SetAuthor(config.Opts.Author)
	dlutil := NewDownloadUtil(handler.ExtractChapterContent, gen)
	dlutil.Process()
	for _, chapter := range chapters {
		if dlutil.AddURL(chapter.Index, chapter.Title, chapter.URL) {
			break
		}
	}
	dlutil.Wait()
	gen.End()

	fmt.Println("已下载", novelURL)
	ch <- true
	return true
}

func downloadBook(novelURL string, ch chan bool) {
	runHanderWrapper := func(handler *config.NovelSiteHandler) bool {
		return runHandler(handler, novelURL, ch)
	}
	if handler.RunHandler(runHanderWrapper) {
		return
	}

	fmt.Println("未下载", novelURL)
	ch <- false
}

func listenAndServe() {
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
	fmt.Println("本机IP：")
	fmt.Println(strings.Join(ips, "\n"))
	fmt.Println("启动HTTP服务器于", config.Opts.ListenAndServe)
	log.Fatal(http.ListenAndServe(config.Opts.ListenAndServe, http.FileServer(http.Dir(dir))))
}

func main() {
	luaVersion := lua.GetLuaRelease()
	luajitVersion := lua.GetLuaJITVersion()
	if luajitVersion != "" {
		luaVersion = luajitVersion
	}
	fmt.Printf("GetNovel集成%s\n提交编号：%s\n构建于%s\n\n", luaVersion, sha1ver, buildTime)
	if len(os.Args) < 2 {
		fmt.Println("使用方法：\n\tgetnovel 小说目录网址")
		handler.ListHandlers()
		return
	}

	args, err := flags.Parse(&config.Opts)
	if err != nil {
		if len(os.Args) == 2 {
			switch os.Args[1] {
			case "-h", "--help", "/help", "/?", "/h":
				return
			}
		}
		log.Fatalln("解析命令行参数失败", err)
	}

	config.ReadLocalBookSource()

	if config.Opts.List {
		handler.ListHandlers()
		return
	}

	if config.Opts.ListenAndServe != "" {
		listenAndServe()
		return
	}

	if config.Opts.ConfigFile != "" {
		if !config.ReadLocalConfigFile(&config.Opts) && !config.ReadRemotePreset(&config.Opts) {
			return
		}
	}

	httputil.SetInsecureSkipVerify(config.Opts.InsecureSkipVerify)

	downloadedChannel := make(chan bool)
	downloadCount := 0
	for _, novelURL := range args {
		_, e := url.Parse(novelURL)
		if e != nil {
			fmt.Println("无效URL", novelURL)
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
		handler.ListHandlers()
	}
}
