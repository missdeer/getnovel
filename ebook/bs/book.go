package bs

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/missdeer/golib/httputil"
)

// Book represent a book instance
type Book struct {
	Tag          string        `json:"tag"`
	Origin       string        `json:"origin"`
	Name         string        `json:"name"`
	Author       string        `json:"author"`
	BookmarkList []interface{} `json:"bookmarkList"`
	ChapterURL   string        `json:"chapterUrl"`
	// BookURL        string            `json:"book_url"`
	CoverURL         string            `json:"coverUrl"`
	Kind             string            `json:"kind"`
	LastChapter      string            `json:"lastChapter"`
	FinalRefreshDate UnixTime          `json:"finalRefreshData"` // typo here
	NoteURL          string            `json:"noteUrl"`
	Introduce        string            `json:"introduce"`
	ChapterList      []*Chapter        `json:"-"`
	BookSourceInst   *BookSourceV2     `json:"-"`
	Page             *goquery.Document `json:"-"`
}

func (b Book) String() string {
	return fmt.Sprintf("%s( %s )", b.Name, b.NoteURL)
}

// findBookSourceForBook find book source for a specified book
func (b *Book) findBookSourceForBook() *BookSourceV2 {
	if b.BookSourceInst != nil {
		return b.BookSourceInst
	}
	if b.Tag == "" {
		if b.NoteURL == "" {
			return nil
		}
		b.Tag = httputil.GetHostByURL(b.NoteURL)
	}
	if bs := allBookSources.FindBookSourceByHost(b.Tag); bs != nil {
		b.BookSourceInst = bs
		b.Origin = bs.BookSourceName
		return bs
	}
	return nil
}

// NewBookFromURL create new Book instance from URL
func NewBookFromURL(bookURL string) (*Book, error) {
	if bookURL == "" {
		return nil, errors.New("no url.")
	}
	_, err := url.ParseRequestURI(bookURL)
	if err != nil {
		return nil, err
	}
	b := &Book{
		NoteURL: bookURL,
		Tag:     httputil.GetHostByURL(bookURL),
	}
	b.GetAuthor()
	b.GetIntroduce()
	b.GetName()
	return b, nil
}

func (b *Book) getBookPage() (*goquery.Document, error) {
	if b.Page != nil {
		return b.Page, nil
	}
	if b.NoteURL == "" {
		return nil, errors.New("No valid book URL")
	}
	bs := b.findBookSourceForBook()
	if bs == nil {
		return nil, errors.New("No valid book source")
	}
	p, err := httputil.GetPage(b.NoteURL, b.findBookSourceForBook().HTTPUserAgent)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(p)
	if err != nil {
		return nil, err
	}
	b.Page = doc
	return b.Page, nil
}

func (b *Book) GetChapterURL() string {
	if b.ChapterURL != "" {
		return b.ChapterURL
	}
	doc, err := b.getBookPage()
	if err == nil {
		_, chapterURL := ParseRules(doc, b.BookSourceInst.RuleChapterURL)
		if chapterURL != "" {
			chapterURL = urlFix(chapterURL, b.Tag)
			log.Printf("chapter url is: %s", chapterURL)
			b.ChapterURL = chapterURL
			return b.ChapterURL
		}
	} else {
		log.Printf("get chapterURL error:%s\n", err.Error())
	}
	return b.NoteURL
}

func (b *Book) GetChapterList() []*Chapter {
	_ = b.UpdateChapterList(len(b.ChapterList))
	return b.ChapterList
}

func (b *Book) UpdateChapterList(startFrom int) error {
	var doc *goquery.Document
	var err error
	bs := b.findBookSourceForBook()
	if bs == nil {
		return errors.New("No valid book source")
	}

	p, err := httputil.GetPage(b.GetChapterURL(), bs.HTTPUserAgent)

	if err != nil {
		log.Printf("error while getting chapter list page: %s", err.Error())
	}
	doc, err = goquery.NewDocumentFromReader(p)
	if err != nil {
		log.Printf("error while parsing chapter list page to goquery: %s", err.Error())
	}

	if doc == nil {
		log.Printf("%s no chapterurl found.got by bookurl.", bs.BookSourceName)
		doc, err = b.getBookPage()
		if err != nil {
			return err
		}
	}
	sel, _ := ParseRules(doc, b.BookSourceInst.RuleChapterList)
	if sel == nil {
		return errors.New("empty chapter list")
	}
	sel.Each(func(i int, s *goquery.Selection) {
		if i < startFrom {
			return
		}
		_, name := ParseRules(s, b.BookSourceInst.RuleChapterName)
		_, url := ParseRules(s, b.BookSourceInst.RuleContentURL)
		url = urlFix(url, b.Tag)
		b.ChapterList = append(b.ChapterList, &Chapter{
			ChapterTitle: name,
			ChapterURL:   url,
			BelongToBook: b,
			Index:        i,
		})
	})
	return nil
}

func (b *Book) GetName() string {
	if b.Name != "" {
		return b.Name
	}
	doc, err := b.getBookPage()
	if err == nil {
		_, title := ParseRules(doc, b.BookSourceInst.RuleBookName)
		if title != "" {
			b.Name = title
		}
	} else {
		log.Printf("get title error:%s\n", err.Error())
	}
	return b.Name
}

func (b *Book) GetIntroduce() string {
	if b.Introduce != "" {
		return b.Introduce
	}
	doc, err := b.getBookPage()
	if err == nil {
		_, intro := ParseRules(doc, b.BookSourceInst.RuleIntroduce)
		if intro != "" {
			b.Introduce = intro
		}
	} else {
		log.Printf("get introduce error:%s\n", err.Error())
	}
	return b.Introduce
}

func (b *Book) GetAuthor() string {
	if b.Author != "" {
		return b.Author
	}
	doc, err := b.getBookPage()
	if err != nil {
		log.Printf("get author error:%s\n", err.Error())
		return b.Author
	}

	if _, intro := ParseRules(doc, b.BookSourceInst.RuleBookAuthor); intro != "" {
		b.Author = intro
	}
	return b.Author
}

func (b *Book) GetCoverURL() string {
	if b.CoverURL != "" {
		return b.CoverURL
	}
	doc, err := b.getBookPage()
	if err != nil {
		log.Printf("get cover error:%s\n", err.Error())
		return b.CoverURL
	}

	if _, cover := ParseRules(doc, b.BookSourceInst.RuleCoverURL); cover != "" {
		cover = urlFix(cover, b.Tag)
		b.CoverURL = cover
	}
	return b.CoverURL
}

func (b *Book) GetOrigin() string {
	if b.Origin != "" {
		return b.Origin
	}
	b.Origin = b.findBookSourceForBook().BookSourceName
	return b.Origin
}

func (b *Book) DownloadCover(coverPath string) error {
	if b.GetCoverURL() == "" {
		return errors.New("No cover found.")
	}
	res, err := http.Get(b.GetCoverURL())
	if err != nil {
		return err
	}
	f, err := os.Create(coverPath)
	if err != nil {
		return err
	}
	defer f.Close()
	io.Copy(f, res.Body)
	return nil
}
