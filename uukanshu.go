package main

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/missdeer/golib/ic"
)

func extractUukanshuChapterList(u string, rawPageContent []byte) (title string, chapters []*NovelChapterInfo) {
	var lines []string
	// 	<li><a href="/b/2816/52791.html" title="调教初唐 第一千零八十五章 调教完毕……" target="_blank">第一千零八十五章 调教完毕……</a></li>
	r := regexp.MustCompile(`<li><a\shref="/b/[0-9]+/([0-9]+\.html)"\stitle="[^"]+"\starget="_blank">([^<]+)</a></li>$`)
	// <h1><a href="/b/2816/" title="调教初唐最新章节">调教初唐最新章节</a></h1>
	re := regexp.MustCompile(`<h1><a\shref="/b/[0-9]+/"\stitle="[^"]+">([^<]+)</a></h1>$`)
	scanner := bufio.NewScanner(bytes.NewReader(rawPageContent))
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
		chapters = append(chapters, &NovelChapterInfo{
			Index: index + 1,
			Title: s[2],
			URL:   finalURL,
		})
	}
	return
}

func extractUukanshuChapterContent(rawPageContent []byte) (c []byte) {
	c = ic.Convert("gbk", "utf-8", rawPageContent)
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

func init() {
	registerNovelSiteHandler(&NovelSiteHandler{
		Title: `UU看书`,
		Urls:  []string{`https://www.uukanshu.net/`},
		CanHandle: func(u string) bool {
			reg := regexp.MustCompile(`https://www\.uukanshu\.net/b/[0-9]+/`)
			return reg.MatchString(u)
		},
		ExtractChapterList:    extractUukanshuChapterList,
		ExtractChapterContent: extractUukanshuChapterContent,
	})
}
