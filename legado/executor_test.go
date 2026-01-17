package legado

import (
	"encoding/json"
	"os"
	"testing"
)

// TestLoadBookSources tests loading book sources from JSON
func TestLoadBookSources(t *testing.T) {
	data, err := os.ReadFile("../testdata/booksource_simple.json")
	if err != nil {
		t.Skip("Test data not found")
	}

	sources, err := LoadBookSources(data)
	if err != nil {
		t.Fatalf("LoadBookSources error: %v", err)
	}

	if len(sources) == 0 {
		t.Fatal("No sources loaded")
	}

	// Check first source
	source := sources[0]
	t.Logf("Loaded source: %s (%s)", source.BookSourceName, source.BookSourceURL)

	if source.BookSourceName == "" {
		t.Error("Source name is empty")
	}
	if source.BookSourceURL == "" {
		t.Error("Source URL is empty")
	}
}

// TestBookSourceRulesParsing tests parsing rules from real book sources
func TestBookSourceRulesParsing(t *testing.T) {
	data, err := os.ReadFile("../testdata/booksource_simple.json")
	if err != nil {
		t.Skip("Test data not found")
	}

	var sources []BookSource
	if err := json.Unmarshal(data, &sources); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	for _, source := range sources {
		t.Run(source.BookSourceName, func(t *testing.T) {
			// Test that rules can be parsed
			if source.RuleSearch.BookList != "" {
				ruleType := DetectRuleType(source.RuleSearch.BookList)
				t.Logf("  Search bookList rule type: %v", ruleType)
			}

			if source.RuleTOC.ChapterList != "" {
				ruleType := DetectRuleType(source.RuleTOC.ChapterList)
				t.Logf("  TOC chapterList rule type: %v", ruleType)
			}

			if source.RuleContent.Content != "" {
				ruleType := DetectRuleType(source.RuleContent.Content)
				t.Logf("  Content rule type: %v", ruleType)
			}
		})
	}
}

// TestParseRealHTMLWithJsoupRules tests parsing real-world HTML patterns
func TestParseRealHTMLWithJsoupRules(t *testing.T) {
	// Simulate HTML from a novel site
	html := `
	<div class="novel_cell">
		<a href="/novel/chapters/123/index.html">
			<amp-img src="https://example.com/cover.jpg"></amp-img>
			<h3>小说标题</h3>
		</a>
		<ul>
			<li>作者：测试作者</li>
			<li>简介：这是一个测试小说</li>
		</ul>
	</div>
	<div class="novel_cell">
		<a href="/novel/chapters/456/index.html">
			<amp-img src="https://example.com/cover2.jpg"></amp-img>
			<h3>另一本小说</h3>
		</a>
		<ul>
			<li>作者：另一个作者</li>
			<li>简介：另一本小说的简介</li>
		</ul>
	</div>
	`

	analyzer := NewRuleAnalyzer([]byte(html), "https://ttks.tw")

	// Test bookList rule
	elements := analyzer.GetElements("class.novel_cell")
	if len(elements) != 2 {
		t.Errorf("Got %d novel cells, want 2", len(elements))
	}

	// Test parsing from element
	if len(elements) > 0 {
		name := analyzer.ParseFromSelection(elements[0], "h3@text")
		if len(name) != 1 || name[0] != "小说标题" {
			t.Errorf("Name parsing got %v, want [小说标题]", name)
		}

		author := analyzer.ParseFromSelection(elements[0], "tag.li.0@text##作者：##")
		if len(author) != 1 || author[0] != "测试作者" {
			t.Errorf("Author parsing got %v, want [测试作者]", author)
		}

		cover := analyzer.ParseFromSelection(elements[0], "amp-img@src")
		if len(cover) != 1 || cover[0] != "https://example.com/cover.jpg" {
			t.Errorf("Cover parsing got %v, want [https://example.com/cover.jpg]", cover)
		}

		bookURL := analyzer.ParseFromSelection(elements[0], "tag.a@href")
		if len(bookURL) != 1 || bookURL[0] != "https://ttks.tw/novel/chapters/123/index.html" {
			t.Errorf("BookURL parsing got %v", bookURL)
		}
	}
}

