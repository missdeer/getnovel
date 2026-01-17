package bs

import (
	"github.com/missdeer/getnovel/legado"
)

// LegadoBookSource wraps legado.BookSource with an executor for full rule support
type LegadoBookSource struct {
	Source   *legado.BookSource
	Executor *legado.Executor
}

// NewLegadoBookSource creates a new LegadoBookSource from legado.BookSource
func NewLegadoBookSource(source *legado.BookSource) *LegadoBookSource {
	return &LegadoBookSource{
		Source:   source,
		Executor: legado.NewExecutor(source),
	}
}

// ConvertLegadoToV2 converts legado.BookSource to BookSourceV2
// Note: This is a basic conversion and may lose some advanced rule features
func ConvertLegadoToV2(ls *legado.BookSource) BookSourceV2 {
	return BookSourceV2{
		BookSourceGroup:       ls.BookSourceGroup,
		BookSourceName:        ls.BookSourceName,
		BookSourceURL:         ls.BookSourceURL,
		Enable:                ls.Enabled,
		HTTPUserAgent:         "", // legado uses Header field
		RuleBookAuthor:        ls.RuleBookInfo.Author,
		RuleBookContent:       ls.RuleContent.Content,
		RuleBookName:          ls.RuleBookInfo.Name,
		RuleChapterList:       ls.RuleTOC.ChapterList,
		RuleChapterName:       ls.RuleTOC.ChapterName,
		RuleChapterURL:        ls.RuleBookInfo.TOCURL,
		RuleContentURL:        ls.RuleTOC.ChapterURL,
		RuleCoverURL:          ls.RuleBookInfo.CoverURL,
		RuleIntroduce:         ls.RuleBookInfo.Intro,
		RuleSearchAuthor:      ls.RuleSearch.Author,
		RuleSearchCoverURL:    ls.RuleSearch.CoverURL,
		RuleSearchKind:        ls.RuleSearch.Kind,
		RuleSearchLastChapter: ls.RuleSearch.LastChapter,
		RuleSearchList:        ls.RuleSearch.BookList,
		RuleSearchName:        ls.RuleSearch.Name,
		RuleSearchNoteURL:     ls.RuleSearch.BookURL,
		RuleSearchURL:         ls.SearchURL,
	}
}

// ConvertV2ToLegado converts BookSourceV2 to legado.BookSource
func ConvertV2ToLegado(bs2 *BookSourceV2) *legado.BookSource {
	return &legado.BookSource{
		BookSourceGroup: bs2.BookSourceGroup,
		BookSourceName:  bs2.BookSourceName,
		BookSourceURL:   bs2.BookSourceURL,
		Enabled:         bs2.Enable,
		SearchURL:       bs2.RuleSearchURL,
		RuleBookInfo: legado.RuleBookInfo{
			Author:   bs2.RuleBookAuthor,
			Name:     bs2.RuleBookName,
			CoverURL: bs2.RuleCoverURL,
			Intro:    bs2.RuleIntroduce,
			TOCURL:   bs2.RuleChapterURL,
		},
		RuleContent: legado.RuleContent{
			Content: bs2.RuleBookContent,
		},
		RuleSearch: legado.RuleSearch{
			Author:      bs2.RuleSearchAuthor,
			BookList:    bs2.RuleSearchList,
			BookURL:     bs2.RuleSearchNoteURL,
			CoverURL:    bs2.RuleSearchCoverURL,
			Kind:        bs2.RuleSearchKind,
			LastChapter: bs2.RuleSearchLastChapter,
			Name:        bs2.RuleSearchName,
		},
		RuleTOC: legado.RuleTOC{
			ChapterList: bs2.RuleChapterList,
			ChapterName: bs2.RuleChapterName,
			ChapterURL:  bs2.RuleContentURL,
		},
	}
}

