package legado

import (
	"regexp"
	"strconv"
	"strings"
)

// RegexParser implements regex-based parsing
// AllInOne format: :pattern for list extraction
// OnlyOne format: ##pattern##replacement### for single match
// Cleanup format: ##pattern##replacement for global replacement
type RegexParser struct {
	content string
	baseURL string
}

// NewRegexParser creates a new regex parser
func NewRegexParser(content string, baseURL string) *RegexParser {
	return &RegexParser{content: content, baseURL: baseURL}
}

// ParseAllInOne parses AllInOne regex rules (starts with :)
// Used for extracting lists from content
// Format: :href="([^"]*)"[^>]*>([^<]*)
// Returns pairs/groups based on capture groups
func (p *RegexParser) ParseAllInOne(rule string) [][]string {
	if !strings.HasPrefix(rule, ":") {
		return nil
	}

	pattern := strings.TrimPrefix(rule, ":")
	pattern = strings.TrimSpace(pattern)

	if pattern == "" {
		return nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}

	matches := re.FindAllStringSubmatch(p.content, -1)
	return matches
}

// ParseOnlyOne applies OnlyOne regex replacement
// Format: ##pattern##replacement###
// Returns the first match with replacement applied
func (p *RegexParser) ParseOnlyOne(content, pattern, replacement string) string {
	if pattern == "" {
		return content
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return content
	}

	// Find first match and replace
	match := re.FindString(content)
	if match == "" {
		return content
	}

	return re.ReplaceAllString(match, replacement)
}

// ApplyReplacement applies regex replacement to content
// Format: ##pattern##replacement
func (p *RegexParser) ApplyReplacement(content, pattern, replacement string) string {
	if pattern == "" {
		return content
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return content
	}

	return re.ReplaceAllString(content, replacement)
}

// ExtractWithGroups extracts content using regex with named or indexed groups
// Returns a map of group index/name to value
func (p *RegexParser) ExtractWithGroups(pattern string) map[string]string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}

	match := re.FindStringSubmatch(p.content)
	if match == nil {
		return nil
	}

	result := make(map[string]string)
	names := re.SubexpNames()

	for i, name := range names {
		if i < len(match) {
			if name != "" {
				result[name] = match[i]
			}
			// Also store by index using strconv for proper handling of i >= 10
			result[strconv.Itoa(i)] = match[i]
		}
	}

	return result
}

// CleanupContent applies cleanup regex rules
// Rules format: pattern1|pattern2|pattern3 (patterns to remove)
// Or: pattern##replacement|pattern2##replacement2
func (p *RegexParser) CleanupContent(content, rules string) string {
	if rules == "" {
		return content
	}

	// Split by | for multiple rules
	parts := strings.Split(rules, "|")
	result := content

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check for pattern##replacement format
		if idx := strings.Index(part, "##"); idx != -1 {
			pattern := part[:idx]
			replacement := ""
			if idx+2 < len(part) {
				replacement = part[idx+2:]
			}

			if re, err := regexp.Compile(pattern); err == nil {
				result = re.ReplaceAllString(result, replacement)
			}
		} else {
			// Just a pattern to remove
			if re, err := regexp.Compile(part); err == nil {
				result = re.ReplaceAllString(result, "")
			}
		}
	}

	return result
}

// ReplaceRegexRule represents a parsed replace regex rule
type ReplaceRegexRule struct {
	Pattern     string
	Replacement string
	IsRegex     bool
}

// ParseReplaceRegex parses replaceRegex field format
// Format: ##pattern##replacement or pattern##replacement
func ParseReplaceRegex(rule string) []ReplaceRegexRule {
	if rule == "" {
		return nil
	}

	var rules []ReplaceRegexRule

	// Handle multiple rules separated by newlines
	lines := strings.Split(rule, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove leading ## if present
		line = strings.TrimPrefix(line, "##")

		// Split by ##
		parts := strings.SplitN(line, "##", 2)
		if len(parts) == 0 || parts[0] == "" {
			continue
		}

		r := ReplaceRegexRule{
			Pattern: parts[0],
			IsRegex: true,
		}
		if len(parts) > 1 {
			r.Replacement = parts[1]
		}

		rules = append(rules, r)
	}

	return rules
}

// ApplyReplaceRules applies a list of replace rules to content
func ApplyReplaceRules(content string, rules []ReplaceRegexRule) string {
	result := content
	for _, rule := range rules {
		if rule.IsRegex {
			if re, err := regexp.Compile(rule.Pattern); err == nil {
				result = re.ReplaceAllString(result, rule.Replacement)
			}
		} else {
			result = strings.ReplaceAll(result, rule.Pattern, rule.Replacement)
		}
	}
	return result
}
