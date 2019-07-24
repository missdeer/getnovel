package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/missdeer/golib/ebook"
	"github.com/missdeer/golib/httputil"
)

var (
	httpHeadersYSSM = http.Header{
		"Referer":                   []string{"http://www.yssm.tv/"},
		"User-Agent":                []string{"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0"},
		"Accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"},
		"Accept-Language":           []string{`en-US,en;q=0.8`},
		"Upgrade-Insecure-Requests": []string{"1"},
	}
)

func downloadYSSMPage(u string) (c []byte) {
	var err error
	c, err = httputil.GetBytes(u, httpHeadersYSSM, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
	if err != nil {
		return
	}

	c = bytes.Replace(c, []byte("\r\n"), []byte(""), -1)
	c = bytes.Replace(c, []byte("\r"), []byte(""), -1)
	c = bytes.Replace(c, []byte("\n"), []byte(""), -1)
	leadingStr := `<div id="content">`
	idx := bytes.Index(c, []byte(leadingStr))
	if idx > 1 {
		c = c[idx+len(leadingStr):]
	}
	idx = bytes.Index(c, []byte("</div>"))
	if idx > 1 {
		c = c[:idx]
	}
	c = bytes.Replace(c, []byte("<br/>　　"), []byte("</p><p>"), -1)
	return
}

func downloadYSSM(u string, gen ebook.IBook) {
	tocURL := u
	b, err := httputil.GetBytes(tocURL, httpHeadersYSSM, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
	if err != nil {
		return
	}

	gen.Begin()

	dlutil := newDownloadUtil(downloadYSSMPage, gen)
	dlutil.process()

	var title string
	index := 0
	r, _ := regexp.Compile(`<dd>\s<a\sstyle=""=style=""\shref="([0-9]+\.html)">([^<]+)</a></dd>$`)
	re, _ := regexp.Compile(`<h1>([^<]+)</h1>$`)
	scanner := bufio.NewScanner(bytes.NewReader(b))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		l := scanner.Text()
		if title == "" {
			ss := re.FindAllStringSubmatch(l, -1)
			if len(ss) > 0 && len(ss[0]) > 0 {
				s := ss[0]
				title = s[1]
				gen.SetTitle(title)
				continue
			}
		}
		if r.MatchString(l) {
			ss := r.FindAllStringSubmatch(l, -1)
			s := ss[0]
			finalURL := fmt.Sprintf("%s%s", tocURL, s[1])
			if dlutil.addURL(index, s[2], finalURL) {
				break
			}
		}
	}
	dlutil.wait()
	gen.End()
}

func init() {
	registerNovelSiteHandler(&novelSiteHandler{
		Title: `幼狮书盟`,
		MatchPatterns: []string{
			`http://www\.yssm\.tv/uctxt/[0-9]+/[0-9]+/`,
		},
		Download: downloadYSSM,
	})
}
