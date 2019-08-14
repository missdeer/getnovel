package main

import (
	"bufio"
	"bytes"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/missdeer/getnovel/ebook"
	"github.com/missdeer/golib/httputil"
	"github.com/missdeer/golib/ic"
)

var (
	httpHeadersSzzyue = http.Header{
		"Referer":                   []string{"http://www.szzyue.com/"},
		"User-Agent":                []string{"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0"},
		"Accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"},
		"Accept-Language":           []string{`en-US,en;q=0.8`},
		"Upgrade-Insecure-Requests": []string{"1"},
	}
)

func downloadSzzyuePage(u string) (c []byte) {
	var err error
	c, err = httputil.GetBytes(u, httpHeadersSzzyue, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
	if err != nil {
		return
	}
	c = ic.Convert("gbk", "utf-8", c)
	c = bytes.Replace(c, []byte("\r\n"), []byte(""), -1)
	c = bytes.Replace(c, []byte("\r"), []byte(""), -1)
	c = bytes.Replace(c, []byte("\n"), []byte(""), -1)
	leadingStr := "<dd id=\"contents\">"
	idx := bytes.Index(c, []byte(leadingStr))
	if idx > 1 {
		c = c[idx+len(leadingStr):]
	}
	endingStr := "</dd>"
	idx = bytes.Index(c, []byte(endingStr))
	if idx > 1 {
		c = c[:idx]
	}
	c = bytes.Replace(c, []byte("<br /><br />&nbsp;&nbsp;&nbsp;&nbsp;"), []byte("</p><p>"), -1)
	c = bytes.Replace(c, []byte("&nbsp;&nbsp;&nbsp;&nbsp;"), []byte(""), -1)
	c = bytes.Replace(c, []byte("<p>　　"), []byte("<p>"), -1)
	return
}

func downloadSzzyue(u string, gen ebook.IBook) {
	b, err := httputil.GetBytes(u, httpHeadersSzzyue, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
	if err != nil {
		return
	}

	b = bytes.Replace(b, []byte("<td class=\"L\">"), []byte("\n<td class=\"L\">"), -1)
	b = bytes.Replace(b, []byte("</td>"), []byte("</td>\n"), -1)

	gen.Begin()

	dlutil := newDownloadUtil(downloadSzzyuePage, gen)
	dlutil.process()

	var title string
	index := 0
	// 	<li class="zl"><a href="12954102.html">阅读指南（重要，必读）</a></li>
	r, _ := regexp.Compile(`<td class="L"><a\shref="([0-9]+\.html)">([^<]+)</a></td>`)
	// <div class="tit"><b>1号球王最新章节列表</b></div>
	re, _ := regexp.Compile(`<h1>([^<]+)</h1>`)
	scanner := bufio.NewScanner(bytes.NewReader(b))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		// convert from gbk to UTF-8
		l := ic.ConvertString("gbk", "utf-8", line)
		if title == "" {
			ss := re.FindAllStringSubmatch(l, -1)
			if len(ss) > 0 && len(ss[0]) > 0 {
				s := ss[0]
				title = s[1]
				if strings.HasSuffix(title, ` 最新章节`) {
					title = title[:len(title)-len(` 最新章节`)]
				}
				gen.SetTitle(title)
				continue
			}
		}
		if r.MatchString(l) {
			ss := r.FindAllStringSubmatch(l, -1)
			s := ss[0]
			finalURL := strings.Replace(u, "index.html", s[1], -1)
			index++
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
		Title:         `新顶点笔趣阁小说网`,
		MatchPatterns: []string{`http://www.szzyue.com/dushu/[0-9]+/[0-9]+/index.html`},
		Download:      downloadSzzyue,
	})
}
