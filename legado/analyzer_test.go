package legado

import (
	"testing"
)

const testAnalyzerHTML = `
<!DOCTYPE html>
<html>
<body>
	<div class="book-list">
		<div class="item">
			<h3><a href="/book/1">Book One</a></h3>
			<span class="author">Author A</span>
			<span class="genre">Fantasy</span>
		</div>
		<div class="item">
			<h3><a href="/book/2">Book Two</a></h3>
			<span class="author">Author B</span>
			<span class="genre">Sci-Fi</span>
		</div>
	</div>
</body>
</html>
`

func TestRuleAnalyzer_ParseRule(t *testing.T) {
	analyzer := NewRuleAnalyzer([]byte(testAnalyzerHTML), "https://example.com")

	tests := []struct {
		rule     string
		expected []string
	}{
		// JSOUP Default
		{"class.author@text", []string{"Author A", "Author B"}},
		// CSS
		{"@css:.item h3 a@text", []string{"Book One", "Book Two"}},
		// With combinator
		{"class.author@text||class.genre@text", []string{"Author A", "Author B"}},
	}

	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			got := analyzer.ParseRule(analyzer.content, tt.rule)
			if len(got) != len(tt.expected) {
				t.Errorf("ParseRule(%q) got %d results, want %d: %v",
					tt.rule, len(got), len(tt.expected), got)
				return
			}
			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("ParseRule(%q)[%d] = %q, want %q",
						tt.rule, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestRuleAnalyzer_Combinators(t *testing.T) {
	analyzer := NewRuleAnalyzer([]byte(testAnalyzerHTML), "")

	// Test || combinator - returns first non-empty result
	result := analyzer.ParseRule(analyzer.content, "class.nonexistent@text||class.author@text")
	if len(result) != 2 || result[0] != "Author A" {
		t.Errorf("|| combinator failed: got %v", result)
	}

	// Test && combinator - merges all results
	result = analyzer.ParseRule(analyzer.content, "class.author@text&&class.genre@text")
	if len(result) != 4 {
		t.Errorf("&& combinator failed: got %d results, want 4", len(result))
	}
}

func TestRuleAnalyzer_GetElements(t *testing.T) {
	analyzer := NewRuleAnalyzer([]byte(testAnalyzerHTML), "")

	elements := analyzer.GetElements("class.item")
	if len(elements) != 2 {
		t.Errorf("GetElements got %d elements, want 2", len(elements))
	}
}

func TestRuleAnalyzer_ParseFromSelection(t *testing.T) {
	analyzer := NewRuleAnalyzer([]byte(testAnalyzerHTML), "https://example.com")

	elements := analyzer.GetElements("class.item")
	if len(elements) < 1 {
		t.Fatal("No elements found")
	}

	// Parse from first element
	result := analyzer.ParseFromSelection(elements[0], "tag.h3@tag.a@text")
	if len(result) != 1 || result[0] != "Book One" {
		t.Errorf("ParseFromSelection got %v, want [Book One]", result)
	}

	// Parse href
	result = analyzer.ParseFromSelection(elements[0], "tag.a@href")
	if len(result) != 1 || result[0] != "https://example.com/book/1" {
		t.Errorf("ParseFromSelection href got %v, want [https://example.com/book/1]", result)
	}
}

func TestRuleAnalyzer_GetString(t *testing.T) {
	html := `<div class="title">Test Title</div>`
	analyzer := NewRuleAnalyzer([]byte(html), "")

	got := analyzer.GetString("class.title@text")
	if got != "Test Title" {
		t.Errorf("GetString = %q, want 'Test Title'", got)
	}
}

func TestRuleAnalyzer_JSONContent(t *testing.T) {
	json := `{"title": "Test Book", "author": "Test Author"}`
	analyzer := NewRuleAnalyzer([]byte(json), "")

	// Should auto-detect JSON and use JSONPath
	result := analyzer.ParseRule([]byte(json), "$.title")
	if len(result) != 1 || result[0] != "Test Book" {
		t.Errorf("JSON parsing got %v, want [Test Book]", result)
	}
}

func TestRuleAnalyzer_MultiStepRule(t *testing.T) {
	html := `<div class="info">  Author: John Doe  </div>`
	analyzer := NewRuleAnalyzer([]byte(html), "")

	// Multi-step rule: first get text, then trim with JS
	result := analyzer.ParseRule([]byte(html), "class.info@text\n@js:result.trim()")
	if len(result) != 1 || result[0] != "Author: John Doe" {
		t.Errorf("Multi-step rule got %v, want [Author: John Doe]", result)
	}
}

func TestRuleAnalyzer_WithTemplate(t *testing.T) {
	// Template processing is complex - skip for now
	t.Skip("Template processing requires full JS engine integration")
}
