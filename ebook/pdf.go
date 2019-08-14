package ebook

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang/freetype/truetype"
	"github.com/signintech/gopdf"
	"github.com/signintech/gopdf/fontmaker/core"
)

// Pdf generate PDF file
type pdfBook struct {
	title           string
	height          float64
	pdf             *gopdf.GoPdf
	config          *gopdf.Config
	leftMargin      float64
	topMargin       float64
	paperWidth      float64
	paperHeight     float64
	contentWidth    float64
	contentHeight   float64
	titleFontSize   float64
	contentFontSize float64
	lineSpacing     float64
	output          string
	fontFamily      string
	fontFile        string
	pageType        string
	pagesPerFile    int
	pages           int
	chaptersPerFile int
	chapters        int
	splitIndex      int
	ttf             *truetype.Font
}

// Output set the output file path
func (m *pdfBook) Output(o string) {
	m.output = o
}

// Info output self information
func (m *pdfBook) Info() {
	fmt.Println("generating PDF file...")
}

// PagesPerFile how many smaller PDF files are expected to be generated
func (m *pdfBook) PagesPerFile(n int) {
	m.pagesPerFile = n
}

// ChaptersPerFile how many smaller PDF files are expected to be generated
func (m *pdfBook) ChaptersPerFile(n int) {
	m.chaptersPerFile = n
}

// SetLineSpacing set document line spacing
func (m *pdfBook) SetLineSpacing(lineSpacing float64) {
	m.lineSpacing = lineSpacing
}

// SetFontFile set custom font file
func (m *pdfBook) SetFontFile(file string) {
	m.fontFile = file

	// check font files
	fontFd, err := os.OpenFile(m.fontFile, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatalln("can't find font file", m.fontFile, err)
		return
	}

	fontContent, err := ioutil.ReadAll(fontFd)
	fontFd.Close()
	if err != nil {
		log.Fatalln("can't read font file", err)
		return
	}

	m.ttf, err = truetype.Parse(fontContent)
	if err != nil {
		log.Fatalln("can't parse TTF font", err)
		return
	}
	m.fontFamily = m.ttf.Name(truetype.NameIDFontFamily)

	// calculate Cap Height
	var parser core.TTFParser
	err = parser.Parse(m.fontFile)
	if err != nil {
		log.Print("can't parse TTF font", err)
		return
	}

	// Measure Height
	// get  CapHeight (https://en.wikipedia.org/wiki/Cap_height)
	capHeight := float64(float64(parser.CapHeight()) * 1000.00 / float64(parser.UnitsPerEm()))
	if m.lineSpacing*1000 < capHeight {
		m.lineSpacing = capHeight / 1000
	}
}

// SetMargins set page margins
func (m *pdfBook) SetMargins(left float64, top float64) {
	m.leftMargin = left
	m.topMargin = top
	m.contentWidth = m.paperWidth - m.leftMargin*2
	m.contentHeight = m.paperHeight - m.topMargin*2
}

// SetPageSize set page size
func (m *pdfBook) SetPageSize(width float64, height float64) {
	// https://www.cl.cam.ac.uk/~mgk25/iso-paper-ps.txt
	m.config = &gopdf.Config{
		PageSize: gopdf.Rect{
			W: width,
			H: height,
		},
		Unit: gopdf.Unit_PT,
	}
	m.paperWidth = width
	m.paperHeight = height
	m.contentWidth = width - m.leftMargin*2
	m.contentHeight = height - m.topMargin*2
}

// SetPageType dummy funciton for interface
func (m *pdfBook) SetPageType(pageType string) {
	m.pageType = pageType
}

// SetFontSize dummy funciton for interface
func (m *pdfBook) SetFontSize(titleFontSize int, contentFontSize int) {
	m.titleFontSize = float64(titleFontSize)
	m.contentFontSize = float64(contentFontSize)
}

// Begin prepare book environment
func (m *pdfBook) Begin() {
	m.beginBook()
	m.newPage()
}

