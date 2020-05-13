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

var (
	getChapterTried int = 0
)

// Chapter represent a chapter of a book
type Chapter struct {
	BookSourceSite string      `json:"source"`
	BookSourceInst *BookSource `json:"-"`
	Content        string      `json:"-"`
	ChapterTitle   string      `json:"title"`
	Read           bool        `json:"is_read"`
	ChapterURL     string      `json:"url"`
	Index          int         `json:"index"`
	BelongToBook   *Book
	Page           *goquery.Document `json:"-"`
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

func (c *Chapter) findBookSourceForChapter() *BookSource {
	if c.BookSourceInst != nil {
		return c.BookSourceInst
	}
	if c.BookSourceSite == "" {
		if c.ChapterURL == "" {
			return nil
		}
		c.BookSourceSite = httputil.GetHostByURL(c.ChapterURL)
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
	p, err := httputil.GetPage(c.ChapterURL, c.findBookSourceForChapter().HTTPUserAgent)
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
	doc, err := c.getChapterPage()
	if err == nil {
		_, content := ParseRules(doc, c.BookSourceInst.RuleBookContent)
		if content != "" {
			// re := regexp.MustCompile("(\b)+")
			// content = re.ReplaceAllString(content, "\n    ")
			c.Content = content
			getChapterTried = 0
			return c.Content
		}
	} else {
		if getChapterTried < 5 {
			getChapterTried++
			time.Sleep(1 * time.Second)
			return c.GetContent()
		} else {
			getChapterTried = 0
		}
		log.Printf("get content error:%s\n", err.Error())
	}
	return c.Content
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
