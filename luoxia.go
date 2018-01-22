package main

import (
	"bufio"
	"bytes"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/dfordsoft/golib/httputil"
)

func init() {
	registerNovelSiteHandler(&novelSiteHandler{
		Title: `落霞`,
		MatchPatterns: []string{
			`http://www\.luoxia\.com/[a-zA-Z0-9\-]+/([a-zA-Z0-9/\-]+)?`,
		},
		Download: func(u string) {
			dlPage := func(u string) (c []byte) {
				var err error
				headers := http.Header{
					"Referer":                   []string{"http://www.luoxia.com/"},
					"User-Agent":                []string{"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0"},
					"Accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"},
					"Accept-Language":           []string{`en-US,en;q=0.8`},
					"Upgrade-Insecure-Requests": []string{"1"},
				}
				c, err = httputil.GetBytes(u, headers, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
				if err != nil {
					return
				}

				c = bytes.Replace(c, []byte("\r\n"), []byte(""), -1)
				c = bytes.Replace(c, []byte("\r"), []byte(""), -1)
				c = bytes.Replace(c, []byte("\n"), []byte(""), -1)
				leadingStr := `(adsbygoogle = window.adsbygoogle || []).push({});`
				idx := bytes.Index(c, []byte(leadingStr))
				if idx > 1 {
					c = c[idx+len(leadingStr):]
				}
				idx = bytes.Index(c, []byte(`<p>`))
				if idx > 1 {
					c = c[idx+len(`<p>`):]
				}

				middleRandomStr := `<!-- Luoxia-middle-random -->`
				idx = bytes.Index(c, []byte(middleRandomStr))
				if idx > 1 {
					idxEnd := bytes.Index(c[idx:], []byte("</div>"))
					if idxEnd > 1 {
						b := c[idx:]
						c = append(c[:idx], b[idxEnd+6:]...)
					}
				}

				endingStr := `<div class="ggad clearfix">`
				idx = bytes.Index(c, []byte(endingStr))
				if idx > 1 {
					c = c[:idx]
				}

				idx = bytes.LastIndex(c, []byte(`</p>`))
				if idx > 1 {
					c = c[:idx]
				}

				middleRandomStr = `<a href="http://www.luoxia.com`
				idx = bytes.Index(c, []byte(middleRandomStr))
				if idx > 1 {
					idxEnd := bytes.Index(c[idx:], []byte("</a>"))
					if idxEnd > 1 {
						b := c[idx:]
						c = append(c[:idx], b[idxEnd+6:]...)
					}
				}

				return
			}

			headers := http.Header{
				"Referer":                   []string{"http://www.luoxia.com/"},
				"User-Agent":                []string{"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0"},
				"Accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"},
				"Accept-Language":           []string{`en-US,en;q=0.8`},
				"Upgrade-Insecure-Requests": []string{"1"},
			}
			b, err := httputil.GetBytes(u, headers, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
			if err != nil {
				return
			}

			b = bytes.Replace(b, []byte("</li><li>"), []byte("</li>\n<li>"), -1)

			gen.Begin()

			dlutil := newDownloadUtil(dlPage, gen)
			dlutil.process()

			var title string
			index := 0
			// <li><a target="_blank" title="第二章&nbsp;破釜沉舟" href="http://www.luoxia.com/jingzhou/32741.htm">第二章&nbsp;破釜沉舟</a></li>
			r, _ := regexp.Compile(`<li><a\starget="_blank"\stitle="[^"]+"\shref="([^"]+)">([^<]+)</a></li>$`)
			// <h1>巴州往事</h1>
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
					finalURL := s[1]
					title := strings.Replace(s[2], `&nbsp;`, ` `, -1)
					index++
					if dlutil.addURL(index, title, finalURL) {
						break
					}
				}
			}
			dlutil.wait()
			gen.End()
		},
	})
}
