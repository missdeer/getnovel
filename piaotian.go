package main

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dfordsoft/golib/httputil"
	"github.com/dfordsoft/golib/ic"
)

func init() {
	registerNovelSiteHandler(&novelSiteHandler{
		Title: `飘天`,
		MatchPatterns: []string{
			`http://www\.piaotian\.com/html/[0-9]/[0-9]+/`,
			`http://www\.piaotian\.com/bookinfo/[0-9]/[0-9]+\.html`,
		},
		Download: func(u string) {
			dlPage := func(u string) (c []byte) {
				var err error
				headers := map[string]string{
					"Referer":                   "http://www.piaotian.com/",
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
				idx := bytes.Index(c, []byte("</tr></table><br>&nbsp;&nbsp;&nbsp;&nbsp;"))
				if idx > 1 {
					c = c[idx+17:]
				}
				idx = bytes.Index(c, []byte("</div>"))
				if idx > 1 {
					c = c[:idx]
				}
				c = bytes.Replace(c, []byte("<br /><br />&nbsp;&nbsp;&nbsp;&nbsp;"), []byte("</p><p>"), -1)
				c = bytes.Replace(c, []byte("&nbsp;&nbsp;&nbsp;&nbsp;"), []byte(""), -1)
				return
			}
			tocURL := u
			r, _ := regexp.Compile(`http://www\.piaotian\.com/bookinfo/([0-9])/([0-9]+)\.html`)
			if r.MatchString(u) {
				ss := r.FindAllStringSubmatch(u, -1)
				s := ss[0]
				tocURL = fmt.Sprintf("http://www.piaotian.com/html/%s/%s/", s[1], s[2])
			}
			fmt.Println("download book from", tocURL)

			headers := map[string]string{
				"Referer":                   "http://www.piaotian.com/",
				"User-Agent":                "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0",
				"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
				"Accept-Language":           `en-US,en;q=0.8`,
				"Upgrade-Insecure-Requests": "1",
			}
			b, err := httputil.GetBytes(tocURL, headers, 60*time.Second, 3)
			if err != nil {
				return
			}

			gen.Begin()

			var title string
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
					c := dlPage(finalURL)
					gen.AppendContent(s[2], finalURL, string(c))
					fmt.Println(s[2], finalURL, len(c), "bytes")
				}
			}
			gen.End()
		},
	})
}