// ConvertV3ToLegado converts BookSourceV3 to legado.BookSource
func ConvertV3ToLegado(bs3 *BookSourceV3) *legado.BookSource {
	return &legado.BookSource{
		BookSourceGroup: bs3.BookSourceGroup,
		BookSourceName:  bs3.BookSourceName,
		BookSourceURL:   bs3.BookSourceURL,
		Enabled:         bs3.Enable,
		Header:          bs3.Header,
		SearchURL:       bs3.SearchURL,
		Weight:          bs3.Weight,
		RuleBookInfo: legado.RuleBookInfo{
			Author:      bs3.RuleBookInfo.Author,
			CoverURL:    bs3.RuleBookInfo.CoverURL,
			Intro:       bs3.RuleBookInfo.Intro,
			Kind:        bs3.RuleBookInfo.Kind,
			LastChapter: bs3.RuleBookInfo.LastChapter,
			Name:        bs3.RuleBookInfo.Name,
			TOCURL:      bs3.RuleBookInfo.TOCURL,
		},
		RuleContent: legado.RuleContent{
			Content:        bs3.RuleContent.Content,
			NextContentURL: bs3.RuleContent.NextContentURL,
			SourceRegex:    bs3.RuleContent.SourceRegex,
		},
		RuleExplore: legado.RuleExplore{
			Author:      bs3.RuleExplore.Author,
			BookList:    bs3.RuleExplore.BookList,
			BookURL:     bs3.RuleExplore.BookURL,
			CoverURL:    bs3.RuleExplore.CoverURL,
			Intro:       bs3.RuleExplore.Intro,
			Kind:        bs3.RuleExplore.Kind,
			LastChapter: bs3.RuleExplore.LastChapter,
			Name:        bs3.RuleExplore.Name,
		},
		RuleSearch: legado.RuleSearch{
			Author:      bs3.RuleSearch.Author,
			BookList:    bs3.RuleSearch.BookList,
			BookURL:     bs3.RuleSearch.BookURL,
			CoverURL:    bs3.RuleSearch.CoverURL,
			Intro:       bs3.RuleSearch.Intro,
			Kind:        bs3.RuleSearch.Kind,
			LastChapter: bs3.RuleSearch.LastChapter,
			Name:        bs3.RuleSearch.Name,
		},
		RuleTOC: legado.RuleTOC{
			ChapterList: bs3.RuleTOC.ChapterList,
			ChapterName: bs3.RuleTOC.ChapterName,
			ChapterURL:  bs3.RuleTOC.ChapterURL,
		},
	}
}

// LegadoSearchBook searches for books using the legado executor
func (ls *LegadoBookSource) LegadoSearchBook(keyword string, page int) ([]*Book, error) {
	results, err := ls.Executor.Search(keyword, page)
	if err != nil {
		return nil, err
	}

	var books []*Book
	for _, result := range results {
		book := &Book{
			Tag:         ls.Source.BookSourceURL,
			Name:        result.Name,
			Author:      result.Author,
			Kind:        result.Kind,
			CoverURL:    result.CoverURL,
			LastChapter: result.LastChapter,
			NoteURL:     result.BookURL,
			Introduce:   result.Intro,
		}
		books = append(books, book)
	}
	return books, nil
}

// LegadoGetBookInfo gets book information using the legado executor
func (ls *LegadoBookSource) LegadoGetBookInfo(bookURL string) (*Book, error) {
	info, err := ls.Executor.GetBookInfo(bookURL)
	if err != nil {
		return nil, err
	}

	return &Book{
		Tag:         ls.Source.BookSourceURL,
		Name:        info.Name,
		Author:      info.Author,
		Kind:        info.Kind,
		CoverURL:    info.CoverURL,
		LastChapter: info.LastChapter,
		NoteURL:     bookURL,
		ChapterURL:  info.TOCURL,
		Introduce:   info.Intro,
	}, nil
}

// LegadoGetChapterList gets chapter list using the legado executor
func (ls *LegadoBookSource) LegadoGetChapterList(tocURL string) ([]*Chapter, error) {
	chapters, err := ls.Executor.GetChapterList(tocURL)
	if err != nil {
		return nil, err
	}

	var result []*Chapter
	for i, ch := range chapters {
		result = append(result, &Chapter{
			ChapterTitle: ch.Name,
			ChapterURL:   ch.URL,
			Index:        i,
		})
	}
	return result, nil
}

// LegadoGetChapterContent gets chapter content using the legado executor
func (ls *LegadoBookSource) LegadoGetChapterContent(chapterURL string) (string, error) {
	return ls.Executor.GetFullChapterContent(chapterURL)
}
