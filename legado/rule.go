package legado

import (
	"regexp"
	"strings"
)

// RuleType represents the type of rule parser to use
type RuleType int

const (
	RuleTypeDefault  RuleType = iota // JSOUP Default with @ separator
	RuleTypeCSS                      // CSS selector (@css:)
	RuleTypeXPath                    // XPath (// or @XPath:)
	RuleTypeJSONPath                 // JSONPath ($. or @json:)
	RuleTypeJS                       // JavaScript (@js: or <js>...</js>)
	RuleTypeRegex                    // Regex AllInOne (:...)
)

// DetectRuleType determines the parser type based on rule prefix
func DetectRuleType(rule string) RuleType {
	rule = strings.TrimSpace(rule)

	switch {
	case strings.HasPrefix(rule, "@css:"):
		return RuleTypeCSS
	case strings.HasPrefix(rule, "@XPath:") || strings.HasPrefix(rule, "//"):
		return RuleTypeXPath
	case strings.HasPrefix(rule, "@json:") || strings.HasPrefix(rule, "$."):
		return RuleTypeJSONPath
	case strings.HasPrefix(rule, "@js:") || strings.HasPrefix(rule, "<js>"):
		return RuleTypeJS
	case strings.HasPrefix(rule, ":"):
		return RuleTypeRegex
	default:
		return RuleTypeDefault
	}
}

// RuleCombinator represents how multiple rules are combined
type RuleCombinator int

const (
	CombinatorNone    RuleCombinator = iota // Single rule
	CombinatorAnd                           // && - merge all results
	CombinatorOr                            // || - first non-empty result
	CombinatorPercent                       // %% - interleave results
)

// SplitRuleByCombinator splits a rule string by combinators
// Returns the parts and the combinator type
func SplitRuleByCombinator(rule string) ([]string, RuleCombinator) {
	// Check for combinators in order of precedence
	if strings.Contains(rule, "&&") {
		return splitAndTrim(rule, "&&"), CombinatorAnd
	}
	if strings.Contains(rule, "||") {
		return splitAndTrim(rule, "||"), CombinatorOr
	}
	if strings.Contains(rule, "%%") {
		return splitAndTrim(rule, "%%"), CombinatorPercent
	}
	return []string{rule}, CombinatorNone
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// ParsedRule represents a parsed rule with optional regex replacement
type ParsedRule struct {
	Rule           string
	Type           RuleType
	RegexPattern   string // ##pattern##
	RegexReplace   string // ##replacement
	HasReplacement bool
}

// regexReplacePattern matches ##pattern##replacement or ##pattern
var regexReplacePattern = regexp.MustCompile(`##(.+?)(?:##(.*))?$`)

// templatePattern matches {{expression}} templates
var templatePattern = regexp.MustCompile(`\{\{(.+?)\}\}`)

// baseURLPattern extracts the scheme and host from a URL
var baseURLPattern = regexp.MustCompile(`^(https?://[^/]+)`)

// ParseRule parses a rule string, extracting any regex replacement
func ParseRule(rule string) ParsedRule {
	parsed := ParsedRule{
		Rule: rule,
		Type: DetectRuleType(rule),
	}

	// Check for regex replacement pattern ##...##...
	if idx := strings.Index(rule, "##"); idx != -1 {
		mainRule := rule[:idx]
		regexPart := rule[idx:]

		if matches := regexReplacePattern.FindStringSubmatch(regexPart); matches != nil {
			parsed.Rule = mainRule
			parsed.RegexPattern = matches[1]
			if len(matches) > 2 {
				parsed.RegexReplace = matches[2]
			}
			parsed.HasReplacement = true
		}
	}

	return parsed
}

// SplitByNewlineAndJS splits a rule that has multiple steps (newline or @js:)
func SplitByNewlineAndJS(rule string) []string {
	// First check if there's a @js: that's not at the start
	if idx := strings.Index(rule, "\n@js:"); idx != -1 {
		return []string{rule[:idx], rule[idx+1:]}
	}
	if idx := strings.Index(rule, "\n"); idx != -1 {
		parts := strings.SplitN(rule, "\n", 2)
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return []string{rule}
}

// ExtractJSBlock extracts JavaScript from <js>...</js> blocks
func ExtractJSBlock(rule string) (before, js, after string, hasJS bool) {
	jsStart := strings.Index(rule, "<js>")
	if jsStart == -1 {
		return rule, "", "", false
	}

	jsEnd := strings.Index(rule, "</js>")
	if jsEnd == -1 {
		// No closing tag, treat rest as JS
		return rule[:jsStart], rule[jsStart+4:], "", true
	}

	return rule[:jsStart], rule[jsStart+4 : jsEnd], rule[jsEnd+5:], true
}

// ExtractTemplates extracts {{...}} templates from a rule
func ExtractTemplates(rule string) []string {
	matches := templatePattern.FindAllStringSubmatch(rule, -1)
	result := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			result = append(result, m[1])
		}
	}
	return result
}

// HasTemplate checks if a rule contains {{...}} templates
func HasTemplate(rule string) bool {
	return strings.Contains(rule, "{{") && strings.Contains(rule, "}}")
}
