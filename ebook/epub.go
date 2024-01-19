package ebook

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bmaupin/go-epub"
)

var (
	css = `	@font-face{
			font-family: "CustomFont";
			src: url(%CustomFontFile%);
		}
		body{
			font-family: "%s";
			font-size: %s;
			margin:0 5px;
		}

		h1{
			font-family: "%s";
			font-size:%s;
			font-weight: bold;
		}

		h2 {
			font-family: "%s";
			font-size: %s;
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
			font-family: "%s";
			font-size: %s;
			text-indent:%s;
			line-height:1.2em;
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
		}`
)

type epubBook struct {
	e              *epub.Epub
	author         string
	title          string
	fontFile       string
	output         string
	h1FontFamily   string
	h1FontSize     string
	h2FontFamily   string
	h2FontSize     string
	bodyFontFamily string
	bodyFontSize   string
	paraFontFamily string
	paraFontSize   string
	paraLineHeight string
}

// Output set the output file path
func (m *epubBook) Output(o string) {
	m.output = o
}

// PagesPerFile dummy funciton for interface
func (m *epubBook) PagesPerFile(int) {

}

// ChaptersPerFile dummy funciton for interface
func (m *epubBook) ChaptersPerFile(int) {

}

// Info output self information
func (m *epubBook) Info() {
	fmt.Println("generating epub file...")
}

// SetLineSpacing dummy funciton for interface
func (m *epubBook) SetLineSpacing(lineSpacing float64) {
}

// SetFontFile set custom font file
func (m *epubBook) SetFontFile(file string) {
	m.fontFile = file
}

// SetPageSize dummy funciton for interface
func (m *epubBook) SetPageSize(width float64, height float64) {
}

// SetMargins dummy funciton for interface
func (m *epubBook) SetMargins(left float64, top float64) {
}

// SetPageType dummy funciton for interface
func (m *epubBook) SetPageType(pageType string) {
}

// SetPDFFontSize dummy funciton for interface
func (m *epubBook) SetPDFFontSize(titleFontSize int, contentFontSize int) {
}

// SetHTMLBodyFont set body font
func (m *epubBook) SetHTMLBodyFont(family string, size string) {
	m.bodyFontFamily = family
	m.bodyFontSize = size
}

// SetHTMLH1Font set H1 font
func (m *epubBook) SetHTMLH1Font(family string, size string) {
	m.h1FontFamily = family
	m.h1FontSize = size
}

// SetHTMLH2Font set H2 font
func (m *epubBook) SetHTMLH2Font(family string, size string) {
	m.h2FontFamily = family
	m.h2FontSize = size
}

// SetHTMLParaFont set paragraph font
func (m *epubBook) SetHTMLParaFont(family string, size string, lineHeight string) {
	m.paraFontFamily = family
	m.paraFontSize = size
	m.paraLineHeight = lineHeight
}

// Begin prepare book environment
func (m *epubBook) Begin() {
	m.e = epub.NewEpub(m.title)
	m.e.SetAuthor(m.author)
	m.e.SetTitle(m.title)
	if m.fontFile != "" {
		f, err := m.e.AddFont(m.fontFile, "")
		if err != nil {
			// handle error
			log.Fatal(err)
		}
		css = strings.Replace(css, "%CustomFontFile%", strings.Replace(f, "\\", "/", -1), -1)
	}
	css = fmt.Sprintf(css, m.bodyFontFamily, m.bodyFontSize, m.h1FontFamily, m.h1FontSize,
		m.h2FontFamily, m.h2FontSize, m.paraFontFamily, m.paraFontSize, m.paraLineHeight)
	cssFd, err := os.OpenFile("style.css", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("opening file style.css for writing failed ", err)
		return
	}
	cssFd.WriteString(css)
	cssFd.Close()
	_, err = m.e.AddCSS("style.css", "")
	if err != nil {
		log.Println("adding style.css failed ", err)
		return
	}
}

// End generate epub file
func (m *epubBook) End() {
	// Write the EPUB
	if m.output == "" {
		m.output = m.title + ".epub"
	}
	err := m.e.Write(m.output)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	os.Remove("style.css")
}

// AppendContent append book content
func (m *epubBook) AppendContent(articleTitle, articleURL, articleContent string) {
	_, err := m.e.AddSection(fmt.Sprintf("<h2>%s</h2><p>%s</p>", articleTitle, articleContent), articleTitle, "", "../css/style.css")
	if err != nil {
		// handle error
		log.Fatal(err)
	}
}

// SetTitle set book title
func (m *epubBook) SetTitle(title string) {
	m.title = title
}

// SetAuthor set book author
func (m *epubBook) SetAuthor(author string) {
	m.author = author
}
