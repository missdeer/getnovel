package legado

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/missdeer/golib/httputil"
)

// Executor executes book source rules to fetch and parse content
type Executor struct {
	source   *BookSource
	timeout  time.Duration
	headers  http.Header
	cookie   string
	jsEngine *JSEngine
}

// NewExecutor creates a new book source executor
func NewExecutor(source *BookSource) *Executor {
	e := &Executor{
		source:  source,
		timeout: 30 * time.Second,
		headers: http.Header{
			"User-Agent": []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36"},
		},
	}

	// Parse custom headers
	if source.Header != "" {
		var headerMap map[string]string
		if err := json.Unmarshal([]byte(source.Header), &headerMap); err == nil {
			for k, v := range headerMap {
				e.headers.Set(k, v)
			}
		}
	}

	return e
}

// SetTimeout sets the request timeout
func (e *Executor) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

// SetCookie sets the cookie for requests
func (e *Executor) SetCookie(cookie string) {
	e.cookie = cookie
}

// getJSEngine gets or creates the JS engine
func (e *Executor) getJSEngine(baseURL string) *JSEngine {
	if e.jsEngine == nil {
		e.jsEngine = NewJSEngine(baseURL)
		if e.source.JsLib != "" {
			e.jsEngine.SetJsLib(e.source.JsLib)
		}
	}
	e.jsEngine.SetBaseURL(baseURL)
	return e.jsEngine
}

// Search searches for books by keyword
func (e *Executor) Search(keyword string, page int) ([]SearchResult, error) {
	if e.source.SearchURL == "" {
		return nil, fmt.Errorf("no search URL configured")
	}

	// Build search URL
	searchURL, err := e.buildURL(e.source.SearchURL, keyword, page)
	if err != nil {
		return nil, fmt.Errorf("build search URL: %w", err)
	}

	// Fetch search page
	content, finalURL, err := e.fetchURL(searchURL)
	if err != nil {
		return nil, fmt.Errorf("fetch search page: %w", err)
	}

	// Parse search results
	return e.parseSearchResults(content, finalURL)
}

// GetBookInfo gets book information from book URL
func (e *Executor) GetBookInfo(bookURL string) (*BookInfo, error) {
	// Fetch book page
	content, finalURL, err := e.fetchURL(bookURL)
	if err != nil {
		return nil, fmt.Errorf("fetch book page: %w", err)
	}

	return e.parseBookInfo(content, finalURL)
}

// GetChapterList gets the table of contents
func (e *Executor) GetChapterList(tocURL string) ([]Chapter, error) {
	// Fetch TOC page
	content, finalURL, err := e.fetchURL(tocURL)
	if err != nil {
		return nil, fmt.Errorf("fetch TOC page: %w", err)
	}

	return e.parseChapterList(content, finalURL)
}

// GetChapterContent gets chapter content
func (e *Executor) GetChapterContent(chapterURL string) (*ChapterContent, error) {
	// Fetch chapter page
	content, finalURL, err := e.fetchURL(chapterURL)
	if err != nil {
		return nil, fmt.Errorf("fetch chapter: %w", err)
	}

	return e.parseChapterContent(content, finalURL)
}

// buildURL builds a URL from template
func (e *Executor) buildURL(template string, key string, page int) (string, error) {
	template = strings.TrimSpace(template)

	// Handle @js: prefix
	if strings.HasPrefix(template, "@js:") {
		engine := e.getJSEngine(e.source.BookSourceURL)
		engine.vm.Set("key", key)
		engine.vm.Set("page", page)

		result, err := engine.EvalString(strings.TrimPrefix(template, "@js:"))
		if err != nil {
			return "", err
		}
		return e.resolveURL(result), nil
	}

	// Replace {{key}} and {{page}}
	result := template
	result = strings.ReplaceAll(result, "{{key}}", url.QueryEscape(key))
	result = strings.ReplaceAll(result, "{{page}}", fmt.Sprintf("%d", page))

	// Handle more complex templates like {{(page-1)*20}}
	result = templatePattern.ReplaceAllStringFunc(result, func(match string) string {
		expr := match[2 : len(match)-2]

		// Simple expression evaluation
		if strings.Contains(expr, "page") {
			engine := e.getJSEngine(e.source.BookSourceURL)
			engine.vm.Set("page", page)
			engine.vm.Set("key", key)

			val, err := engine.EvalString(expr)
			if err != nil {
				return match
			}
			return val
		}
		return match
	})

	return e.resolveURL(result), nil
}

// resolveURL resolves a relative URL against the book source URL
func (e *Executor) resolveURL(urlStr string) string {
	urlStr = strings.TrimSpace(urlStr)
	if urlStr == "" {
		return ""
	}

	// Already absolute
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") {
		return urlStr
	}

	// Protocol-relative
	if strings.HasPrefix(urlStr, "//") {
		return "https:" + urlStr
	}

	// Get base URL
	baseURL := e.source.BookSourceURL
	if baseURL == "" {
		return urlStr
	}

	// Parse base URL
	base, err := url.Parse(baseURL)
	if err != nil {
		return urlStr
	}

	// Absolute path
	if strings.HasPrefix(urlStr, "/") {
		return fmt.Sprintf("%s://%s%s", base.Scheme, base.Host, urlStr)
	}

	// Relative path
	return fmt.Sprintf("%s://%s/%s", base.Scheme, base.Host, urlStr)
}

