package legado

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// RuleAnalyzer provides a unified interface for parsing rules
type RuleAnalyzer struct {
	content    []byte
	contentStr string
	baseURL    string
	doc        *goquery.Document
	htmlNode   *html.Node
	jsEngine   *JSEngine
	isJSON     bool
}

// NewRuleAnalyzer creates a new rule analyzer
func NewRuleAnalyzer(content []byte, baseURL string) *RuleAnalyzer {
	ra := &RuleAnalyzer{
		content:    content,
		contentStr: string(content),
		baseURL:    baseURL,
		isJSON:     IsValidJSON(content),
	}

	// Initialize goquery document for HTML parsing
	if !ra.isJSON {
		if doc, err := goquery.NewDocumentFromReader(strings.NewReader(ra.contentStr)); err == nil {
			ra.doc = doc
		}
	}

	return ra
}

// SetJSEngine sets the JavaScript engine
func (ra *RuleAnalyzer) SetJSEngine(engine *JSEngine) {
	ra.jsEngine = engine
	if ra.jsEngine != nil {
		ra.jsEngine.SetAnalyzer(ra)
	}
}

// GetJSEngine gets or creates the JavaScript engine
func (ra *RuleAnalyzer) GetJSEngine() *JSEngine {
	if ra.jsEngine == nil {
		ra.jsEngine = NewJSEngine(ra.baseURL)
		ra.jsEngine.SetResult(ra.contentStr)
		ra.jsEngine.SetAnalyzer(ra)
	}
	return ra.jsEngine
}

// ParseRule parses a rule and returns results
func (ra *RuleAnalyzer) ParseRule(content []byte, rule string) []string {
	if rule == "" {
		return nil
	}

	rule = strings.TrimSpace(rule)

	// Handle rule combinators first
	parts, combinator := SplitRuleByCombinator(rule)
	if combinator != CombinatorNone && len(parts) > 1 {
		return ra.applyCombinator(content, parts, combinator)
	}

	// Handle multi-step rules (separated by newline)
	steps := SplitByNewlineAndJS(rule)
	if len(steps) > 1 {
		return ra.processMultiStepRule(content, steps)
	}

	// Parse the single rule
	return ra.parseSingleRule(content, rule)
}

// parseSingleRule parses a single rule without combinators
func (ra *RuleAnalyzer) parseSingleRule(content []byte, rule string) []string {
	// Parse for regex replacement
	parsed := ParseRule(rule)

	// Determine content type and get results
	var results []string

	contentStr := string(content)
	ruleType := parsed.Type

	switch ruleType {
	case RuleTypeCSS:
		results = ra.parseCSS(content, parsed.Rule)
	case RuleTypeXPath:
		results = ra.parseXPath(content, parsed.Rule)
	case RuleTypeJSONPath:
		results = ra.parseJSONPath(contentStr, parsed.Rule)
	case RuleTypeJS:
		results = ra.parseJS(contentStr, parsed.Rule)
	case RuleTypeRegex:
		results = ra.parseRegex(contentStr, parsed.Rule)
	default:
		results = ra.parseJsoup(content, parsed.Rule)
	}

	// Apply regex replacement if specified
	if parsed.HasReplacement && len(results) > 0 {
		re, err := regexp.Compile(parsed.RegexPattern)
		if err == nil {
			for i, r := range results {
				results[i] = re.ReplaceAllString(r, parsed.RegexReplace)
			}
		}
	}

	return results
}

// parseCSS parses CSS selector rules
func (ra *RuleAnalyzer) parseCSS(content []byte, rule string) []string {
	doc := ra.doc
	if doc == nil {
		var err error
		doc, err = goquery.NewDocumentFromReader(strings.NewReader(string(content)))
		if err != nil {
			return nil
		}
	}

	parser := NewCSSParser(doc, ra.baseURL)
	return parser.Parse(rule)
}

// parseXPath parses XPath rules
func (ra *RuleAnalyzer) parseXPath(content []byte, rule string) []string {
	parser, err := NewXPathParser(content, ra.baseURL)
	if err != nil {
		return nil
	}
	return parser.Parse(rule)
}

// parseJSONPath parses JSONPath rules
func (ra *RuleAnalyzer) parseJSONPath(content string, rule string) []string {
	parser := NewJSONPathParserFromString(content, ra.baseURL)
	return parser.Parse(rule)
}

// parseJS parses JavaScript rules
func (ra *RuleAnalyzer) parseJS(content string, rule string) []string {
	engine := ra.GetJSEngine()
	engine.SetResult(content)

	result, err := engine.ProcessJSRule(rule, content)
	if err != nil {
		return nil
	}

	if result != "" {
		return []string{result}
	}
	return nil
}

// parseJsoup parses JSOUP default rules
func (ra *RuleAnalyzer) parseJsoup(content []byte, rule string) []string {
	doc := ra.doc
	if doc == nil {
		var err error
		doc, err = goquery.NewDocumentFromReader(strings.NewReader(string(content)))
		if err != nil {
			return nil
		}
	}

	parser := NewJsoupParser(doc, ra.baseURL)
	return parser.Parse(rule)
}

