package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/missdeer/golib/ic"
)

type Uukanshu struct {
	removeRegexp    *regexp.Regexp
	canHandleRegexp *regexp.Regexp
}

func (uu *Uukanshu) extractChapterList(u string, rawPageContent []byte) (title string, chapters []*NovelChapterInfo) {
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

func (uu *Uukanshu) extractChapterContent(rawPageContent []byte) (c []byte) {
	c = ic.Convert("gbk", "utf-8", rawPageContent)
	c = bytes.Replace(c, []byte("\r\n"), []byte(""), -1)
	c = bytes.Replace(c, []byte("\r"), []byte(""), -1)
	c = bytes.Replace(c, []byte("\n"), []byte(""), -1)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(c))
	if err != nil {
		log.Fatal(err)
	}

	divSelection := doc.Find("div#contentbox.uu_cont")

	divSelection.Find("div.ad_content").Remove()

	divHtml, err := divSelection.Html()
	if err != nil {
		log.Fatal(err)
	}

	c = []byte(divHtml)

	c = bytes.Replace(c, []byte(`</p><p>`), []byte(`<p>`), -1)
	c = bytes.Replace(c, []byte(`<br />`), []byte(`<br/>`), -1)
	c = bytes.Replace(c, []byte(`<br/><br/>　　`), []byte(`<br/>`), -1)
	c = bytes.Replace(c, []byte(`<br/><br/>`), []byte(`<br/>`), -1)
	c = bytes.Replace(c, []byte(`<br/>　　`), []byte(`<br/>`), -1)
	c = bytes.Replace(c, []byte(`<br/><br/>&nbsp;&nbsp;&nbsp;&nbsp;`), []byte(`<p>`), -1)
	c = bytes.Replace(c, []byte(`&nbsp;&nbsp;&nbsp;&nbsp;`), []byte(""), -1)
	c = bytes.Replace(c, []byte(`<p>　　`), []byte(`<p>`), -1)
	c = bytes.Replace(c, []byte(`<p>`), []byte(`</p><p>`), -1)
	c = bytes.Replace(c, []byte(`<!--MC-->`), []byte(``), -1)
	c = bytes.Replace(c, []byte(`手机用户请浏览阅读，掌上阅读更方便。`), []byte(``), -1)
	// use regexp to remove <!--flag[a-zA-Z0-9_]*--><!--MC-->
	c = uu.removeRegexp.ReplaceAll(c, []byte(""))

	return
}

func (uu *Uukanshu) preprocessChapterListURL(u string) string {
	if strings.HasSuffix(u, "#gsc.tab=0") {
		return strings.TrimSuffix(u, "#gsc.tab=0")
	}
	return u
}

func (uu *Uukanshu) canHandle(u string) bool {
	return uu.canHandleRegexp.MatchString(u)
}

func init() {
	u := &Uukanshu{
		removeRegexp:    regexp.MustCompile(`<!--flag[a-zA-Z0-9_]*-->`),
		canHandleRegexp: regexp.MustCompile(`https://www\.uukanshu\.net/b/[0-9]+/`),
	}
	registerNovelSiteHandler(&NovelSiteHandler{
		Title:                    `UU看书`,
		Urls:                     []string{`https://www.uukanshu.net/`},
		CanHandle:                u.canHandle,
		PreprocessChapterListURL: u.preprocessChapterListURL,
		ExtractChapterList:       u.extractChapterList,
		ExtractChapterContent:    u.extractChapterContent,
	})
}
