// Package legado provides support for parsing Legado book source rules
package legado

// BookSource represents a Legado book source configuration
type BookSource struct {
	BookSourceComment string `json:"bookSourceComment,omitempty"`
	BookSourceGroup   string `json:"bookSourceGroup,omitempty"`
	BookSourceName    string `json:"bookSourceName"`
	BookSourceType    int    `json:"bookSourceType"` // 0: text, 1: audio, 2: image
	BookSourceURL     string `json:"bookSourceUrl"`
	BookURLPattern    string `json:"bookUrlPattern,omitempty"`
	CustomOrder       int    `json:"customOrder,omitempty"`
	Enabled           bool   `json:"enabled"`
	EnabledCookieJar  bool   `json:"enabledCookieJar,omitempty"`
	EnabledExplore    bool   `json:"enabledExplore,omitempty"`
	ExploreURL        string `json:"exploreUrl,omitempty"`
	Header            string `json:"header,omitempty"`
	JsLib             string `json:"jsLib,omitempty"`
	LastUpdateTime    int64  `json:"lastUpdateTime,omitempty"`
	LoginURL          string `json:"loginUrl,omitempty"`
	RespondTime       int64  `json:"respondTime,omitempty"`
	SearchURL         string `json:"searchUrl,omitempty"`
	Weight            int    `json:"weight,omitempty"`

	RuleBookInfo RuleBookInfo `json:"ruleBookInfo,omitempty"`
	RuleContent  RuleContent  `json:"ruleContent,omitempty"`
	RuleExplore  RuleExplore  `json:"ruleExplore,omitempty"`
	RuleSearch   RuleSearch   `json:"ruleSearch,omitempty"`
	RuleTOC      RuleTOC      `json:"ruleToc,omitempty"`
}

// RuleBookInfo contains rules for parsing book information page
type RuleBookInfo struct {
	Author      string `json:"author,omitempty"`
	CoverURL    string `json:"coverUrl,omitempty"`
	Init        string `json:"init,omitempty"`
	Intro       string `json:"intro,omitempty"`
	Kind        string `json:"kind,omitempty"`
	LastChapter string `json:"lastChapter,omitempty"`
	Name        string `json:"name,omitempty"`
	TOCURL      string `json:"tocUrl,omitempty"`
	WordCount   string `json:"wordCount,omitempty"`
}

// RuleContent contains rules for parsing chapter content
type RuleContent struct {
	Content        string `json:"content,omitempty"`
	NextContentURL string `json:"nextContentUrl,omitempty"`
	ReplaceRegex   string `json:"replaceRegex,omitempty"`
	SourceRegex    string `json:"sourceRegex,omitempty"`
	WebJS          string `json:"webJs,omitempty"`
}

// RuleExplore contains rules for parsing explore/discovery pages
type RuleExplore struct {
	Author      string `json:"author,omitempty"`
	BookList    string `json:"bookList,omitempty"`
	BookURL     string `json:"bookUrl,omitempty"`
	CoverURL    string `json:"coverUrl,omitempty"`
	Intro       string `json:"intro,omitempty"`
	Kind        string `json:"kind,omitempty"`
	LastChapter string `json:"lastChapter,omitempty"`
	Name        string `json:"name,omitempty"`
	WordCount   string `json:"wordCount,omitempty"`
}

// RuleSearch contains rules for parsing search results
type RuleSearch struct {
	Author       string `json:"author,omitempty"`
	BookList     string `json:"bookList,omitempty"`
	BookURL      string `json:"bookUrl,omitempty"`
	CheckKeyWord string `json:"checkKeyWord,omitempty"`
	CoverURL     string `json:"coverUrl,omitempty"`
	Intro        string `json:"intro,omitempty"`
	Kind         string `json:"kind,omitempty"`
	LastChapter  string `json:"lastChapter,omitempty"`
	Name         string `json:"name,omitempty"`
	WordCount    string `json:"wordCount,omitempty"`
}

// RuleTOC contains rules for parsing table of contents
type RuleTOC struct {
	ChapterList    string `json:"chapterList,omitempty"`
	ChapterName    string `json:"chapterName,omitempty"`
	ChapterURL     string `json:"chapterUrl,omitempty"`
	IsVIP          string `json:"isVip,omitempty"`
	IsVolume       string `json:"isVolume,omitempty"`
	NextTOCURL     string `json:"nextTocUrl,omitempty"`
	UpdateTime     string `json:"updateTime,omitempty"`
	ChapterPayType string `json:"chapterPayType,omitempty"`
}

// SearchResult represents a book found in search
type SearchResult struct {
	Name        string
	Author      string
	Kind        string
	LastChapter string
	Intro       string
	CoverURL    string
	BookURL     string
	WordCount   string
}

// BookInfo represents detailed book information
type BookInfo struct {
	Name        string
	Author      string
	Kind        string
	LastChapter string
	Intro       string
	CoverURL    string
	TOCURL      string
	WordCount   string
}

// Chapter represents a single chapter
type Chapter struct {
	Name     string
	URL      string
	IsVIP    bool
	IsVolume bool
}

// ChapterContent represents chapter content
type ChapterContent struct {
	Content     string
	NextPageURL string
}