// fetchURL fetches content from a URL
func (e *Executor) fetchURL(urlStr string) ([]byte, string, error) {
	urlStr = strings.TrimSpace(urlStr)

	// Parse URL options
	var requestURL string
	var method = "GET"
	var body string
	var charset = "UTF-8"

	// Check for options after comma
	if idx := strings.Index(urlStr, ",{"); idx != -1 {
		requestURL = strings.TrimSpace(urlStr[:idx])

		optionsStr := urlStr[idx+1:]
		var options map[string]interface{}
		if err := json.Unmarshal([]byte(optionsStr), &options); err == nil {
			if m, ok := options["method"].(string); ok {
				method = strings.ToUpper(m)
			}
			if b, ok := options["body"].(string); ok {
				body = b
			}
			if c, ok := options["charset"].(string); ok {
				charset = c
			}
		}
	} else {
		requestURL = urlStr
	}

	requestURL = e.resolveURL(requestURL)

	// Set headers
	headers := e.headers.Clone()
	if e.cookie != "" {
		headers.Set("Cookie", e.cookie)
	}

	var content []byte
	var err error

	if method == "POST" {
		headers.Set("Content-Type", "application/x-www-form-urlencoded")
		content, err = postBytes(requestURL, headers, []byte(body), e.timeout, 3)
	} else {
		content, err = httputil.GetBytes(requestURL, headers, e.timeout, 3)
	}

	if err != nil {
		return nil, requestURL, err
	}

	// Handle charset conversion if needed
	_ = charset // TODO: implement charset conversion

	return content, requestURL, nil
}

// parseSearchResults parses search result page
func (e *Executor) parseSearchResults(content []byte, baseURL string) ([]SearchResult, error) {
	rules := e.source.RuleSearch
	if rules.BookList == "" {
		return nil, fmt.Errorf("no bookList rule")
	}

	analyzer := NewRuleAnalyzer(content, baseURL)
	analyzer.SetJSEngine(e.getJSEngine(baseURL))

	// Get book list elements
	elements := analyzer.GetElements(rules.BookList)
	if len(elements) == 0 {
		return nil, nil
	}

	var results []SearchResult
	for _, elem := range elements {
		result := SearchResult{}

		if rules.Name != "" {
			if vals := analyzer.ParseFromSelection(elem, rules.Name); len(vals) > 0 {
				result.Name = vals[0]
			}
		}
		if rules.Author != "" {
			if vals := analyzer.ParseFromSelection(elem, rules.Author); len(vals) > 0 {
				result.Author = vals[0]
			}
		}
		if rules.Kind != "" {
			if vals := analyzer.ParseFromSelection(elem, rules.Kind); len(vals) > 0 {
				result.Kind = vals[0]
			}
		}
		if rules.LastChapter != "" {
			if vals := analyzer.ParseFromSelection(elem, rules.LastChapter); len(vals) > 0 {
				result.LastChapter = vals[0]
			}
		}
		if rules.Intro != "" {
			if vals := analyzer.ParseFromSelection(elem, rules.Intro); len(vals) > 0 {
				result.Intro = vals[0]
			}
		}
		if rules.CoverURL != "" {
			if vals := analyzer.ParseFromSelection(elem, rules.CoverURL); len(vals) > 0 {
				result.CoverURL = e.resolveURL(vals[0])
			}
		}
		if rules.BookURL != "" {
			if vals := analyzer.ParseFromSelection(elem, rules.BookURL); len(vals) > 0 {
				result.BookURL = e.resolveURL(vals[0])
			}
		}
		if rules.WordCount != "" {
			if vals := analyzer.ParseFromSelection(elem, rules.WordCount); len(vals) > 0 {
				result.WordCount = vals[0]
			}
		}

		// Only add if we have at least a name
		if result.Name != "" {
			results = append(results, result)
		}
	}

	return results, nil
}

// parseBookInfo parses book information page
func (e *Executor) parseBookInfo(content []byte, baseURL string) (*BookInfo, error) {
	rules := e.source.RuleBookInfo
	analyzer := NewRuleAnalyzer(content, baseURL)
	analyzer.SetJSEngine(e.getJSEngine(baseURL))

	info := &BookInfo{}

	if rules.Name != "" {
		info.Name = analyzer.GetString(rules.Name)
	}
	if rules.Author != "" {
		info.Author = analyzer.GetString(rules.Author)
	}
	if rules.Kind != "" {
		info.Kind = analyzer.GetString(rules.Kind)
	}
	if rules.LastChapter != "" {
		info.LastChapter = analyzer.GetString(rules.LastChapter)
	}
	if rules.Intro != "" {
		info.Intro = analyzer.GetString(rules.Intro)
	}
	if rules.CoverURL != "" {
		info.CoverURL = e.resolveURL(analyzer.GetString(rules.CoverURL))
	}
	if rules.TOCURL != "" {
		tocURL := analyzer.GetString(rules.TOCURL)
		if tocURL == "baseUrl" {
			info.TOCURL = baseURL
		} else {
			info.TOCURL = e.resolveURL(tocURL)
		}
	} else {
		info.TOCURL = baseURL
	}
	if rules.WordCount != "" {
		info.WordCount = analyzer.GetString(rules.WordCount)
	}

	return info, nil
}

