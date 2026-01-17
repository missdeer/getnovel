package legado

import (
	"testing"
)

func TestDetectRuleType(t *testing.T) {
	tests := []struct {
		rule     string
		expected RuleType
	}{
		{"@css:div.content > p", RuleTypeCSS},
		{"@XPath://div[@class='content']", RuleTypeXPath},
		{"//div[@class='content']", RuleTypeXPath},
		{"@json:$.data.items", RuleTypeJSONPath},
		{"$.data.items", RuleTypeJSONPath},
		{"@js:result.replace(/\\s/g, '')", RuleTypeJS},
		{":href=\"([^\"]*)", RuleTypeRegex},
		{"class.content@text", RuleTypeDefault},
		{"tag.div.0@text", RuleTypeDefault},
	}

	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			got := DetectRuleType(tt.rule)
			if got != tt.expected {
				t.Errorf("DetectRuleType(%q) = %v, want %v", tt.rule, got, tt.expected)
			}
		})
	}
}

func TestSplitRuleByCombinator(t *testing.T) {
	tests := []struct {
		rule           string
		expectedParts  int
		expectedComb   RuleCombinator
	}{
		{"class.a@text&&class.b@text", 2, CombinatorAnd},
		{"class.a@text||class.b@text", 2, CombinatorOr},
		{"class.a@text%%class.b@text", 2, CombinatorPercent},
		{"class.a@text", 1, CombinatorNone},
		{"a&&b&&c", 3, CombinatorAnd},
	}

	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			parts, comb := SplitRuleByCombinator(tt.rule)
			if len(parts) != tt.expectedParts {
				t.Errorf("SplitRuleByCombinator(%q) got %d parts, want %d", tt.rule, len(parts), tt.expectedParts)
			}
			if comb != tt.expectedComb {
				t.Errorf("SplitRuleByCombinator(%q) got combinator %v, want %v", tt.rule, comb, tt.expectedComb)
			}
		})
	}
}

func TestParseRule(t *testing.T) {
	tests := []struct {
		rule           string
		expectedRule   string
		hasReplacement bool
		pattern        string
		replacement    string
	}{
		{"class.content@text##\\s+##", "class.content@text", true, "\\s+", ""},
		{"class.content@text##author:\\s*(.+)##$1", "class.content@text", true, "author:\\s*(.+)", "$1"},
		{"class.content@text", "class.content@text", false, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			parsed := ParseRule(tt.rule)
			if parsed.Rule != tt.expectedRule {
				t.Errorf("ParseRule(%q).Rule = %q, want %q", tt.rule, parsed.Rule, tt.expectedRule)
			}
			if parsed.HasReplacement != tt.hasReplacement {
				t.Errorf("ParseRule(%q).HasReplacement = %v, want %v", tt.rule, parsed.HasReplacement, tt.hasReplacement)
			}
			if parsed.RegexPattern != tt.pattern {
				t.Errorf("ParseRule(%q).RegexPattern = %q, want %q", tt.rule, parsed.RegexPattern, tt.pattern)
			}
			if parsed.RegexReplace != tt.replacement {
				t.Errorf("ParseRule(%q).RegexReplace = %q, want %q", tt.rule, parsed.RegexReplace, tt.replacement)
			}
		})
	}
}

func TestExtractJSBlock(t *testing.T) {
	tests := []struct {
		rule   string
		before string
		js     string
		after  string
		hasJS  bool
	}{
		{"prefix<js>code</js>suffix", "prefix", "code", "suffix", true},
		{"<js>code</js>", "", "code", "", true},
		{"no js here", "no js here", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			before, js, after, hasJS := ExtractJSBlock(tt.rule)
			if before != tt.before || js != tt.js || after != tt.after || hasJS != tt.hasJS {
				t.Errorf("ExtractJSBlock(%q) = (%q, %q, %q, %v), want (%q, %q, %q, %v)",
					tt.rule, before, js, after, hasJS, tt.before, tt.js, tt.after, tt.hasJS)
			}
		})
	}
}

func TestExtractTemplates(t *testing.T) {
	tests := []struct {
		rule     string
		expected []string
	}{
		{"{{title}} by {{author}}", []string{"title", "author"}},
		{"no templates", nil},
		{"{{single}}", []string{"single"}},
	}

	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			got := ExtractTemplates(tt.rule)
			if len(got) != len(tt.expected) {
				t.Errorf("ExtractTemplates(%q) = %v, want %v", tt.rule, got, tt.expected)
				return
			}
			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("ExtractTemplates(%q)[%d] = %q, want %q", tt.rule, i, v, tt.expected[i])
				}
			}
		})
	}
}
