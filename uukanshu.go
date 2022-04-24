package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/missdeer/getnovel/ebook"
	"github.com/missdeer/golib/httputil"
	"github.com/missdeer/golib/ic"
)

var (
	httpHeadersUukanshu = http.Header{
		"Referer":                   []string{"https://www.uukanshu.com/"},
		"User-Agent":                []string{"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0"},
		"Accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"},
		"Accept-Language":           []string{`en-US,en;q=0.8`},
		"Upgrade-Insecure-Requests": []string{"1"},
	}
)

func downloadUukanshuPage(u string) (c []byte) {
	var err error
	c, err = httputil.GetBytes(u, httpHeadersUukanshu, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
	if err != nil {
		return
	}
	c = ic.Convert("gbk", "utf-8", c)
	c = bytes.Replace(c, []byte("\r\n"), []byte(""), -1)
	c = bytes.Replace(c, []byte("\r"), []byte(""), -1)
	c = bytes.Replace(c, []byte("\n"), []byte(""), -1)

	startStr := []byte("<div class=\"ad_content\">")
	endStr := []byte(`</div>`)
	idx := bytes.Index(c, startStr)
	if idx > 1 {
		idxEnd := bytes.Index(c[idx:], endStr)
		if idxEnd > 1 {
			b := c[idx:]
			c = b[idxEnd+len(endStr):]
		}
	}

	adStr := []byte(`<div class="ad_content"><!-- 桌面内容中2 -->`)
	idx = bytes.Index(c, adStr)
	if idx > 1 {
		idxEnd := bytes.Index(c[idx:], endStr)
		if idxEnd > 1 {
			b := c[:idx]
			c = append(b, c[idx+idxEnd+len(endStr):]...)
		}
	}

	idx = bytes.Index(c, endStr)
	if idx > 1 {
		c = c[:idx]
	}

	c = bytes.Replace(c, []byte(`</p><p>`), []byte(`<p>`), -1)
	c = bytes.Replace(c, []byte(`<br />`), []byte(`<br/>`), -1)
	c = bytes.Replace(c, []byte(`<br/><br/>　　`), []byte(`<br/>`), -1)
	c = bytes.Replace(c, []byte(`<br/><br/>`), []byte(`<br/>`), -1)
	c = bytes.Replace(c, []byte(`<br/>　　`), []byte(`<br/>`), -1)
	c = bytes.Replace(c, []byte(`<br/><br/>&nbsp;&nbsp;&nbsp;&nbsp;`), []byte(`<p>`), -1)
	c = bytes.Replace(c, []byte(`&nbsp;&nbsp;&nbsp;&nbsp;`), []byte(""), -1)
	c = bytes.Replace(c, []byte(`<p>　　`), []byte(`<p>`), -1)
	c = bytes.Replace(c, []byte(`<p>`), []byte(`</p><p>`), -1)
	return
}

func downloadUukanshu(u string, gen ebook.IBook) {
	b, err := httputil.GetBytes(u, httpHeadersUukanshu, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
	if err != nil {
		return
	}

	gen.Begin()

	dlutil := newDownloadUtil(downloadUukanshuPage, gen)
	dlutil.process()

	var title string
	var lines []string
	// 	<li><a href="/b/2816/52791.html" title="调教初唐 第一千零八十五章 调教完毕……" target="_blank">第一千零八十五章 调教完毕……</a></li>
	r, _ := regexp.Compile(`<li><a\shref="/b/[0-9]+/([0-9]+\.html)"\stitle="[^"]+"\starget="_blank">([^<]+)</a></li>$`)
	// <h1><a href="/b/2816/" title="调教初唐最新章节">调教初唐最新章节</a></h1>
	re, _ := regexp.Compile(`<h1><a\shref="/b/[0-9]+/"\stitle="[^"]+">([^<]+)</a></h1>$`)
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
				idx := strings.Index(title, `最新章节`)
				if idx > 0 {
					title = title[:idx]
				}
				gen.SetTitle(title)
				continue
			}
		}
		if r.MatchString(l) {
			lines = append([]string{l}, lines...)
		}
	}
	lines = lines[:len(lines)-1]
	for index, l := range lines {
		ss := r.FindAllStringSubmatch(l, -1)
		s := ss[0]
		finalURL := fmt.Sprintf("%s%s", u, s[1])
		if dlutil.addURL(index+1, s[2], finalURL) {
			break
		}
	}
	dlutil.wait()
	gen.End()
}

func init() {
	registerNovelSiteHandler(&novelSiteHandler{
		Title:         `UU看书`,
		MatchPatterns: []string{`https://www\.uukanshu\.com/b/[0-9]+/`},
		Download:      downloadUukanshu,
	})
}
