package handler

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/missdeer/getnovel/config"
	"github.com/missdeer/golib/ic"
)

func preprocessPiaotianChapterListURL(u string) string {
	reg := regexp.MustCompile(`https://www\.piaotia\.com/bookinfo/([0-9]+)/([0-9]+)\.html`)
	if reg.MatchString(u) {
		ss := reg.FindAllStringSubmatch(u, -1)
		s := ss[0]
		return fmt.Sprintf("https://www.piaotia.com/html/%s/%s/", s[1], s[2])
	}
	return u
}

func extractPiaotianChapterList(u string, rawPageContent []byte) (title string, chapters []*config.NovelChapterInfo) {
	index := 0
	r := regexp.MustCompile(`^<li><a\shref="([0-9]+\.html)">([^<]+)</a></li>$`)
	re := regexp.MustCompile(`^<h1>([^<]+)</h1>$`)
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
			ss := r.FindAllStringSubmatch(l, -1)
			s := ss[0]
			finalURL := fmt.Sprintf("%s%s", u, s[1])
			index++
			chapters = append(chapters, &config.NovelChapterInfo{
				Index: index,
				Title: s[2],
				URL:   finalURL,
			})
		}
	}
	return
}

func extractPiaotianChapterContent(u string, rawPageContent []byte) (c []byte) {
	c = ic.Convert("gbk", "utf-8", rawPageContent)
	c = bytes.Replace(c, []byte("\r\n"), []byte(""), -1)
	c = bytes.Replace(c, []byte("\r"), []byte(""), -1)
	c = bytes.Replace(c, []byte("\n"), []byte(""), -1)
	c = bytes.Replace(c, []byte(`更多更快章节请到。`), []byte(""), -1)
	c = bytes.Replace(c, []byte(`第一时间更新`), []byte(""), -1)
	c = bytes.Replace(c, []byte(`本书首发来自17K小说网，第一时间看正版内容！`), []byte(""), -1)
	c = bytes.Replace(c, []byte(`手机用户请访问http://m.Piaotian.net`), []byte(""), -1)
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

func init() {
	registerNovelSiteHandler(&config.NovelSiteHandler{
		Sites: []config.NovelSite{
			{
				Title: `飘天文学`,
				Urls:  []string{`https://www.piaotia.com/`},
			},
		},
		CanHandle: func(u string) bool {
			patterns := []string{
				`https://www\.piaotia\.com/html/[0-9]+/[0-9]+/`,
				`https://www\.piaotia\.com/bookinfo/[0-9]+/[0-9]+\.html`,
			}
			for _, pattern := range patterns {
				reg := regexp.MustCompile(pattern)
				if reg.MatchString(u) {
					return true
				}
			}
			return false
		},
		PreprocessChapterListURL: preprocessPiaotianChapterListURL,
		ExtractChapterList:       extractPiaotianChapterList,
		ExtractChapterContent:    extractPiaotianChapterContent,
	})
}
