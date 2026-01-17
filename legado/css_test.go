package legado

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestCSSParser_Parse(t *testing.T) {
	html := `
	<html>
	<body>
		<div class="container">
			<h1>Title</h1>
			<div class="items">
				<div class="item"><a href="/1">Item 1</a></div>
				<div class="item"><a href="/2">Item 2</a></div>
			</div>
		</div>
	</body>
	</html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	parser := NewCSSParser(doc, "https://example.com")

	tests := []struct {
		rule     string
		expected []string
	}{
		{"@css:h1@text", []string{"Title"}},
		{"@css:.item a@text", []string{"Item 1", "Item 2"}},
		{"@css:.item a@href", []string{"https://example.com/1", "https://example.com/2"}},
		{"@css:div.container > h1@text", []string{"Title"}},
	}

	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			got := parser.Parse(tt.rule)
			if len(got) != len(tt.expected) {
				t.Errorf("Parse(%q) got %d results, want %d: %v", tt.rule, len(got), len(tt.expected), got)
				return
			}
			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("Parse(%q)[%d] = %q, want %q", tt.rule, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestCSSParser_WithRegex(t *testing.T) {
	html := `<div class="info">Author: John Doe</div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	parser := NewCSSParser(doc, "")

	result := parser.Parse("@css:.info@text##Author:\\s*##")
	if len(result) != 1 || result[0] != "John Doe" {
		t.Errorf("CSS with regex got %v, want [John Doe]", result)
	}
}

func TestCSSParser_GetElements(t *testing.T) {
	html := `
	<ul>
		<li>Item 1</li>
		<li>Item 2</li>
		<li>Item 3</li>
	</ul>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	parser := NewCSSParser(doc, "")

	elements := parser.GetElements("@css:li")
	if len(elements) != 3 {
		t.Errorf("GetElements got %d elements, want 3", len(elements))
	}
}
