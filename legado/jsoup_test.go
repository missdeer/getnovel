package legado

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

const testHTML = `
<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body>
	<div class="container">
		<div class="book-list">
			<div class="book-item">
				<h3 class="book-title"><a href="/book/1">Book One</a></h3>
				<p class="author">Author A</p>
				<span class="genre">Fantasy</span>
			</div>
			<div class="book-item">
				<h3 class="book-title"><a href="/book/2">Book Two</a></h3>
				<p class="author">Author B</p>
				<span class="genre">Sci-Fi</span>
			</div>
			<div class="book-item">
				<h3 class="book-title"><a href="/book/3">Book Three</a></h3>
				<p class="author">Author C</p>
				<span class="genre">Romance</span>
			</div>
		</div>
		<div id="content">
			<p>Paragraph 1</p>
			<p>Paragraph 2</p>
		</div>
	</div>
</body>
</html>
`

func getTestDoc() *goquery.Document {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(testHTML))
	return doc
}

func TestJsoupParser_Parse(t *testing.T) {
	doc := getTestDoc()
	parser := NewJsoupParser(doc, "https://example.com")

	tests := []struct {
		rule     string
		expected []string
	}{
		{"class.book-title@text", []string{"Book One", "Book Two", "Book Three"}},
		{"class.book-title.0@text", []string{"Book One"}},
		{"class.book-title.-1@text", []string{"Book Three"}},
		{"class.book-item@class.author@text", []string{"Author A", "Author B", "Author C"}},
		{"id.content@tag.p@text", []string{"Paragraph 1", "Paragraph 2"}},
		{"class.book-title@tag.a@href", []string{"https://example.com/book/1", "https://example.com/book/2", "https://example.com/book/3"}},
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

func TestJsoupParser_GetElements(t *testing.T) {
	doc := getTestDoc()
	parser := NewJsoupParser(doc, "https://example.com")

	elements := parser.GetElements("class.book-item")
	if len(elements) != 3 {
		t.Errorf("GetElements got %d elements, want 3", len(elements))
	}
}

func TestJsoupParser_ReverseList(t *testing.T) {
	doc := getTestDoc()
	parser := NewJsoupParser(doc, "https://example.com")

	// Test reverse flag
	got := parser.Parse("-class.book-title@text")
	expected := []string{"Book Three", "Book Two", "Book One"}

	if len(got) != len(expected) {
		t.Errorf("Reverse parse got %d results, want %d", len(got), len(expected))
		return
	}

	for i, v := range got {
		if v != expected[i] {
			t.Errorf("Reverse parse[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestJsoupParser_URLResolve(t *testing.T) {
	doc := getTestDoc()
	parser := NewJsoupParser(doc, "https://example.com/books/")

	tests := []struct {
		href     string
		expected string
	}{
		{"/book/1", "https://example.com/book/1"},
		{"book/1", "https://example.com/books/book/1"},
		{"https://other.com/book", "https://other.com/book"},
		{"//cdn.example.com/img.jpg", "https://cdn.example.com/img.jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.href, func(t *testing.T) {
			got := parser.resolveURL(tt.href)
			if got != tt.expected {
				t.Errorf("resolveURL(%q) = %q, want %q", tt.href, got, tt.expected)
			}
		})
	}
}
