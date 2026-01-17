package bs

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/missdeer/golib/httputil"
)

const maxChapterRetries = 5

// Chapter represent a chapter of a book
type Chapter struct {
	BookSourceSite string            `json:"source"`
	BookSourceInst *BookSourceV2     `json:"-"`
	Content        string            `json:"-"`
	ChapterTitle   string            `json:"title"`
	Read           bool              `json:"is_read"`
	ChapterURL     string            `json:"url"`
	Index          int               `json:"index"`
	BelongToBook   *Book
	Page           *goquery.Document `json:"-"`
	LegadoSource   *LegadoBookSource `json:"-"` // Legado source for full rule support
}

func NewChapterFromURL(chapterURL string) (*Chapter, error) {
	if chapterURL == "" {
		return nil, errors.New("no url.")
	}

	if _, err := url.ParseRequestURI(chapterURL); err != nil {
		return nil, err
	}

	c := &Chapter{
		BookSourceSite: httputil.GetHostByURL(chapterURL),
		ChapterURL:     chapterURL,
	}
	return c, nil
}

func (c Chapter) String() string {
	return fmt.Sprintf("%s( %s )", c.ChapterTitle, c.ChapterURL)
}

func (c *Chapter) findBookSourceForChapter() *BookSourceV2 {
	if c.BookSourceInst != nil {
		return c.BookSourceInst
	}
	if c.BookSourceSite == "" {
		if c.ChapterURL == "" {
			return nil
		}
		c.BookSourceSite = httputil.GetHostByURL(c.ChapterURL)
	}

	// Try to find legado source first for full rule support
	if c.LegadoSource == nil {
		// Check if book has legado source
		if c.BelongToBook != nil && c.BelongToBook.LegadoSource != nil {
			c.LegadoSource = c.BelongToBook.LegadoSource
		} else if ls := FindLegadoSourceByHost(c.BookSourceSite); ls != nil {
			c.LegadoSource = ls
		}
	}

	if bs := allBookSources.FindBookSourceByHost(c.BookSourceSite); bs != nil {
		c.BookSourceInst = bs
		return bs
	}
	return nil
}

func (c *Chapter) getChapterPage() (*goquery.Document, error) {
	if c.Page != nil {
		return c.Page, nil
	}
	bs := c.findBookSourceForChapter()
	if c.ChapterURL == "" || bs == nil {
		return nil, errors.New("can't get chapter page.")
	}
	p, err := httputil.GetPage(c.ChapterURL, bs.HTTPUserAgent)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(p)
	if err != nil {
		return nil, err
	}
	c.Page = doc
	return c.Page, nil
}

func (c *Chapter) GetContent() string {
	if c.Content != "" {
		return c.Content
	}

	// Try using legado executor for full rule support
	c.findBookSourceForChapter()
	if c.LegadoSource != nil {
		content, err := c.LegadoSource.LegadoGetChapterContent(c.ChapterURL)
		if err == nil && content != "" {
			c.Content = content
			return c.Content
		}
		// Fall back to regular parsing if legado fails
		log.Printf("Legado content fetch failed, falling back: %v", err)
	}

	c.getContentWithRetry(0)
	return c.Content
}

func (c *Chapter) getContentWithRetry(tried int) {
	doc, err := c.getChapterPage()
	if err == nil {
		_, content := ParseRules(doc, c.BookSourceInst.RuleBookContent)
		if content != "" {
			c.Content = content
			return
		}
	} else {
		if tried < maxChapterRetries {
			time.Sleep(1 * time.Second)
			c.getContentWithRetry(tried + 1)
			return
		}
		log.Printf("get content error:%s\n", err.Error())
	}
}

func (c *Chapter) GetTitle() string {
	if c.ChapterTitle != "" {
		return c.ChapterTitle
	}
	return ""
}

func (c *Chapter) GetBook() *Book {
	if c.BelongToBook != nil {
		return c.BelongToBook
	}
	return nil
}

func (c *Chapter) GetIndex() int {
	if c.Index != -1 {
		return c.Index
	}
	return -1
}