// parseChapterList parses table of contents
func (e *Executor) parseChapterList(content []byte, baseURL string) ([]Chapter, error) {
	rules := e.source.RuleTOC
	if rules.ChapterList == "" {
		return nil, fmt.Errorf("no chapterList rule")
	}

	analyzer := NewRuleAnalyzer(content, baseURL)
	analyzer.SetJSEngine(e.getJSEngine(baseURL))

	// Get chapter list elements
	elements := analyzer.GetElements(rules.ChapterList)
	if len(elements) == 0 {
		return nil, nil
	}

	var chapters []Chapter
	for _, elem := range elements {
		chapter := Chapter{}

		if rules.ChapterName != "" {
			if vals := analyzer.ParseFromSelection(elem, rules.ChapterName); len(vals) > 0 {
				chapter.Name = vals[0]
			}
		}
		if rules.ChapterURL != "" {
			if vals := analyzer.ParseFromSelection(elem, rules.ChapterURL); len(vals) > 0 {
				chapter.URL = e.resolveURL(vals[0])
			}
		}
		if rules.IsVIP != "" {
			if vals := analyzer.ParseFromSelection(elem, rules.IsVIP); len(vals) > 0 {
				chapter.IsVIP = vals[0] != "" && vals[0] != "false" && vals[0] != "0"
			}
		}
		if rules.IsVolume != "" {
			if vals := analyzer.ParseFromSelection(elem, rules.IsVolume); len(vals) > 0 {
				chapter.IsVolume = vals[0] != "" && vals[0] != "false" && vals[0] != "0"
			}
		}

		// Only add if we have a name or URL
		if chapter.Name != "" || chapter.URL != "" {
			chapters = append(chapters, chapter)
		}
	}

	return chapters, nil
}

// parseChapterContent parses chapter content
func (e *Executor) parseChapterContent(content []byte, baseURL string) (*ChapterContent, error) {
	rules := e.source.RuleContent
	if rules.Content == "" {
		return nil, fmt.Errorf("no content rule")
	}

	analyzer := NewRuleAnalyzer(content, baseURL)
	analyzer.SetJSEngine(e.getJSEngine(baseURL))

	cc := &ChapterContent{}

	// Get content
	contentStr := analyzer.GetString(rules.Content)

	// Apply replaceRegex if specified
	if rules.ReplaceRegex != "" {
		replaceRules := ParseReplaceRegex(rules.ReplaceRegex)
		contentStr = ApplyReplaceRules(contentStr, replaceRules)
	}

	cc.Content = contentStr

	// Get next page URL
	if rules.NextContentURL != "" {
		cc.NextPageURL = e.resolveURL(analyzer.GetString(rules.NextContentURL))
	}

	return cc, nil
}

// MaxChapterPages is the maximum number of pages to fetch for a single chapter
const MaxChapterPages = 100

// GetFullChapterContent gets complete chapter content including all pages
func (e *Executor) GetFullChapterContent(chapterURL string) (string, error) {
	var fullContent strings.Builder

	currentURL := chapterURL
	visited := make(map[string]bool)

	for currentURL != "" && !visited[currentURL] && len(visited) < MaxChapterPages {
		visited[currentURL] = true

		cc, err := e.GetChapterContent(currentURL)
		if err != nil {
			if fullContent.Len() > 0 {
				break // Return what we have
			}
			return "", err
		}

		fullContent.WriteString(cc.Content)

		currentURL = cc.NextPageURL
		if currentURL != "" {
			fullContent.WriteString("\n")
		}
	}

	return fullContent.String(), nil
}

// CreateAnalyzer creates a RuleAnalyzer for custom parsing
func (e *Executor) CreateAnalyzer(content []byte, baseURL string) *RuleAnalyzer {
	analyzer := NewRuleAnalyzer(content, baseURL)
	analyzer.SetJSEngine(e.getJSEngine(baseURL))
	return analyzer
}

// ParseFromHTML parses HTML content using goquery
func ParseFromHTML(content []byte) (*goquery.Document, error) {
	return goquery.NewDocumentFromReader(strings.NewReader(string(content)))
}

// LoadBookSources loads book sources from JSON
func LoadBookSources(data []byte) ([]BookSource, error) {
	var sources []BookSource
	if err := json.Unmarshal(data, &sources); err != nil {
		// Try single source
		var source BookSource
		if err := json.Unmarshal(data, &source); err != nil {
			return nil, err
		}
		return []BookSource{source}, nil
	}
	return sources, nil
}
