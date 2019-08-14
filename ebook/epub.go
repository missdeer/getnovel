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
			font-family: "CustomFont";
			font-size: 1.0em;
			margin:0 5px;
		}

		h1{
			font-family: "CustomFont";
			font-size:4em;
			font-weight: bold;
		}

		h2 {
			font-family: "CustomFont";
			font-size: 1.5em;
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
			font-size: 1.0em;
			text-indent:2.0em;
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
	e        *epub.Epub
	title    string
	fontFile string
	output   string
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

// SetFontSize dummy funciton for interface
func (m *epubBook) SetFontSize(titleFontSize int, contentFontSize int) {
}

// Begin prepare book environment
func (m *epubBook) Begin() {
	m.e = epub.NewEpub(m.title)
	m.e.SetAuthor(`GetNovel用户制作成epub，并非小说原作者`)
	m.e.SetTitle(m.title)
	if m.fontFile != "" {
		f, err := m.e.AddFont(m.fontFile, "")
		if err != nil {
			// handle error
			log.Fatal(err)
		}
		css = strings.Replace(css, "%CustomFontFile%", strings.Replace(f, "\\", "/", -1), -1)
	}
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
