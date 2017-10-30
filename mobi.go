package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type Mobi struct {
	title      string
	uid        int64
	count      int
	tocTmp     *os.File
	contentTmp *os.File
	navTmp     *os.File
}

var (
	contentHTMLTemplate = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">
	<html xmlns="http://www.w3.org/1999/xhtml">
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8"> 
		<title>%s</title>
		<style type="text/css">
		@font-face{
			font-family: "CustomFont";
			src: url(fonts/CustomFont.ttf);
		}
		body{
			font-family: "CustomFont";
			font-size: 1.2em;
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
	<h1 id="title">%s</h1>
	<a href="#content">跳到第一篇</a><br />%s
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

	tocNCXTemplate = `<?xml version="1.0" encoding="UTF-8"?>
	<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1" xml:lang="zh-CN">
	<head>
	<meta name="dtb:uid" content="%d" />
	<meta name="dtb:depth" content="4" />
	<meta name="dtb:totalPageCount" content="0" />
	<meta name="dtb:maxPageNumber" content="0" />
	</head>
	<docTitle><text>%s</text></docTitle>
	<docAuthor><text>类库大魔王</text></docAuthor>
	<navMap>		
		<navPoint class="book">
			<navLabel><text>%s</text></navLabel>
			<content src="content.html" />
			%s        
		</navPoint>			
	</navMap>
	</ncx>`

	contentOPFTemplate = `<?xml version="1.0" encoding="utf-8"?>
	<package xmlns="http://www.idpf.org/2007/opf" version="2.0" unique-identifier="uid">
	<metadata>
	<dc-metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:opf="http://www.idpf.org/2007/opf">
		<dc:title>%s</dc:title>
		<dc:language>zh-CN</dc:language>
		<dc:identifier id="uid">%d%s</dc:identifier>
		<dc:creator>GetNovel</dc:creator>
		<dc:publisher>类库大魔王</dc:publisher>
		<dc:subject>%s</dc:subject>
		<dc:date>%s</dc:date>
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
)

func (m *Mobi) Begin() {
	var err error
	m.tocTmp, err = os.OpenFile(`toc.tmp`, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("opening file toc.tmp for writing failed ", err)
		return
	}
	m.contentTmp, err = os.OpenFile(`content.tmp`, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("opening file content.tmp for writing failed ", err)
		return
	}
	m.navTmp, err = os.OpenFile(`nav.tmp`, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("opening file nav.tmp for writing failed ", err)
		return
	}
}

func (m *Mobi) End() {
	m.tocTmp.Close()
	m.contentTmp.Close()
	m.navTmp.Close()

	m.writeContentHTML()
	m.writeTocNCX()
	m.writeContentOPF()

	os.Remove(`toc.tmp`)
	os.Remove(`content.tmp`)
	os.Remove(`nav.tmp`)
}

func (m *Mobi) AppendContent(articleTitle, articleURL, articleContent string) {
	m.tocTmp.WriteString(fmt.Sprintf(`<li><a href="#article_%d">%s</a></li>`, m.count, articleTitle))
	m.contentTmp.WriteString(fmt.Sprintf(`<div id="article_%d" class="article">
		<h2 class="do_article_title">				  
		  <a href="%s">%s</a>				  
		</h2>				
		<div>
		<p>%s</p>
		</div>
		</div>`, m.count, articleURL, articleTitle, articleContent))
	m.navTmp.WriteString(fmt.Sprintf(`
		<navPoint class="chapter" id="%d" playOrder="1">
			<navLabel><text>%s</text></navLabel>
			<content src="content.html#article_%d" />
		</navPoint>
		`, m.count, articleTitle, m.count))

	m.count++
}

func (m *Mobi) SetTitle(title string) {
	m.title = title
}

func (m *Mobi) writeContentHTML() {
	contentHTML, err := os.OpenFile(`content.html`, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("opening file content.html for writing failed ", err)
		return
	}

	tocTmp, err := os.OpenFile(`toc.tmp`, os.O_RDONLY, 0644)
	if err != nil {
		log.Println("opening file toc.tmp for reading failed ", err)
		return
	}
	tocC, err := ioutil.ReadAll(tocTmp)
	if err != nil {
		log.Println("reading file toc.tmp failed ", err)
		return
	}
	contentTmp, err := os.OpenFile(`content.tmp`, os.O_RDONLY, 0644)
	if err != nil {
		log.Println("opening file content.tmp for reading failed ", err)
		return
	}
	contentC, err := ioutil.ReadAll(contentTmp)
	if err != nil {
		log.Println("reading file content.tmp failed ", err)
		return
	}

	contentHTML.WriteString(fmt.Sprintf(contentHTMLTemplate, m.title, m.title, time.Now().String(),
		string(tocC), string(contentC)))
	contentHTML.Close()

	tocTmp.Close()
	contentTmp.Close()
}

func (m *Mobi) writeContentOPF() {
	contentOPF, err := os.OpenFile("content.opf", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("opening file content.opf for writing failed ", err)
		return
	}
	contentOPF.WriteString(fmt.Sprintf(contentOPFTemplate,
		m.title, m.uid, time.Now().String(), m.title, time.Now().String()))
	contentOPF.Close()
}

func (m *Mobi) writeTocNCX() {
	tocNCX, err := os.OpenFile("toc.ncx", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("opening file toc.ncx for writing failed ", err)
		return
	}

	m.uid = time.Now().UnixNano()

	navTmp, err := os.OpenFile(`nav.tmp`, os.O_RDONLY, 0644)
	if err != nil {
		log.Println("opening file nav.tmp for reading failed ", err)
		return
	}
	navC, err := ioutil.ReadAll(navTmp)
	if err != nil {
		log.Println("reading file nav.tmp failed ", err)
		return
	}
	tocNCX.WriteString(fmt.Sprintf(tocNCXTemplate, m.uid, m.title, m.title, string(navC)))
	tocNCX.Close()
	navTmp.Close()
}
