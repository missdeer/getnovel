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
		Title: `一流吧`,
		MatchPatterns: []string{
			`http://www\.168xs\.com/du/[0-9]+/`,
		},
		Download: func(u string) {
			dlPage := func(u string) (c []byte) {
				var err error
				headers := map[string]string{
					"Referer":                   "http://www.168xs.com/",
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
				leadingStr := `<div id="BookText">`
				idx := bytes.Index(c, []byte(leadingStr))
				if idx > 1 {
					c = c[idx+len(leadingStr):]
				}
				idx = bytes.Index(c, []byte("</div>"))
				if idx > 1 {
					c = c[:idx]
				}
				c = bytes.Replace(c, []byte("<br /><br />&nbsp;&nbsp;&nbsp;&nbsp;"), []byte("</p><p>"), -1)
				c = bytes.Replace(c, []byte("&nbsp;&nbsp;&nbsp;&nbsp;"), []byte(""), -1)
				c = bytes.Replace(c, []byte(`<a href="http://www.168xs.com" target="_blank">www.168xs.com</a> 更新好快。`), []byte(""), -1)
				c = bytes.Replace(c, []byte(`（<a href="http://www.168xs.com" target="_blank">www.168xs.com</a>）`), []byte(""), -1)
				c = bytes.Replace(c, []byte(`（<a href="http://www.168xs.com" target="_blank">www.168xs.com</a>最快更新）`), []byte(""), -1)
				c = bytes.Replace(c, []byte(`（一流吧小说网<a href="http://www.168xs.com" target="_blank">www.168xs.com</a>最快更新）`), []byte(""), -1)
				c = bytes.Replace(c, []byte(`⒈⒍⒏ｘｓ．ｃｏｍ`), []byte(""), -1)
				c = bytes.Replace(c, []byte(`щщщ.７９ＸＳ.сОΜ`), []byte(""), -1)

				middleRandomStr := `重磅推荐【`
				idx = bytes.Index(c, []byte(middleRandomStr))
				if idx > 1 {
					idxEnd := bytes.Index(c[idx:], []byte("</b></a>"))
					if idxEnd > 1 {
						b := c[idx:]
						c = append(c[:idx], b[idxEnd+8:]...)
					}
				}
				return
			}
			tocURL := u
			headers := map[string]string{
				"Referer":                   "http://www.168xs.com/",
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
			b = bytes.Replace(b, []byte("</a></dd><dd><a"), []byte("</a></dd>\n<dd><a"), -1)

			gen.Begin()

			var title string
			r, _ := regexp.Compile(`<dd><a\shref="([0-9]+\.html)">([^<]+)</a></dd>$`)
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
					c := dlPage(finalURL)
					gen.AppendContent(s[2], finalURL, string(c))
					fmt.Println(s[2], finalURL, len(c), "bytes")
				}
			}
			gen.End()
		},
	})
}
