package main

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"time"

	"github.com/dfordsoft/golib/httputil"
	"github.com/dfordsoft/golib/ic"
)

func init() {
	registerNovelSiteHandler(&novelSiteHandler{
		Title: `无图小说`,
		MatchPatterns: []string{
			`http://www\.wutuxs\.com/html/[0-9]/[0-9]+/`,
		},
		Download: func(u string) {
			dlPage := func(u string) (c []byte) {
				var err error
				headers := map[string]string{
					"Referer":                   "http://www.wutuxs.com/",
					"User-Agent":                "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0",
					"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
					"Accept-Language":           `en-US,en;q=0.8`,
					"Upgrade-Insecure-Requests": "1",
				}
				c, err = httputil.GetBytes(u, headers, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
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
			headers := map[string]string{
				"Referer":                   "http://www.wutuxs.com/",
				"User-Agent":                "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0",
				"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
				"Accept-Language":           `en-US,en;q=0.8`,
				"Upgrade-Insecure-Requests": "1",
			}
			b, err := httputil.GetBytes(u, headers, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
			if err != nil {
				return
			}

			gen.Begin()

			dlutil := newDownloadUtil(dlPage, gen)
			dlutil.process()

			var title string
			// 	<td class="L"><a href="/html/7/7542/5843860.html">第一章.超级网吧系统</a></td>
			r, _ := regexp.Compile(`<td class="L"><a\shref="(/html/[0-9]+/[0-9]+/[0-9]+\.html)">([^<]+)</a></td>$`)
			// <h1>系统的黑科技网吧</h1>
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
						gen.SetTitle(title)
						continue
					}
				}
				if r.MatchString(l) {
					ss := r.FindAllStringSubmatch(l, -1)
					s := ss[0]
					finalURL := fmt.Sprintf("http://www.wutuxs.com%s", s[1])
					dlutil.maxPage++
					dlutil.addURL(dlutil.maxPage, s[2], finalURL)
				}
			}
			dlutil.wait()
			gen.End()
		},
	})
}
