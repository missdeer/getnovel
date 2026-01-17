package handler

import (
	"log"
	"strings"

	"github.com/missdeer/getnovel/config"
	"github.com/missdeer/getnovel/ebook/bs"
)

var (
	// currentLegadoSource holds the matched legado source for the current download
	currentLegadoSource *bs.LegadoBookSource
)

func init() {
	registerNovelSiteHandler(&config.NovelSiteHandler{
		Sites: []config.NovelSite{
			{
				Title: `Legado书源`,
				Urls:  []string{`支持已加载的Legado书源`},
			},
		},
		CanHandle: func(u string) bool {
			// Try to find a matching legado source for this URL
			source := bs.FindLegadoSourceByURL(u)
			if source != nil {
				currentLegadoSource = source
				log.Printf("Matched legado source: %s (%s)", source.Source.BookSourceName, source.Source.BookSourceURL)
				return true
			}
			return false
		},
		ExtractChapterList: extractLegadoChapterList,
		ExtractChapterContent: func(u string, rawPageContent []byte) []byte {
			// The legado handler uses Download function instead
			// This is a fallback that shouldn't normally be called
			if currentLegadoSource == nil {
				return rawPageContent
			}
			content, err := currentLegadoSource.LegadoGetChapterContent(u)
			if err != nil {
				log.Printf("Failed to get chapter content via legado: %v", err)
				return rawPageContent
			}
			// Convert content to paragraph format
			content = formatContent(content)
			return []byte(content)
		},
	})
}

// extractLegadoChapterList extracts chapter list using legado source
func extractLegadoChapterList(u string, rawPageContent []byte) (title string, chapters []*config.NovelChapterInfo) {
	if currentLegadoSource == nil {
		log.Println("No legado source matched")
		return "", nil
	}

	// Get book info first
	bookInfo, err := currentLegadoSource.LegadoGetBookInfo(u)
	if err != nil {
		log.Printf("Failed to get book info: %v", err)
		return "", nil
	}

	title = bookInfo.Name
	if title == "" {
		title = "未知书名"
	}

	// Determine TOC URL
	tocURL := bookInfo.ChapterURL
	if tocURL == "" {
		tocURL = u
	}

	// Get chapter list
	chapterList, err := currentLegadoSource.LegadoGetChapterList(tocURL)
	if err != nil {
		log.Printf("Failed to get chapter list: %v", err)
		return title, nil
	}

	// Convert to NovelChapterInfo
	for _, ch := range chapterList {
		chapters = append(chapters, &config.NovelChapterInfo{
			Index: ch.Index + 1, // 1-based index
			Title: ch.ChapterTitle,
			URL:   ch.ChapterURL,
		})
	}

	log.Printf("Found %d chapters for book: %s", len(chapters), title)
	return title, chapters
}

// formatContent formats the raw content with proper paragraph tags
func formatContent(content string) string {
	// Split by newlines and wrap in paragraph tags
	lines := strings.Split(content, "\n")
	var result strings.Builder
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		result.WriteString("<p>")
		result.WriteString(line)
		result.WriteString("</p>")
	}
	return result.String()
}
