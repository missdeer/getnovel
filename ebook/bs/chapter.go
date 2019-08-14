package bs

import (
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/missdeer/golib/httputil"
)

type Chapter struct {
	BookSourceSite string      `json:"source"`
	BookSourceInst *BookSource `json:"-"`
	Content        string      `json:"content"`
	ChapterTitle   string      `json:"chapter_title"`
	Read           bool        `json:"is_read"`
	ChapterURL     string      `json:"url"`
	Index          int         `json:"index"`
	BelongToBook   *Book
	Page           *goquery.Document `json:"-"`
}

func (c Chapter) String() string {
	return fmt.Sprintf("%s( %s )", c.ChapterTitle, c.ChapterURL)
}
func (c *Chapter) FromURL(chapterURL string) error {
	if chapterURL == "" {
		return errors.New("no url.")
	}
	_, err := url.ParseRequestURI(chapterURL)
	if err != nil {
		return err
	}
	c.ChapterURL = chapterURL
	c.BookSourceSite = httputil.GetHostByURL(c.ChapterURL)
	return nil
}

func (c *Chapter) GetBookSource() *BookSource {
	if c.BookSourceInst != nil {
		return c.BookSourceInst
	}
	if c.BookSourceSite == "" {
		if c.ChapterURL == "" {
			return nil
		}
		c.BookSourceSite = httputil.GetHostByURL(c.ChapterURL)
	}
	if bsItem, ok := bsCache.Get(c.BookSourceSite); ok {
		if bs, ok := bsItem.(BookSource); ok {
			c.BookSourceInst = &bs
			return &bs
		} else {
			return nil
		}
	}
	return nil
}

func (c *Chapter) getChapterPage() (*goquery.Document, error) {
	if c.Page != nil {
		return c.Page, nil
	}
	bs := c.GetBookSource()
	if c.ChapterURL != "" && bs != nil {
		p, err := httputil.GetPage(c.ChapterURL, c.GetBookSource().HTTPUserAgent)
		if err == nil {
			doc, err := goquery.NewDocumentFromReader(p)
			if err == nil {
				c.Page = doc

				return c.Page, err
			}
		}
		return nil, err
	}
	return nil, errors.New("can't get chapter page.")
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
		}
	} else {
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
