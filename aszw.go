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
		Title: `爱上中文`,
		MatchPatterns: []string{
			`https://www\.aszw\.org/book/[0-9]+/[0-9]+/`,
		},
		Download: func(u string) {
			dlPage := func(u string) (c []byte) {
				var err error
				headers := map[string]string{
					"Referer":                   "http://www.aszw.org/",
					"User-Agent":                "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0",
					"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
					"Accept-Language":           `en-US,en;q=0.8`,
					"Upgrade-Insecure-Requests": "1",
				}
				c, err = httputil.GetBytes(u, headers, 60*time.Second, 3)
				if err != nil {
					return
				}

				c = ic.Convert("gbk", "utf-8", c)
				c = bytes.Replace(c, []byte("\r\n"), []byte(""), -1)
				c = bytes.Replace(c, []byte("\r"), []byte(""), -1)
				c = bytes.Replace(c, []byte("\n"), []byte(""), -1)
				leadingStr := `<div id="contents">`
				idx := bytes.Index(c, []byte(leadingStr))
				if idx > 1 {
					c = c[idx+len(leadingStr):]
				}
				idx = bytes.Index(c, []byte("</div>"))
				if idx > 1 {
					c = c[:idx]
				}
				c = bytes.Replace(c, []byte("<br /><br />&nbsp;&nbsp;&nbsp;&nbsp;"), []byte("</p><p>"), -1)
				c = bytes.Replace(c, []byte("<br/><br/>"), []byte("</p><p>"), -1)
				c = bytes.Replace(c, []byte("<br/>　　"), []byte("</p><p>"), -1)
				return
			}
			tocURL := u
			headers := map[string]string{
				"Referer":                   "http://www.aszw.org/",
				"User-Agent":                "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0",
				"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
				"Accept-Language":           `en-US,en;q=0.8`,
				"Upgrade-Insecure-Requests": "1",
			}
			b, err := httputil.GetBytes(tocURL, headers, 60*time.Second, 3)
			if err != nil {
				return
			}

			b = ic.Convert("gbk", "utf-8", b)
			b = bytes.Replace(b, []byte("<tr><td class=\"L\">"), []byte("<tr>\n<td class=\"L\">"), -1)
			b = bytes.Replace(b, []byte("</td><td class=\"L\">"), []byte("</td>\n<td class=\"L\">"), -1)

			gen.Begin()

			var title string
			// <td class="L"><a href="43118588.html">1、我会对你负责的</a></td>
			r, _ := regexp.Compile(`<td\sclass="L"><a\shref="([0-9]+\.html)">([^<]+)</a></td>$`)
			re, _ := regexp.Compile(`<h1>([^<]+)</h1>`)
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
					c := dlPage(finalURL)
					gen.AppendContent(s[2], finalURL, string(c))
					fmt.Println(s[2], finalURL, len(c), "bytes")
				}
			}
			gen.End()
		},
	})
}
