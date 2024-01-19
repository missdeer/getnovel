// Package ebook generate ebook files such as .mobi or it's input,
// currently only mobi is supported
package ebook

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/missdeer/golib/fsutil"
	pinyin "github.com/mozillazg/go-pinyin"
)

// singleHTMLBook generate files that used to make a mobi file by kindlegen
type singleHTMLBook struct {
	author         string
	title          string
	uid            int64
	count          int
	output         string
	dirName        string
	fontFilePath   string
	h1FontFamily   string
	h1FontSize     string
	h2FontFamily   string
	h2FontSize     string
	bodyFontFamily string
	bodyFontSize   string
	paraFontFamily string
	paraFontSize   string
	paraLineHeight string
	tocTmp         *os.File
	contentTmp     *os.File
}

// Output set the output file path
func (m *singleHTMLBook) Output(o string) {
	m.output = o
}

// Info output self information
func (m *singleHTMLBook) Info() {
	fmt.Println("generating single HTML file...")
}

// PagesPerFile dummy funciton for interface
func (m *singleHTMLBook) PagesPerFile(int) {

}

// ChaptersPerFile dummy funciton for interface
func (m *singleHTMLBook) ChaptersPerFile(int) {

}

// SetPageSize dummy funciton for interface
func (m *singleHTMLBook) SetPageSize(width float64, height float64) {
}

// SetMargins dummy funciton for interface
func (m *singleHTMLBook) SetMargins(left float64, top float64) {

}

// SetPageType dummy funciton for interface
func (m *singleHTMLBook) SetPageType(pageType string) {

}

// SetPDFFontSize dummy funciton for interface
func (m *singleHTMLBook) SetPDFFontSize(titleFontSize int, contentFontSize int) {

}

// SetHTMLBodyFont set body font
func (m *singleHTMLBook) SetHTMLBodyFont(family string, size string) {
	m.bodyFontFamily = family
	m.bodyFontSize = size
}

// SetHTMLH1Font set H1 font
func (m *singleHTMLBook) SetHTMLH1Font(family string, size string) {
	m.h1FontFamily = family
	m.h1FontSize = size
}

// SetHTMLH2Font set H2 font
func (m *singleHTMLBook) SetHTMLH2Font(family string, size string) {
	m.h2FontFamily = family
	m.h2FontSize = size
}

// SetHTMLParaFont set paragraph font
func (m *singleHTMLBook) SetHTMLParaFont(family string, size string, lineHeight string) {
	m.paraFontFamily = family
	m.paraFontSize = size
	m.paraLineHeight = lineHeight
}

// SetFontFile set custom font file
func (m *singleHTMLBook) SetFontFile(file string) {
	m.fontFilePath = file
}

// SetLineSpacing dummy funciton for interface
func (m *singleHTMLBook) SetLineSpacing(float64) {

}

// Begin prepare book environment
func (m *singleHTMLBook) Begin() {
	if b, e := fsutil.FileExists(m.fontFilePath); e != nil || !b {
		contentHTMLTemplate = strings.Replace(contentHTMLTemplate, `@font-face{	font-family: "CustomFont";	src: url(fonts/CustomFont.ttf);	}";`, "", -1)
		contentHTMLTemplate = strings.Replace(contentHTMLTemplate, `font-family: "CustomFont";`, "", -1)
		return
	}
}

// End generate files that kindlegen needs
func (m *singleHTMLBook) End() {
	m.tocTmp.Close()
	m.contentTmp.Close()

	m.writeContentHTML()

	os.Remove(filepath.Join(m.dirName, `toc.tmp`))
	os.Remove(filepath.Join(m.dirName, `content.tmp`))

	fmt.Println(filepath.Join(m.dirName, m.dirName+".html"), "is generated.")
}

// AppendContent append book content
func (m *singleHTMLBook) AppendContent(articleTitle, articleURL, articleContent string) {
	m.tocTmp.WriteString(fmt.Sprintf(`<li><a href="#article_%d">%s</a></li>`, m.count, articleTitle))
	m.contentTmp.WriteString(fmt.Sprintf(`<div id="article_%d" class="article"><h2 class="do_article_title"><a href="%s">%s</a></h2><div><p>%s</p></div></div>`,
		m.count, articleURL, articleTitle, articleContent))

	m.count++
}

// SetAuthor set book author
func (m *singleHTMLBook) SetAuthor(author string) {
	m.author = author
}

// SetTitle set book title
func (m *singleHTMLBook) SetTitle(title string) {
	m.title = title

	finalName := ""
	t := m.title
	isCJK := false
	for len(t) > 0 {
		r, size := utf8.DecodeRuneInString(t)
		if size == 1 {
			if isCJK {
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
}

func (m *singleHTMLBook) writeContentHTML() {
	tocTmp, err := os.OpenFile(filepath.Join(m.dirName, `toc.tmp`), os.O_RDONLY, 0644)
	if err != nil {
		log.Println("opening file toc.tmp for reading failed ", err)
		return
	}
	tocC, err := io.ReadAll(tocTmp)
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
	contentC, err := io.ReadAll(contentTmp)
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

	contentHTML.WriteString(fmt.Sprintf(contentHTMLTemplate, m.bodyFontFamily, m.bodyFontSize, m.h1FontFamily, m.h1FontSize,
		m.h2FontFamily, m.h2FontSize, m.paraFontFamily, m.paraFontSize, m.paraLineHeight, m.title, m.title, time.Now().String(),
		string(tocC), string(contentC)))
	contentHTML.Close()
}
