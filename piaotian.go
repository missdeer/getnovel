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
	httpHeadersPtwxz = http.Header{
		"Referer":                   []string{"https://www.ptwxz.com/"},
		"User-Agent":                []string{"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0"},
		"Accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"},
		"Accept-Language":           []string{`en-US,en;q=0.8`},
		"Upgrade-Insecure-Requests": []string{"1"},
	}
)

func downloadPtwxzPage(u string) (c []byte) {
	var err error
	c, err = httputil.GetBytes(u, httpHeadersPtwxz, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
	if err != nil {
		return
	}
	c = ic.Convert("gbk", "utf-8", c)
	c = bytes.Replace(c, []byte("\r\n"), []byte(""), -1)
	c = bytes.Replace(c, []byte("\r"), []byte(""), -1)
	c = bytes.Replace(c, []byte("\n"), []byte(""), -1)
	c = bytes.Replace(c, []byte(`更多更快章节请到。`), []byte(""), -1)
	c = bytes.Replace(c, []byte(`第一时间更新`), []byte(""), -1)
	c = bytes.Replace(c, []byte(`本书首发来自17K小说网，第一时间看正版内容！`), []byte(""), -1)
	c = bytes.Replace(c, []byte(`手机用户请访问http://m.ptwxz.net`), []byte(""), -1)
	idx := bytes.Index(c, []byte(`&nbsp;&nbsp;&nbsp;&nbsp;`))
	if idx > 1 {
		c = c[idx:]
	} else {
		leadingStr := `</tr></table><br>`
		idx = bytes.Index(c, []byte(leadingStr))
		if idx > 1 {
			c = c[idx+len(leadingStr):]
		}
	}

	idx = bytes.Index(c, []byte("</div>"))
	if idx > 1 {
		c = c[:idx]
	}
	c = bytes.Replace(c, []byte("<br /><br />&nbsp;&nbsp;&nbsp;&nbsp;"), []byte("</p><p>"), -1)
	c = bytes.Replace(c, []byte("&nbsp;&nbsp;&nbsp;&nbsp;"), []byte(""), -1)
	return
}

func downloadPtwxz(u string, gen ebook.IBook) {
	tocURL := u
	r, _ := regexp.Compile(`https://www\.ptwxz\.com/bookinfo/([0-9]+)/([0-9]+)\.html`)
	if r.MatchString(u) {
		ss := r.FindAllStringSubmatch(u, -1)
		s := ss[0]
		tocURL = fmt.Sprintf("https://www.ptwxz.com/html/%s/%s/", s[1], s[2])
	}
	fmt.Println("download book from", tocURL)

	b, err := httputil.GetBytes(tocURL, httpHeadersPtwxz, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
	if err != nil {
		return
	}

	gen.Begin()

	dlutil := newDownloadUtil(downloadPtwxzPage, gen)
	dlutil.process()

	var title string
	index := 0
	r, _ = regexp.Compile(`^<li><a\shref="([0-9]+\.html)">([^<]+)</a></li>$`)
	re, _ := regexp.Compile(`^<h1>([^<]+)</h1>$`)
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
			ss := r.FindAllStringSubmatch(l, -1)
			s := ss[0]
			finalURL := fmt.Sprintf("%s%s", tocURL, s[1])
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
		Title: `飘天文学`,
		MatchPatterns: []string{
			`https://www\.ptwxz\.com/html/[0-9]+/[0-9]+/`,
			`https://www\.ptwxz\.com/bookinfo/[0-9]+/[0-9]+\.html`,
		},
		Download: downloadPtwxz,
	})
}