// parseRegex parses regex AllInOne rules
func (ra *RuleAnalyzer) parseRegex(content string, rule string) []string {
	parser := NewRegexParser(content, ra.baseURL)
	matches := parser.ParseAllInOne(rule)

	var results []string
	for _, match := range matches {
		if len(match) > 0 {
			// Return full match or first capture group
			if len(match) > 1 {
				results = append(results, match[1])
			} else {
				results = append(results, match[0])
			}
		}
	}
	return results
}

// applyCombinator applies rule combinators
func (ra *RuleAnalyzer) applyCombinator(content []byte, parts []string, combinator RuleCombinator) []string {
	switch combinator {
	case CombinatorAnd:
		// Merge all results
		var allResults []string
		for _, part := range parts {
			results := ra.parseSingleRule(content, part)
			allResults = append(allResults, results...)
		}
		return allResults

	case CombinatorOr:
		// Return first non-empty result
		for _, part := range parts {
			results := ra.parseSingleRule(content, part)
			if len(results) > 0 {
				return results
			}
		}
		return nil

	case CombinatorPercent:
		// Interleave results
		var allPartResults [][]string
		maxLen := 0
		for _, part := range parts {
			results := ra.parseSingleRule(content, part)
			allPartResults = append(allPartResults, results)
			if len(results) > maxLen {
				maxLen = len(results)
			}
		}

		var interleaved []string
		for i := 0; i < maxLen; i++ {
			for _, partResults := range allPartResults {
				if i < len(partResults) {
					interleaved = append(interleaved, partResults[i])
				}
			}
		}
		return interleaved
	}

	return nil
}

// processMultiStepRule processes rules with multiple steps
func (ra *RuleAnalyzer) processMultiStepRule(content []byte, steps []string) []string {
	currentContent := string(content)

	for i, step := range steps {
		step = strings.TrimSpace(step)
		if step == "" {
			continue
		}

		// Check if this is a JS step
		if strings.HasPrefix(step, "@js:") || strings.HasPrefix(step, "<js>") {
			engine := ra.GetJSEngine()
			engine.SetResult(currentContent)

			result, err := engine.ProcessJSRule(step, currentContent)
			if err != nil {
				return nil
			}
			currentContent = result
		} else {
			// Regular rule step
			results := ra.parseSingleRule([]byte(currentContent), step)
			if len(results) == 0 {
				return nil
			}

			// If not last step, use first result as content for next step
			if i < len(steps)-1 {
				currentContent = results[0]
			} else {
				return results
			}
		}
	}

	if currentContent != "" {
		return []string{currentContent}
	}
	return nil
}

// GetElements returns elements for list processing
func (ra *RuleAnalyzer) GetElements(rule string) []*goquery.Selection {
	if ra.doc == nil {
		return nil
	}

	rule = strings.TrimSpace(rule)
	ruleType := DetectRuleType(rule)

	switch ruleType {
	case RuleTypeCSS:
		parser := NewCSSParser(ra.doc, ra.baseURL)
		return parser.GetElements(rule)
	default:
		parser := NewJsoupParser(ra.doc, ra.baseURL)
		return parser.GetElements(rule)
	}
}

// ParseFromSelection parses a rule from a goquery Selection
func (ra *RuleAnalyzer) ParseFromSelection(sel *goquery.Selection, rule string) []string {
	if sel == nil || rule == "" {
		return nil
	}

	rule = strings.TrimSpace(rule)

	// Handle templates
	if HasTemplate(rule) {
		engine := ra.GetJSEngine()
		html, _ := goquery.OuterHtml(sel)
		engine.SetResult(html)

		result, err := engine.ProcessTemplate(rule, nil)
		if err != nil {
			return nil
		}
		return []string{result}
	}

	// Parse the rule
	parsed := ParseRule(rule)
	var results []string

	switch parsed.Type {
	case RuleTypeCSS:
		parser := NewCSSParser(nil, ra.baseURL)
		results = parser.ParseFromSelection(sel, parsed.Rule)
	case RuleTypeJS:
		html, _ := goquery.OuterHtml(sel)
		engine := ra.GetJSEngine()
		engine.SetResult(html)
		result, err := engine.ProcessJSRule(parsed.Rule, html)
		if err == nil && result != "" {
			results = []string{result}
		}
	default:
		parser := NewJsoupParser(nil, ra.baseURL)
		results = parser.ParseFromSelection(sel, parsed.Rule)
	}

	// Apply regex replacement
	if parsed.HasReplacement && len(results) > 0 {
		re, err := regexp.Compile(parsed.RegexPattern)
		if err == nil {
			for i, r := range results {
				results[i] = re.ReplaceAllString(r, parsed.RegexReplace)
			}
		}
	}

	return results
}

// GetString returns the first result of a rule
func (ra *RuleAnalyzer) GetString(rule string) string {
	results := ra.ParseRule(ra.content, rule)
	if len(results) > 0 {
		return results[0]
	}
	return ""
}

// GetStringList returns all results of a rule
func (ra *RuleAnalyzer) GetStringList(rule string) []string {
	return ra.ParseRule(ra.content, rule)
}
