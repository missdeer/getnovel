package legado

import (
	"testing"
)

const testXPathHTML = `
<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
	<div class="container">
		<div class="book">
			<span class="title">Book One</span>
			<span class="author">Author A</span>
		</div>
		<div class="book">
			<span class="title">Book Two</span>
			<span class="author">Author B</span>
		</div>
	</div>
</body>
</html>
`

func TestXPathParser_Parse(t *testing.T) {
	parser, err := NewXPathParser([]byte(testXPathHTML), "https://example.com")
	if err != nil {
		t.Fatalf("NewXPathParser error: %v", err)
	}

	tests := []struct {
		rule     string
		expected []string
	}{
		{"//span[@class='title']", []string{"Book One", "Book Two"}},
		{"@XPath://span[@class='author']", []string{"Author A", "Author B"}},
		{"//div[@class='book'][1]/span[@class='title']", []string{"Book One"}},
		{"//title", []string{"Test"}},
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

func TestXPathParser_GetElements(t *testing.T) {
	parser, err := NewXPathParser([]byte(testXPathHTML), "")
	if err != nil {
		t.Fatalf("NewXPathParser error: %v", err)
	}

	elements := parser.GetElements("//div[@class='book']")
	if len(elements) != 2 {
		t.Errorf("GetElements got %d elements, want 2", len(elements))
	}
}

func TestXPathParser_WithRegex(t *testing.T) {
	html := `<div class="info">作者：张三</div>`
	parser, _ := NewXPathParser([]byte(html), "")

	result := parser.Parse("//div[@class='info']##作者：##")
	if len(result) != 1 || result[0] != "张三" {
		t.Errorf("XPath with regex got %v, want [张三]", result)
	}
}
