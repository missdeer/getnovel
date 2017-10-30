package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/dfordsoft/golib/ic"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
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
	c = bytes.Replace(c, []byte("<br />&nbsp;&nbsp;&nbsp;&nbsp;"), []byte(""), -1)
	c = bytes.Replace(c, []byte("&nbsp;&nbsp;&nbsp;&nbsp;"), []byte(""), -1)
	return
}

func dlPiaotian(u string) {
	tocURL := u
	r, _ := regexp.Compile(`http://www\.piaotian\.com/bookinfo/([0-9])/([0-9]+)\.html`)
	if r.MatchString(u) {
		ss := r.FindAllStringSubmatch(u, -1)
		s := ss[0]
		tocURL = fmt.Sprintf("http://www.piaotian.com/html/%s/%s/", s[1], s[2])
	}
	fmt.Println("download book from", tocURL)

	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	retry := 0
	req, err := http.NewRequest("GET", tocURL, nil)
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

	contentHTML, err := os.OpenFile(`content.html`, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("opening file content.html for writing failed ", err)
		return
	}

	contentHTMLTemplate := `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">
	<html xmlns="http://www.w3.org/1999/xhtml">
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8"> 
		<title>Get Novel</title>
		<style type="text/css">
		@font-face{
			font-family: "CustomFont";
			src: url(fonts/CustomFont.ttf);
		}
		body{
			font-family: "CustomFont";
			font-size: 1.1em;
			margin:0 5px;
		}
	
		h1{
			font-family: "CustomFont";
			font-size:4em;
			font-weight:bold;
		}
	
		h2 {
			font-family: "CustomFont";
			font-size: 1.2em;
			font-weight: bold;
			margin:0;
		}
		a {
			color: inherit;
			text-decoration: inherit;
			cursor: default
		}
		a[href] {
			color: blue;
			text-decoration: underline;
			cursor: pointer
		}
		p{
			font-family: "CustomFont";
			text-indent:1.5em;
			line-height:1.3em;
			margin-top:0;
			margin-bottom:0;
		}
		.italic {
			font-style: italic
		}
		.do_article_title{
			line-height:1.5em;
			page-break-before: always;
		}
		#cover{
			text-align:center;
		}
		#toc{
			page-break-before: always;
		}
		#content{
			margin-top:10px;
			page-break-after: always;
		}
		</style>
	</head>
	<body>
	<div id="cover">
	<h1 id="title">Get Novel</h1>
	<a href="#content">Go straight to first item</a><br />	06/17 06:31
	</div>
	<div id="toc">
	<h2>目录</h2> 
	<ol> 
		%s
	</ol>
	</div>
	<mbp:pagebreak></mbp:pagebreak>
	<div id="content">	
	<div id="section_1" class="section">
		%s
	</div>
	</div">
	</body>
	</html>`

	var toc, content []string
	var navPoint []string
	for scanner.Scan() {
		line := scanner.Text()
		// convert from gbk to UTF-8
		l := ic.ConvertString("gbk", "utf-8", line)
		if r.MatchString(l) {
			ss := r.FindAllStringSubmatch(l, -1)
			s := ss[0]
			finalURL := fmt.Sprintf("%s%s", tocURL, s[1])
			idx := len(toc)
			toc = append(toc, fmt.Sprintf(`<li><a href="#article_%d">%s</a></li>`, idx, s[2]))
			c := dlPiaotianPage(finalURL)
			content = append(content, fmt.Sprintf(`<div id="article_%d" class="article">
				<h2 class="do_article_title">				  
				  <a href="%s">%s</a>				  
				</h2>				
				<div>
				<p>%s</p>
				</div>
				</div>`, idx, finalURL, s[2], string(c)))
			navPoint = append(navPoint, fmt.Sprintf(`
				<navPoint class="chapter" id="%d" playOrder="1">
					<navLabel><text>%s</text></navLabel>
					<content src="content.html#article_%d" />
				</navPoint>
				`, idx, s[2], idx))

			fmt.Println(s[2], finalURL, len(c), "bytes")
		}
	}
	contentHTML.WriteString(fmt.Sprintf(contentHTMLTemplate, strings.Join(toc, "\n"), strings.Join(content, "\n")))
	contentHTML.Close()

	tocNCX, err := os.OpenFile("toc.ncx", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("opening file toc.ncx for writing failed ", err)
		return
	}

	tocNCXTemplate := `<?xml version="1.0" encoding="UTF-8"?>
	<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1" xml:lang="zh-CN">
	<head>
	<meta name="dtb:uid" content="11562530804848545888" />
	<meta name="dtb:depth" content="4" />
	<meta name="dtb:totalPageCount" content="0" />
	<meta name="dtb:maxPageNumber" content="0" />
	</head>
	<docTitle><text>Get Novel</text></docTitle>
	<docAuthor><text>类库</text></docAuthor>
	<navMap>		
		<navPoint class="book">
			<navLabel><text>Get Novel</text></navLabel>
			<content src="content.html" />
			%s        
		</navPoint>			
	</navMap>
	</ncx>`

	tocNCX.WriteString(fmt.Sprintf(tocNCXTemplate, strings.Join(navPoint, "\n")))
	tocNCX.Close()

	contentOPF, err := os.OpenFile("content.opf", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("opening file content.opf for writing failed ", err)
		return
	}
	contentOPFTemplate := `<?xml version="1.0" encoding="utf-8"?>
	<package xmlns="http://www.idpf.org/2007/opf" version="2.0" unique-identifier="uid">
	<metadata>
	<dc-metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:opf="http://www.idpf.org/2007/opf">
		<dc:title>Get Novel</dc:title>
		<dc:language>zh-CN</dc:language>
		<dc:identifier id="uid">115625308048485458882013-06-16T22:31:08Z</dc:identifier>
		<dc:creator>kindlereader</dc:creator>
		<dc:publisher>kindlereader</dc:publisher>
		<dc:subject>Get Novel</dc:subject>
		<dc:date>2013-06-16T22:31:08Z</dc:date>
		<dc:description></dc:description>
	</dc-metadata>
	
	</metadata>
	<manifest>
		<item id="content" media-type="application/xhtml+xml" href="content.html"></item>
		<item id="toc" media-type="application/x-dtbncx+xml" href="toc.ncx"></item>
	</manifest>
	
	<spine toc="toc">
		<itemref idref="content"/>
	</spine>
	
	<guide>
		<reference type="start" title="start" href="content.html#content"></reference>
		<reference type="toc" title="toc" href="content.html#toc"></reference>
		<reference type="text" title="cover" href="content.html#cover"></reference>
	</guide>
	</package>
	`
	contentOPF.WriteString(contentOPFTemplate)
	contentOPF.Close()
}