func (m *pdfBook) beginBook() {
	m.pdf = &gopdf.GoPdf{}
	m.pdf.Start(*m.config)
	m.pdf.SetCompressLevel(9)
	m.pdf.SetLeftMargin(m.leftMargin)
	m.pdf.SetTopMargin(m.topMargin)

	if m.fontFile != "" {
		if err := m.pdf.AddTTFFont(m.fontFamily, m.fontFile); err != nil {
			log.Println("embed font failed", err)
			return
		}
	}
}

// End generate files that kindlegen needs
func (m *pdfBook) End() {
	m.endBook()
}

func (m *pdfBook) endBook() {
	m.pdf.SetInfo(gopdf.PdfInfo{
		Title:        m.title,
		Author:       `golib/ebook/pdf 用户制作成PDF，并非一定是作品原作者`,
		Creator:      `golib/ebook/pdf，仅限个人研究学习，对其造成的所有后果，软件/库作者不承担任何责任`,
		Producer:     `golib/ebook/pdf，仅限个人研究学习，对其造成的所有后果，软件/库作者不承担任何责任`,
		Subject:      m.title,
		CreationDate: time.Now(),
	})
	if m.pagesPerFile > 0 || m.chaptersPerFile > 0 {
		m.splitIndex++
		m.pdf.WritePdf(fmt.Sprintf("%s_%s(%.4d).pdf", m.title, m.pageType, m.splitIndex))
	} else {
		if m.output == "" {
			m.output = fmt.Sprintf("%s_%s.pdf", m.title, m.pageType)
		}
		m.pdf.WritePdf(m.output)
	}
}

func (m *pdfBook) preprocessContent(content string) string {
	c := strings.Replace(content, `<br/>`, "\n", -1)
	c = strings.Replace(c, `&amp;`, `&`, -1)
	c = strings.Replace(c, `&lt;`, `<`, -1)
	c = strings.Replace(c, `&gt;`, `>`, -1)
	c = strings.Replace(c, `&quot;`, `"`, -1)
	c = strings.Replace(c, `&#39;`, `'`, -1)
	c = strings.Replace(c, `&nbsp;`, ` `, -1)
	c = strings.Replace(c, `</p><p>`, "\n", -1)
	for idx := strings.Index(c, "\n\n"); idx >= 0; idx = strings.Index(c, "\n\n") {
		c = strings.Replace(c, "\n\n", "\n", -1)
	}
	for len(c) > 0 && (c[0] == byte(' ') || c[0] == byte('\n')) {
		c = c[1:]
	}
	for len(c) > 0 && strings.HasPrefix(c, `　`) {
		c = c[len(`　`):]
	}
	return c
}

func (m *pdfBook) newPage() {
	if m.pages > 0 && m.pages == m.pagesPerFile {
		m.endBook()
		m.beginBook()
		m.pages = 0
	}
	m.pdf.AddPage()
	m.pages++
	m.height = 0
	if err := m.pdf.SetFont(m.fontFamily, "", int(m.contentFontSize)); err != nil {
		log.Println("set new page font failed", err)
	}
}

func (m *pdfBook) newChapter() {
	if m.chapters > 0 && m.chapters == m.chaptersPerFile {
		m.endBook()
		m.beginBook()
		m.chapters = 0
		m.pages = 0

		m.pdf.AddPage()
		m.pages++
		m.height = 0
	}
	m.chapters++
}

