// Package ebook generate ebook files such as .mobi or it's input,
// currently only mobi is supported
package ebook

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/missdeer/golib/fsutil"
	"github.com/mozillazg/go-pinyin"
)

// mobiBook generate files that used to make a mobi file by kindlegen
type mobiBook struct {
	title        string
	uid          int64
	count        int
	output       string
	dirName      string
	fontFilePath string
	tocTmp       *os.File
	contentTmp   *os.File
	navTmp       *os.File
}

var (
	contentHTMLTemplate = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">
	<html xmlns="http://www.w3.org/1999/xhtml">
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
		<title>%s</title>
		<style type="text/css">
		@font-face{	font-family: "CustomFont";	src: url(fonts/CustomFont.ttf);	}
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
	</div>
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
		<dc:creator>GetNovel用户制作</dc:creator>
		<dc:publisher>GetNovel，仅限个人研究学习，对其造成的任何后果，软件作者不负任何责任</dc:publisher>
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

// Output set the output file path
func (m *mobiBook) Output(o string) {
	m.output = o
}

// Info output self information
func (m *mobiBook) Info() {
	fmt.Println("generating source files for mobi file, please run kindlegen to generate mobi file after this application exits...")
}

// PagesPerFile dummy funciton for interface
func (m *mobiBook) PagesPerFile(int) {

}

// ChaptersPerFile dummy funciton for interface
func (m *mobiBook) ChaptersPerFile(int) {

}

// SetPageSize dummy funciton for interface
func (m *mobiBook) SetPageSize(width float64, height float64) {
}

// SetMargins dummy funciton for interface
func (m *mobiBook) SetMargins(left float64, top float64) {

}

// SetPageType dummy funciton for interface
func (m *mobiBook) SetPageType(pageType string) {

}

// SetFontSize dummy funciton for interface
func (m *mobiBook) SetFontSize(titleFontSize int, contentFontSize int) {

}

// SetFontFile set custom font file
func (m *mobiBook) SetFontFile(file string) {
	m.fontFilePath = file
}

// SetLineSpacing dummy funciton for interface
func (m *mobiBook) SetLineSpacing(float64) {

}

// Begin prepare book environment
func (m *mobiBook) Begin() {
	if b, e := fsutil.FileExists(m.fontFilePath); e != nil || !b {
		contentHTMLTemplate = strings.Replace(contentHTMLTemplate, `@font-face{	font-family: "CustomFont";	src: url(fonts/CustomFont.ttf);	}";`, "", -1)
		contentHTMLTemplate = strings.Replace(contentHTMLTemplate, `font-family: "CustomFont";`, "", -1)
		return
	}
}

// End generate files that kindlegen needs
func (m *mobiBook) End() {
	m.tocTmp.Close()
	m.contentTmp.Close()
	m.navTmp.Close()

	m.writeContentHTML()
	m.writeTocNCX()
	m.writeContentOPF()

	os.Remove(filepath.Join(m.dirName, `toc.tmp`))
	os.Remove(filepath.Join(m.dirName, `content.tmp`))
	os.Remove(filepath.Join(m.dirName, `nav.tmp`))

	if b, e := fsutil.FileExists(m.fontFilePath); e == nil && b {
		os.Mkdir(filepath.Join(m.dirName, "fonts"), 0755)
		if runtime.GOOS == "windows" {
			if _, err := fsutil.CopyFile(m.fontFilePath, filepath.Join(m.dirName, "fonts", "CustomFont.ttf")); err != nil {
				log.Println(err)
			}
		} else {
			var err error
			fp := m.fontFilePath
			if !filepath.IsAbs(fp) {
				fp, err = filepath.Abs(m.fontFilePath)
				if err != nil {
					log.Println(err)
				}
			}

			err = os.Symlink(fp, filepath.Join(m.dirName, "fonts", "CustomFont.ttf"))
			if err != nil {
				log.Println(err)
			}
		}
	}

	kindlegen := os.Getenv(`KINDLEGEN_PATH`)
	if b, e := fsutil.FileExists(kindlegen); e != nil || !b {
		kindlegen, _ = exec.LookPath(`kindlegen`)
	}

	if b, e := fsutil.FileExists(kindlegen); e != nil || !b {
		if dir, err := filepath.Abs(filepath.Dir(os.Args[0])); err == nil {
			kindlegen = filepath.Join(dir, `kindlegen`)
		}
	}

	if b, e := fsutil.FileExists(kindlegen); e != nil || !b {
		fmt.Println(`You need to run kindlegen utility to generate the final mobi file in directory`, m.dirName)
	}

	finalName := m.dirName

	if b, e := fsutil.FileExists(kindlegen); e != nil || !b {
		fmt.Printf("For example: kindlegen -dont_append_source -c2 -o %s.mobi content.opf\n", finalName)
		return
	}
	if !filepath.IsAbs(kindlegen) {
		kindlegen, _ = filepath.Abs(kindlegen)
	}
	cmd := exec.Command(kindlegen, "-dont_append_source", "-c2", "-o", finalName+".mobi", "content.opf")
	cmd.Dir = m.dirName
	fmt.Println("Invoking kindlegen to generate", filepath.Join(m.dirName, finalName+".mobi"), "...")
	err := cmd.Run()
	if b, _ := fsutil.FileExists(filepath.Join(m.dirName, finalName+".mobi")); err != nil && !b {
		log.Println(err)
		return
	}

	if m.output != "" {
		from, err := os.Open(filepath.Join(m.dirName, finalName+".mobi"))
		if err != nil {
			log.Println(err)
			return
		}
		defer from.Close()

		to, err := os.OpenFile(m.output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Println(err)
			return
		}
		defer to.Close()

		_, err = io.Copy(to, from)
		if err != nil {
			log.Println(err)
			return
		}

		err = to.Sync()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(m.output, "is generated.")
		return
	}
	fmt.Println(filepath.Join(m.dirName, finalName+".mobi"), "is generated.")
}

// AppendContent append book content
func (m *mobiBook) AppendContent(articleTitle, articleURL, articleContent string) {
	m.tocTmp.WriteString(fmt.Sprintf(`<li><a href="#article_%d">%s</a></li>`, m.count, articleTitle))
	m.contentTmp.WriteString(fmt.Sprintf(`<div id="article_%d" class="article"><h2 class="do_article_title"><a href="%s">%s</a></h2><div><p>%s</p></div></div>`,
		m.count, articleURL, articleTitle, articleContent))
	m.navTmp.WriteString(fmt.Sprintf(`<navPoint class="chapter" id="%d" playOrder="1"><navLabel><text>%s</text></navLabel><content src="content.html#article_%d" /></navPoint>`,
		m.count, articleTitle, m.count))

	m.count++
}

// SetTitle set book title
func (m *mobiBook) SetTitle(title string) {
	m.title = title

	finalName := ""
	t := m.title
	isCJK := false
	for len(t) > 0 {
		r, size := utf8.DecodeRuneInString(t)
		if size == 1 {
			if isCJK == true {
				isCJK = false
				finalName += "-"
			}
			finalName += string(r)
		} else {
			isCJK = true
			py := pinyin.LazyPinyin(string(r), pinyin.NewArgs())
			if len(py) > 0 {
				if finalName == "" {
					finalName = py[0]
				} else {
					finalName += "-" + py[0]
				}
			}
		}
		t = t[size:]
	}
	m.dirName = finalName
	os.Mkdir(m.dirName, 0755)

	var err error
	if m.tocTmp == nil {
		m.tocTmp, err = os.OpenFile(filepath.Join(m.dirName, `toc.tmp`), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Println("opening file toc.tmp for writing failed ", err)
			return
		}
	}
	if m.contentTmp == nil {
		m.contentTmp, err = os.OpenFile(filepath.Join(m.dirName, `content.tmp`), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Println("opening file content.tmp for writing failed ", err)
			return
		}
	}
	if m.navTmp == nil {
		m.navTmp, err = os.OpenFile(filepath.Join(m.dirName, `nav.tmp`), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Println("opening file nav.tmp for writing failed ", err)
			return
		}
	}
}

func (m *mobiBook) writeContentHTML() {
	tocTmp, err := os.OpenFile(filepath.Join(m.dirName, `toc.tmp`), os.O_RDONLY, 0644)
	if err != nil {
		log.Println("opening file toc.tmp for reading failed ", err)
		return
	}
	tocC, err := ioutil.ReadAll(tocTmp)
	tocTmp.Close()
	if err != nil {
		log.Println("reading file toc.tmp failed ", err)
		return
	}

	contentTmp, err := os.OpenFile(filepath.Join(m.dirName, `content.tmp`), os.O_RDONLY, 0644)
	if err != nil {
		log.Println("opening file content.tmp for reading failed ", err)
		return
	}
	contentC, err := ioutil.ReadAll(contentTmp)
	contentTmp.Close()
	if err != nil {
		log.Println("reading file content.tmp failed ", err)
		return
	}

	contentHTML, err := os.OpenFile(filepath.Join(m.dirName, `content.html`), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("opening file content.html for writing failed ", err)
		return
	}

	contentHTML.WriteString(fmt.Sprintf(contentHTMLTemplate, m.title, m.title, time.Now().String(),
		string(tocC), string(contentC)))
	contentHTML.Close()
}

func (m *mobiBook) writeContentOPF() {
	contentOPF, err := os.OpenFile(filepath.Join(m.dirName, "content.opf"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("opening file content.opf for writing failed ", err)
		return
	}
	contentOPF.WriteString(fmt.Sprintf(contentOPFTemplate,
		m.title, m.uid, time.Now().String(), m.title, time.Now().String()))
	contentOPF.Close()
}

func (m *mobiBook) writeTocNCX() {
	tocNCX, err := os.OpenFile(filepath.Join(m.dirName, "toc.ncx"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("opening file toc.ncx for writing failed ", err)
		return
	}

	m.uid = time.Now().UnixNano()

	navTmp, err := os.OpenFile(filepath.Join(m.dirName, `nav.tmp`), os.O_RDONLY, 0644)
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
