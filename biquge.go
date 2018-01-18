package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/dfordsoft/golib/httputil"
	"github.com/dfordsoft/golib/ic"
)

func init() {
	dl := func(u string, tocPatterns []tocPattern, pageContentMarkers []pageContentMarker) {
		dlPage := func(u string) (c []byte) {
			var err error
			theURL, _ := url.Parse(u)
			headers := map[string]string{
				"Referer":                   fmt.Sprintf("%s://%s", theURL.Scheme, theURL.Host),
				"User-Agent":                "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0",
				"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
				"Accept-Language":           `en-US,en;q=0.8`,
				"Upgrade-Insecure-Requests": "1",
			}
			c, err = httputil.GetBytes(u, headers, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
			if err != nil {
				return
			}

			if bytes.Index(c, []byte("charset=gbk")) > 0 {
				c = ic.Convert("gbk", "utf-8", c)
			}
			c = bytes.Replace(c, []byte("\r\n"), []byte(""), -1)
			c = bytes.Replace(c, []byte("\r"), []byte(""), -1)
			c = bytes.Replace(c, []byte("\n"), []byte(""), -1)
			for _, m := range pageContentMarkers {
				if theURL.Host == m.host {
					idx := bytes.Index(c, m.start)
					if idx > 1 {
						//fmt.Println("found start")
						c = c[idx+len(m.start):]
					}
					idx = bytes.Index(c, m.end)
					if idx > 1 {
						//fmt.Println("found end")
						c = c[:idx]
					}
					break
				}
			}

			c = bytes.Replace(c, []byte("<br /><br />&nbsp;&nbsp;&nbsp;&nbsp;"), []byte("</p><p>"), -1)
			c = bytes.Replace(c, []byte("<br />&nbsp;&nbsp;&nbsp;&nbsp;"), []byte("</p><p>"), -1)
			c = bytes.Replace(c, []byte("<br/><br/>"), []byte("</p><p>"), -1)
			c = bytes.Replace(c, []byte(`　　`), []byte(""), -1)
			return
		}

		theURL, _ := url.Parse(u)
		headers := map[string]string{
			"Referer":                   fmt.Sprintf("%s://%s", theURL.Scheme, theURL.Host),
			"User-Agent":                "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0",
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			"Accept-Language":           `en-US,en;q=0.8`,
			"Upgrade-Insecure-Requests": "1",
		}
		b, err := httputil.GetBytes(u, headers, time.Duration(opts.Timeout)*time.Second, opts.RetryCount)
		if err != nil {
			return
		}

		b = bytes.Replace(b, []byte("<dd>"), []byte("\n<dd>"), -1)
		b = bytes.Replace(b, []byte("</dd>"), []byte("</dd>\n"), -1)
		if bytes.Index(b, []byte("charset=gbk")) > 0 {
			b = ic.Convert("gbk", "utf-8", b)
		}

		gen.Begin()

		dlutil := newDownloadUtil(dlPage, gen)
		dlutil.process()

		var title string
		var lines []string

		var p tocPattern
		for _, patt := range tocPatterns {
			if theURL.Host == patt.host {
				p = patt
				break
			}
		}
		r, _ := regexp.Compile(p.item)
		re, _ := regexp.Compile(p.bookTitle)
		scanner := bufio.NewScanner(bytes.NewReader(b))
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			line := scanner.Text()
			if title == "" {
				ss := re.FindAllStringSubmatch(line, -1)
				if len(ss) > 0 && len(ss[0]) > 0 {
					s := ss[0]
					title = s[p.bookTitlePos]
					gen.SetTitle(title)
					continue
				}
			}
			if r.MatchString(line) {
				lines = append(lines, line)
			}
		}

		for i := len(lines) - 1; i >= 0 && i < len(lines) && lines[0] == lines[i]; i -= 2 {
			lines = lines[1:]
		}

		for index, line := range lines {
			ss := r.FindAllStringSubmatch(line, -1)
			s := ss[0]
			articleURL := s[p.articleURLPos]
			finalURL := fmt.Sprintf("%s://%s%s", theURL.Scheme, theURL.Host, articleURL)
			if articleURL[0] != '/' {
				finalURL = fmt.Sprintf("%s%s", u, articleURL)
			}
			if strings.HasPrefix(articleURL, "http") {
				finalURL = articleURL
			}

			dlutil.addURL(index+1, s[p.articleTitlePos], finalURL)
		}
		dlutil.wait()
		gen.End()
	}
	registerNovelSiteHandler(&novelSiteHandler{
		Title:         `八一中文网`,
		MatchPatterns: []string{`https://www\.zwdu\.com/book/[0-9]+/`},
		Download: func(u string) {
			tocPatterns := []tocPattern{
				{
					host:            "www.zwdu.com",
					bookTitle:       `<h1>([^<]+)</h1>$`,
					bookTitlePos:    1,
					item:            `<dd>\s*<a\s+href="([^"]+)"(\sclass="empty")?>([^<]+)</a></dd>$`,
					articleURLPos:   1,
					articleTitlePos: 3,
					isAbsoluteURL:   true,
				},
			}
			pageContentMarkers := []pageContentMarker{
				{
					host:  "www.zwdu.com",
					start: []byte(`<div id="content">`),
					end:   []byte(`</div>`),
				},
			}
			dl(u, tocPatterns, pageContentMarkers)
		},
	})

	registerNovelSiteHandler(&novelSiteHandler{
		Title:         `少年文学网`,
		MatchPatterns: []string{`https://www\.snwx8\.com/book/[0-9]+/[0-9]+/`},
		Download: func(u string) {
			tocPatterns := []tocPattern{
				{
					host:            "www.snwx8.com",
					bookTitle:       `<h1>([^<]+)</h1>$`,
					bookTitlePos:    1,
					item:            `<dd>\s*<a\s+href="([^"]+)"\s+title="[^"]+">([^<]+)</a></dd>$`,
					articleURLPos:   1,
					articleTitlePos: 2,
				},
			}
			pageContentMarkers := []pageContentMarker{
				{
					host:  "www.snwx8.com",
					start: []byte(`<div id="BookText">`),
					end:   []byte(`</div>`),
				},
			}
			dl(u, tocPatterns, pageContentMarkers)
		},
	})
	registerNovelSiteHandler(&novelSiteHandler{
		Title:         `大海中文`,
		MatchPatterns: []string{`https://www\.dhzw\.org/book/[0-9]+/[0-9]+/`},
		Download: func(u string) {
			tocPatterns := []tocPattern{
				{
					host:            "www.dhzw.org",
					bookTitle:       `<h1>([^<]+)</h1>$`,
					bookTitlePos:    1,
					item:            `<dd>\s*<a\s+href="([^"]+)"\s+title="[^"]+">([^<]+)</a></dd>$`,
					articleURLPos:   1,
					articleTitlePos: 2,
				},
			}
			pageContentMarkers := []pageContentMarker{
				{
					host:  "www.dhzw.org",
					start: []byte(`<div id="BookText">`),
					end:   []byte(`</div>`),
				},
			}
			dl(u, tocPatterns, pageContentMarkers)
		},
	})
	registerNovelSiteHandler(&novelSiteHandler{
		Title:         `手牵手`,
		MatchPatterns: []string{`https://www\.sqsxs\.com/book/[0-9]+/[0-9]+/`},
		Download: func(u string) {
			tocPatterns := []tocPattern{
				{
					host:            "www.sqsxs.com",
					bookTitle:       `<h1>([^<]+)</h1>$`,
					bookTitlePos:    1,
					item:            `<dd>\s*<a\s+href="([^"]+)"\s+class="f-green">([^<]+)</a></a></dd>$`,
					articleURLPos:   1,
					articleTitlePos: 2,
				},
			}
			pageContentMarkers := []pageContentMarker{
				{
					host:  "www.sqsxs.com",
					start: []byte(`<div id="content">`),
					end:   []byte(`</div>`),
				},
			}
			dl(u, tocPatterns, pageContentMarkers)
		},
	})
	registerNovelSiteHandler(&novelSiteHandler{
		Title:         `燃文小说`,
		MatchPatterns: []string{`http://www\.ranwena\.com/files/article/[0-9]+/[0-9]+/`},
		Download: func(u string) {
			tocPatterns := []tocPattern{
				{
					host:            "www.ranwena.com",
					bookTitle:       `<h1>([^<]+)</h1>$`,
					bookTitlePos:    1,
					item:            `<dd>\s*<a\s*href="([^"]+)">([^<]+)</a></dd>$`,
					articleURLPos:   1,
					articleTitlePos: 2,
					isAbsoluteURL:   true,
				},
			}
			pageContentMarkers := []pageContentMarker{
				{
					host:  "www.ranwena.com",
					start: []byte(`<div id="content">`),
					end:   []byte(`</div>`),
				},
			}
			dl(u, tocPatterns, pageContentMarkers)
		},
	})
	registerNovelSiteHandler(&novelSiteHandler{
		Title: `笔趣阁系列`,
		MatchPatterns: []string{
			`http://www\.biqudu\.com/[0-9]+_[0-9]+/`,
			`http://www\.biquge\.cm/[0-9]+/[0-9]+/`,
			`http://www\.qu\.la/book/[0-9]+/`,
			`http://www\.biqugezw\.com/[0-9]+_[0-9]+/`,
			`http://www\.630zw\.com/[0-9]+_[0-9]+/`,
			`http://www\.biquge\.lu/book/[0-9]+/`,
			`http://www\.biquge5200\.com/[0-9]+_[0-9]+/`,
			`http://www\.xxbiquge\.com/[0-9]+_[0-9]+/`,
			`http://www\.biqugev\.com/[0-9]+_[0-9]+/`,
		},
		Download: func(u string) {
			tocPatterns := []tocPattern{
				{
					host:            "www.biqudu.com",
					bookTitle:       `<h1>([^<]+)</h1>$`,
					bookTitlePos:    1,
					item:            `<dd>\s*<a\s*href="([^"]+)">([^<]+)</a></dd>$`,
					articleURLPos:   1,
					articleTitlePos: 2,
				},
				{
					host:            "www.biquge.cm",
					bookTitle:       `<h1>([^<]+)</h1>$`,
					bookTitlePos:    1,
					item:            `<dd>\s*<a\s*href="([^"]+)">([^<]+)</a></dd>$`,
					articleURLPos:   1,
					articleTitlePos: 2,
				},
				{
					host:            "www.qu.la",
					bookTitle:       `<h1>([^<]+)</h1>$`,
					bookTitlePos:    1,
					item:            `<dd>\s*<a\s*(style=""\s*)?href="([^"]+)">([^<]+)</a></dd>$`,
					articleURLPos:   2,
					articleTitlePos: 3,
				},
				{
					host:            "www.biqugezw.com",
					bookTitle:       `<h1>([^<]+)</h1>$`,
					bookTitlePos:    1,
					item:            `<dd>\s*<a\s*href="([^"]+)">([^<]+)</a></dd>$`,
					articleURLPos:   1,
					articleTitlePos: 2,
				},
				{
					host:            "www.630zw.com",
					bookTitle:       `<h1>([^<]+)</h1>$`,
					bookTitlePos:    1,
					item:            `<dd>\s*<a\s*href="([^"]+)">([^<]+)</a></dd>$`,
					articleURLPos:   1,
					articleTitlePos: 2,
				},
				{
					host:            "www.biquge.lu",
					bookTitle:       `<h2>([^<]+)</h2>$`,
					bookTitlePos:    1,
					item:            `<dd>\s*<a\s*href="([^"]+)">([^<]+)</a></dd>$`,
					articleURLPos:   1,
					articleTitlePos: 2,
				},
				{
					host:            "www.biquge5200.com",
					bookTitle:       `<h1>([^<]+)</h1>$`,
					bookTitlePos:    1,
					item:            `<dd>\s*<a\s*href="([^"]+)">([^<]+)</a></dd>$`,
					articleURLPos:   1,
					articleTitlePos: 2,
					isAbsoluteURL:   true,
				},
				{
					host:            "www.xxbiquge.com",
					bookTitle:       `^<h1>([^<]+)</h1>$`,
					bookTitlePos:    1,
					item:            `<dd>\s*<a\s*href="([^"]+)"(\sclass="empty")?>([^<]+)</a></dd>$`,
					articleURLPos:   1,
					articleTitlePos: 3,
					isAbsoluteURL:   true,
				},
				{
					host:            "www.biqugev.com",
					bookTitle:       `^<h1>([^<]+)</h1>$`,
					bookTitlePos:    1,
					item:            `<dd>\s*<a\s*href="([^"]+)"(\sclass="empty")?>([^<]+)</a></dd>$`,
					articleURLPos:   1,
					articleTitlePos: 3,
				},
			}
			pageContentMarkers := []pageContentMarker{
				{
					host:  "www.biqudu.com",
					start: []byte(`<div id="content"><script>readx();</script>`),
					end:   []byte(`<script>chaptererror();</script>`),
				},
				{
					host:  "www.biquge.cm",
					start: []byte(`<div id="content">&nbsp;&nbsp;&nbsp;&nbsp;`),
					end:   []byte(`找本站搜索"笔趣阁CM" 或输入网址:www.biquge.cm</div>`),
				},
				{
					host:  "www.qu.la",
					start: []byte(`<div id="content">`),
					end:   []byte(`<script>chaptererror();</script>`),
				},
				{
					host:  "www.biqugezw.com",
					start: []byte(`<div id="content">&nbsp;&nbsp;&nbsp;&nbsp;一秒记住【笔趣阁中文网<a href="http://www.biqugezw.com" target="_blank">www.biqugezw.com</a>】，为您提供精彩小说阅读。`),
					end:   []byte(`手机用户请浏览m.biqugezw.com阅读，更优质的阅读体验。</div>`),
				},
				{
					host:  "www.630zw.com",
					start: []byte(`<div id="content">&nbsp;&nbsp;&nbsp;&nbsp;`),
					end:   []byte(`(新笔趣阁：biqugee.cc，手机笔趣阁 m.biqugee.cc )</div>`),
				},
				{
					host:  "www.biquge.lu",
					start: []byte(`<div id="content" class="showtxt">&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;`),
					end:   []byte(`请记住本书首发域名：www.biquge.lu。笔趣阁手机版阅读网址：m.biquge.lu</div>`),
				},
				{
					host:  "www.biquge5200.com",
					start: []byte(`<div id="content">`),
					end:   []byte(`</div>`),
				},
				{
					host:  "www.xxbiquge.com",
					start: []byte(`<div id="content">`),
					end:   []byte(`</div>`),
				},
				{
					host:  "www.biqugev.com",
					start: []byte(`<div id="content">`),
					end:   []byte(`</div>`),
				},
			}
			dl(u, tocPatterns, pageContentMarkers)
		},
	})
}