// TestParseChapterList tests parsing chapter lists
func TestParseChapterList(t *testing.T) {
	html := `
	<div class="chapters_frame">
		<div class="pure-g">
			<div class="chapter_cell"><a href="/chapter/1.html">第一章</a></div>
			<div class="chapter_cell"><a href="/chapter/2.html">第二章</a></div>
			<div class="chapter_cell"><a href="/chapter/3.html">第三章</a></div>
		</div>
	</div>
	`

	analyzer := NewRuleAnalyzer([]byte(html), "https://example.com")

	// Rule from real book source: class.chapters_frame@class.pure-g@class.chapter_cell@tag.a
	elements := analyzer.GetElements("class.chapters_frame@class.pure-g@class.chapter_cell@tag.a")
	if len(elements) != 3 {
		t.Errorf("Got %d chapters, want 3", len(elements))
	}

	// Parse chapter names and URLs
	for i, elem := range elements {
		name := analyzer.ParseFromSelection(elem, "text")
		url := analyzer.ParseFromSelection(elem, "href")

		t.Logf("Chapter %d: %v - %v", i+1, name, url)

		if len(name) == 0 {
			t.Errorf("Chapter %d has no name", i+1)
		}
		if len(url) == 0 {
			t.Errorf("Chapter %d has no URL", i+1)
		}
	}
}

// TestParseContent tests parsing chapter content
func TestParseContent(t *testing.T) {
	html := `
	<div class="content">
		<p>第一段内容。</p>
		<p>第二段内容。</p>
		<p>【广告】请访问我们的网站</p>
		<p>第三段内容。</p>
	</div>
	`

	analyzer := NewRuleAnalyzer([]byte(html), "")

	// Rule from real book source: class.content@p@textNodes
	content := analyzer.GetString("class.content@tag.p@text")
	if content == "" {
		t.Error("Content is empty")
	}

	// Apply replaceRegex
	rules := ParseReplaceRegex("##【广告】.*##")
	cleaned := ApplyReplaceRules(content, rules)
	t.Logf("Cleaned content: %s", cleaned)
}

// TestExecutorBuildURL tests URL building
func TestExecutorBuildURL(t *testing.T) {
	source := &BookSource{
		BookSourceURL: "https://example.com",
		SearchURL:     "/search?q={{key}}&page={{page}}",
	}

	executor := NewExecutor(source)

	tests := []struct {
		key      string
		page     int
		expected string
	}{
		{"测试", 1, "https://example.com/search?q=%E6%B5%8B%E8%AF%95&page=1"},
		{"test", 2, "https://example.com/search?q=test&page=2"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, err := executor.buildURL(source.SearchURL, tt.key, tt.page)
			if err != nil {
				t.Fatalf("buildURL error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("buildURL = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestComplexJsoupRule tests complex JSOUP rules with multiple segments
func TestComplexJsoupRule(t *testing.T) {
	html := `
	<div class="book-module">
		<div class="pipe-z">
			<span>作者：张三</span>
		</div>
		<div class="pipe-z-s">
			<span>分类：玄幻</span>
		</div>
	</div>
	`

	analyzer := NewRuleAnalyzer([]byte(html), "")

	// Complex rule: .book-module@class.pipe-z@text
	result := analyzer.GetString("class.book-module@class.pipe-z@text")
	if result == "" {
		t.Error("Complex rule returned empty")
	}
	t.Logf("Result: %s", result)
}

// TestCSSAdvanced tests advanced CSS selectors
func TestCSSAdvanced(t *testing.T) {
	html := `
	<table>
		<tr><td>Header</td></tr>
		<tr><td><img src="/img1.jpg"></td></tr>
		<tr><td><img src="/img2.jpg"></td></tr>
	</table>
	`

	analyzer := NewRuleAnalyzer([]byte(html), "https://example.com")

	// CSS with nth-of-type
	result := analyzer.ParseRule([]byte(html), "@css:tr:nth-of-type(2) img@src")
	if len(result) != 1 || result[0] != "https://example.com/img1.jpg" {
		t.Errorf("CSS nth-of-type got %v", result)
	}
}
