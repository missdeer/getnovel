package legado

import (
	"testing"
)

func TestRegexParser_ParseAllInOne(t *testing.T) {
	content := `<a href="/book/1">Book One</a><a href="/book/2">Book Two</a>`
	parser := NewRegexParser(content, "")

	matches := parser.ParseAllInOne(`:href="([^"]+)">([^<]+)`)
	if len(matches) != 2 {
		t.Fatalf("ParseAllInOne got %d matches, want 2", len(matches))
	}

	// Check first match
	if matches[0][1] != "/book/1" || matches[0][2] != "Book One" {
		t.Errorf("First match = %v, want [/book/1, Book One]", matches[0])
	}

	// Check second match
	if matches[1][1] != "/book/2" || matches[1][2] != "Book Two" {
		t.Errorf("Second match = %v, want [/book/2, Book Two]", matches[1])
	}
}

func TestRegexParser_ApplyReplacement(t *testing.T) {
	parser := NewRegexParser("", "")

	tests := []struct {
		content     string
		pattern     string
		replacement string
		expected    string
	}{
		{"Hello World", "World", "Go", "Hello Go"},
		{"Author: John", "Author:\\s*", "", "John"},
		{"【完结】Book Title", "【[^】]+】", "", "Book Title"},
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			got := parser.ApplyReplacement(tt.content, tt.pattern, tt.replacement)
			if got != tt.expected {
				t.Errorf("ApplyReplacement(%q, %q, %q) = %q, want %q",
					tt.content, tt.pattern, tt.replacement, got, tt.expected)
			}
		})
	}
}

func TestParseReplaceRegex(t *testing.T) {
	tests := []struct {
		rule          string
		expectedCount int
	}{
		{"##pattern##replacement", 1},
		{"pattern##replacement", 1},
		{"pattern1##replace1\npattern2##replace2", 2},
	}

	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			rules := ParseReplaceRegex(tt.rule)
			if len(rules) != tt.expectedCount {
				t.Errorf("ParseReplaceRegex got %d rules, want %d", len(rules), tt.expectedCount)
			}
		})
	}
}

func TestApplyReplaceRules(t *testing.T) {
	rules := []ReplaceRegexRule{
		{Pattern: "\\s+", Replacement: " ", IsRegex: true},
		{Pattern: "【[^】]+】", Replacement: "", IsRegex: true},
	}

	content := "【完结】Hello    World"
	expected := "Hello World"

	got := ApplyReplaceRules(content, rules)
	if got != expected {
		t.Errorf("ApplyReplaceRules = %q, want %q", got, expected)
	}
}

func TestRegexParser_CleanupContent(t *testing.T) {
	parser := NewRegexParser("", "")

	tests := []struct {
		content  string
		rules    string
		expected string
	}{
		{"Hello广告World", "广告", "HelloWorld"},
		{"Test【广告】Content", "【[^】]+】", "TestContent"},
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			got := parser.CleanupContent(tt.content, tt.rules)
			if got != tt.expected {
				t.Errorf("CleanupContent(%q, %q) = %q, want %q",
					tt.content, tt.rules, got, tt.expected)
			}
		})
	}
}
