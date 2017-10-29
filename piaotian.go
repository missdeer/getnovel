package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"
)

func init() {
	registerNovelSiteHandler(&NovelSiteHandler{
		Match:    isPiaotian,
		Download: dlPiaotian,
	})
}

func isPiaotian(u string) bool {
	r, _ := regexp.Compile(`http://www\.piaotian\.com/html/[0-9]/[0-9]+/`)
	if r.MatchString(u) {
		return true
	}
	r, _ = regexp.Compile(`http://www\.piaotian\.com/bookinfo/[0-9]/[0-9]+\.html`)
	if r.MatchString(u) {
		return true
	}
	return false
}

func dlPiaotianPage(u string) (c []byte) {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	retry := 0
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Println("piaotian - Could not parse novel page request:", err)
		return
	}

	req.Header.Set("Referer", "http://www.piaotian.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("accept-language", `en-US,en;q=0.8`)
	req.Header.Set("Upgrade-Insecure-Requests", "1")
doRequest:
	resp, err := client.Do(req)
	if err != nil {
		log.Println("piaotian - Could not send novel page request:", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Println("piaotian - novel page request not 200")
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	c, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("piaotian - novel page content reading failed")
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}
	c = bytes.Replace(c, []byte(`\r\n`), []byte(""), -1)
	c = bytes.Replace(c, []byte(`\n`), []byte(""), -1)
	idx := bytes.Index(c, []byte(`</tr></table><br>&nbsp;&nbsp;&nbsp;&nbsp;`))
	if idx > 1 {
		c = c[idx+13:]
	}
	idx = bytes.Index(c, []byte(`</div>`))
	if idx > 1 {
		c = c[:idx]
	}
	c = bytes.Replace(c, []byte(`<br>&nbsp;&nbsp;&nbsp;&nbsp;`), []byte(""), -1)
	return Convert("gbk", "utf-8", c)
}

func dlPiaotian(u string) {
	toc := u
	r, _ := regexp.Compile(`http://www\.piaotian\.com/bookinfo/([0-9])/([0-9]+)\.html`)
	if r.MatchString(u) {
		ss := r.FindAllStringSubmatch(u, -1)
		s := ss[0]
		toc = fmt.Sprintf("http://www.piaotian.com/html/%s/%s/", s[1], s[2])
	}
	fmt.Println("download book from", toc)

	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	retry := 0
	req, err := http.NewRequest("GET", toc, nil)
	if err != nil {
		log.Println("piaotian - Could not parse novel request:", err)
		return
	}

	req.Header.Set("Referer", "http://www.piaotian.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("accept-language", `en-US,en;q=0.8`)
	req.Header.Set("Upgrade-Insecure-Requests", "1")
doRequest:
	resp, err := client.Do(req)
	if err != nil {
		log.Println("piaotian - Could not send novel request:", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Println("piaotian - novel request not 200")
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	r, _ = regexp.Compile(`^<li><a\shref="([0-9]+\.html)">([^<]+)</a></li>$`)
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		// convert from gbk to UTF-8
		l := ConvertString("gbk", "utf-8", line)
		if r.MatchString(l) {
			ss := r.FindAllStringSubmatch(l, -1)
			s := ss[0]
			finalURL := fmt.Sprintf("%s%s", toc, s[1])
			fmt.Println(s[2], finalURL)
			c := dlPiaotianPage(finalURL)
		}
	}
}
