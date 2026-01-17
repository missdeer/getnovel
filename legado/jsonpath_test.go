package legado

import (
	"testing"
)

const testJSON = `{
	"code": 0,
	"data": {
		"books": [
			{"id": 1, "name": "Book One", "author": "Author A"},
			{"id": 2, "name": "Book Two", "author": "Author B"},
			{"id": 3, "name": "Book Three", "author": "Author C"}
		],
		"total": 3
	}
}`

func TestJSONPathParser_Parse(t *testing.T) {
	parser := NewJSONPathParserFromString(testJSON, "")

	tests := []struct {
		rule     string
		expected []string
	}{
		{"$.data.total", []string{"3"}},
		{"$.data.books.#.name", []string{"Book One", "Book Two", "Book Three"}},
		{"$.data.books.0.name", []string{"Book One"}},
		{"$.data.books.#.author", []string{"Author A", "Author B", "Author C"}},
		{"@json:$.code", []string{"0"}},
		// Test without $. prefix (gjson native syntax)
		{"data.total", []string{"3"}},
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

func TestJSONPathParser_GetElements(t *testing.T) {
	parser := NewJSONPathParserFromString(testJSON, "")

	elements := parser.GetElements("$.data.books")
	if len(elements) != 3 {
		t.Errorf("GetElements got %d elements, want 3", len(elements))
	}
}

func TestJSONPathParser_ParseValue(t *testing.T) {
	parser := NewJSONPathParserFromString("", "")

	element := `{"id": 1, "name": "Test Book", "author": "Test Author"}`

	tests := []struct {
		path     string
		expected string
	}{
		{"name", "Test Book"},
		{"author", "Test Author"},
		{"id", "1"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := parser.ParseValue(element, tt.path)
			if got != tt.expected {
				t.Errorf("ParseValue(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

func TestJSONPathParser_WithRegex(t *testing.T) {
	json := `{"title": "【完结】Test Book"}`
	parser := NewJSONPathParserFromString(json, "")

	result := parser.Parse("$.title##【[^】]+】##")
	if len(result) != 1 || result[0] != "Test Book" {
		t.Errorf("JSONPath with regex got %v, want [Test Book]", result)
	}
}