// AppendContent append book content
func (m *pdfBook) AppendContent(articleTitle, articleURL, articleContent string) {
	m.newChapter()
	if m.height+m.titleFontSize*m.lineSpacing > m.contentHeight {
		m.writePageNumber()
		m.newPage()
	}
	if err := m.pdf.SetFont(m.fontFamily, "", int(m.titleFontSize)); err != nil {
		log.Println("set title font failed", err)
	}
	m.writeTextLine(articleTitle, m.titleFontSize)
	if err := m.pdf.SetFont(m.fontFamily, "", int(m.contentFontSize)); err != nil {
		log.Println("set content font failed", err)
	}

	c := m.preprocessContent(articleContent)
	lineBreak := "\n"
	for pos := strings.Index(c, lineBreak); ; pos = strings.Index(c, lineBreak) {
		if pos <= 0 {
			if len(c) > 0 {
				m.writeText(c, m.contentFontSize)
			}
			break
		}
		t := c[:pos]
		m.writeText(t, m.contentFontSize)
		c = c[pos+len(lineBreak):]
	}
	// append a new line at the end of chapter
	if m.height+m.contentFontSize*m.lineSpacing < m.contentHeight {
		m.pdf.Br(m.contentFontSize * m.lineSpacing)
		m.height += m.contentFontSize * m.lineSpacing
	}
}

// SetTitle set book title
func (m *pdfBook) SetTitle(title string) {
	m.title = title
	m.writeCover()
	m.newPage()
}

func (m *pdfBook) writePageNumber() {
	if err := m.pdf.SetFont(m.fontFamily, "", int(m.contentFontSize/2)); err != nil {
		log.Println("set page number font failed", err)
	}
	m.pdf.SetY(m.paperHeight - m.contentFontSize/2)
	m.pdf.SetX(m.paperWidth / 2)
	if err := m.pdf.Cell(nil, strconv.Itoa(m.pages-1)); err != nil {
		log.Println("cell failed", err)
	}
}

func (m *pdfBook) writeCover() {
	titleOnCoverFontSize := 48
	if err := m.pdf.SetFont(m.fontFamily, "", titleOnCoverFontSize); err != nil {
		log.Println("set title font on cover failed", err)
	}
	m.pdf.SetY(m.contentHeight/2 - float64(titleOnCoverFontSize))
	m.pdf.SetX(m.leftMargin)
	m.writeText(m.title, float64(titleOnCoverFontSize))
	m.pdf.Br(float64(titleOnCoverFontSize) * m.lineSpacing)
	subtitleOnCoverFontSize := 20
	if err := m.pdf.SetFont(m.fontFamily, "", subtitleOnCoverFontSize); err != nil {
		log.Println("set subtitle font on cover failed", err)
	}
	m.writeText(time.Now().Format(time.RFC3339), float64(subtitleOnCoverFontSize))
}

func (m *pdfBook) writeTextLine(t string, fontSize float64) {
	if e := m.pdf.Cell(nil, t); e != nil {
		log.Println("write text line cell error:", e, t)
	}
	m.pdf.Br(fontSize * m.lineSpacing)
	m.height += fontSize * m.lineSpacing
}

func (m *pdfBook) writeText(t string, fontSize float64) {
	t = `　　` + strings.Replace(t, "	", "", -1)
	for index := 0; index < len(t); {
		r, length := utf8.DecodeRuneInString(t[index:])
		if r == utf8.RuneError {
			// fmt.Println(t, r, index)
			t = t[:index] + t[index+1:]
			continue
		}
		if m.ttf.Index(r) == 0 {
			// fmt.Println(t[index:index+length], r, length, m.ttf.Index(r))
			t = t[:index] + t[index+length:]
			continue
		}
		index += length
	}

	count := 0
	index := 0
	for {
		r, length := utf8.DecodeRuneInString(t[index:])
		if r == utf8.RuneError {
			break
		}
		count += length
		if width, _ := m.pdf.MeasureTextWidth(t[:count]); width > m.contentWidth {
			if m.height+m.contentFontSize*m.lineSpacing > m.contentHeight {
				m.writePageNumber()
				m.newPage()
			}
			count -= length
			m.writeTextLine(t[:count], m.contentFontSize)
			t = t[count:]
			index = 0
			count = 0
		} else {
			index += length
		}
	}
	if len(t) > 0 {
		if m.height+m.contentFontSize*m.lineSpacing > m.contentHeight {
			m.writePageNumber()
			m.newPage()
		}
		m.writeTextLine(t, m.contentFontSize)
	}
}
