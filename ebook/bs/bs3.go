package bs

import (
	"fmt"
)

// BookSourceV3 book source structure
type BookSourceV3 struct {
	BookSourceGroup string `json:"bookSourceGroup"`
	BookSourceName  string `json:"bookSourceName"`
	BookSourceURL   string `json:"bookSourceUrl"`
	Enable          bool   `json:"enabled"`
	Header          string `json:"header"`
	RuleBookInfo    struct {
		Author      string `json:"author"`
		CoverURL    string `json:"coverUrl"`
		Intro       string `json:"intro"`
		Kind        string `json:"kind"`
		LastChapter string `json:"lastChapter"`
		Name        string `json:"name"`
		TOCURL      string `json:"tocUrl"`
	} `json:"ruleBookInfo"`
	RuleContent struct {
		Content        string `json:"content"`
		NextContentURL string `json:"nextContentUrl"`
		SourceRegex    string `json:"sourceRegex"`
	} `json:"ruleContent"`
	RuleExplore struct {
		Author      string `json:"author"`
		BookList    string `json:"bookList"`
		BookURL     string `json:"bookUrl"`
		CoverURL    string `json:"coverUrl"`
		Intro       string `json:"intro"`
		Kind        string `json:"kind"`
		LastChapter string `json:"lastChapter"`
		Name        string `json:"name"`
	} `json:"ruleExplore"`
	RuleSearch struct {
		Author      string `json:"author"`
		BookList    string `json:"bookList"`
		BookURL     string `json:"bookUrl"`
		CoverURL    string `json:"coverUrl"`
		Intro       string `json:"intro"`
		Kind        string `json:"kind"`
		LastChapter string `json:"lastChapter"`
		Name        string `json:"name"`
	} `json:"ruleSearch"`
	RuleTOC struct {
		ChapterList string `json:"chapterList"`
		ChapterName string `json:"chapterName"`
		ChapterURL  string `json:"chapterUrl"`
	} `json:"ruleToc"`
	SearchURL string `json:"searchUrl"`
	Weight    int    `json:"weight"`
}

func (bs BookSourceV3) String() string {
	return fmt.Sprintf("%s( %s )", bs.BookSourceName, bs.BookSourceURL)
}
