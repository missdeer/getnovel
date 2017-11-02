package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/dfordsoft/golib/ebook"
	"github.com/dfordsoft/golib/ic"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func init() {
	registerNovelSiteHandler(&NovelSiteHandler{
		Match:    isUUKanshu,
		Download: dlUUKanshu,
	})
}

func isUUKanshu(u string) bool {
	r, _ := regexp.Compile(`http://www\.uukanshu\.net/b/[0-9]+/`)
	if r.MatchString(u) {
		return true
	}
	return false
}

func dlUUKanshu(u string) {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	retry := 0
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Println("uukanshu - Could not parse novel request:", err)
		return
	}

	req.Header.Set("Referer", "http://www.uukanshu.net/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("accept-language", `en-US,en;q=0.8`)
	req.Header.Set("Upgrade-Insecure-Requests", "1")
doRequest:
	resp, err := client.Do(req)
	if err != nil {
		log.Println("uukanshu - Could not send novel request:", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Println("uukanshu - novel request not 200")
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("uukanshu - Reading response body failed", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	mobi := &ebook.Mobi{}
	mobi.Begin()

	var title string
	var lines []string
	// 	<li><a href="/b/2816/52791.html" title="调教初唐 第一千零八十五章 调教完毕……" target="_blank">第一千零八十五章 调教完毕……</a></li>
	r, _ := regexp.Compile(`<li><a\shref="/b/[0-9]+/([0-9]+\.html)"\stitle="[^"]+"\starget="_blank">([^<]+)</a></li>$`)
	scanner := bufio.NewScanner(bytes.NewReader(b))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		// convert from gbk to UTF-8
		l := ic.ConvertString("gbk", "utf-8", line)
		if title == "" {
			// <h1><a href="/b/2816/" title="调教初唐最新章节">调教初唐最新章节</a></h1>
			re, _ := regexp.Compile(`<h1><a\shref="/b/[0-9]+/"\stitle="[^"]+">([^<]+)</a></h1>$`)
			ss := re.FindAllStringSubmatch(l, -1)
			if len(ss) > 0 && len(ss[0]) > 0 {
				s := ss[0]
				title = s[1]
				idx := strings.Index(title, `最新章节`)
				if idx > 0 {
					title = title[:idx]
				}
				mobi.SetTitle(title)
				continue
			}
		}
		if r.MatchString(l) {
			lines = append([]string{l}, lines...)
		}
	}
	lines = lines[:len(lines)-1]
	for _, l := range lines {
		ss := r.FindAllStringSubmatch(l, -1)
		s := ss[0]
		finalURL := fmt.Sprintf("%s%s", u, s[1])
		c := dlUUKanshuPage(finalURL)
		mobi.AppendContent(s[2], finalURL, string(c))
		fmt.Println(s[2], finalURL, len(c), "bytes")
	}
	mobi.End()
}

func dlUUKanshuPage(u string) (c []byte) {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	retry := 0
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Println("uukanshu - Could not parse novel page request:", err)
		return
	}

	req.Header.Set("Referer", "http://www.uukanshu.net/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("accept-language", `en-US,en;q=0.8`)
	req.Header.Set("Upgrade-Insecure-Requests", "1")
doRequest:
	resp, err := client.Do(req)
	if err != nil {
		log.Println("uukanshu - Could not send novel page request:", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Println("uukanshu - novel page request not 200")
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	c, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("uukanshu - novel page content reading failed")
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}
	c = ic.Convert("gbk", "utf-8", c)
	c = bytes.Replace(c, []byte("\r\n"), []byte(""), -1)
	c = bytes.Replace(c, []byte("\r"), []byte(""), -1)
	c = bytes.Replace(c, []byte("\n"), []byte(""), -1)
	idx := bytes.Index(c, []byte("<!-- 桌面内容顶部 -->"))
	if idx > 1 {
		c = c[idx:]
	}
	idx = bytes.Index(c, []byte(`</div>`))
	if idx > 1 {
		c = c[idx+6:]
	}
	startStr := []byte("<div class=\"ad_content\">")
	idx = bytes.Index(c, startStr)
	if idx > 1 {
		idxEnd := bytes.Index(c[idx:], []byte("</div>"))
		if idxEnd > 1 {
			b := c[idx:]
			c = append(c[:idx], b[idxEnd+6:]...)
		}
	}
	idx = bytes.Index(c, []byte("</div>"))
	if idx > 1 {
		c = c[:idx]
	}
	c = bytes.Replace(c, []byte("<br /><br />&nbsp;&nbsp;&nbsp;&nbsp;"), []byte("</p><p>"), -1)
	c = bytes.Replace(c, []byte("&nbsp;&nbsp;&nbsp;&nbsp;"), []byte(""), -1)
	c = bytes.Replace(c, []byte("<p>　　"), []byte("<p>"), -1)
	return
}
