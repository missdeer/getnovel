package legado

import (
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// JsoupParser implements the JSOUP Default rule parser
// Syntax: segment1@segment2@segment3@...@contentType
// Each segment: type.name.position or type.name!excludePos
// Types: class, id, tag, text, children
// Content types: text, textNodes, ownText, href, src, html, all
type JsoupParser struct {
	doc     *goquery.Document
	baseURL string
}

// NewJsoupParser creates a new JSOUP parser
func NewJsoupParser(doc *goquery.Document, baseURL string) *JsoupParser {
	return &JsoupParser{doc: doc, baseURL: baseURL}
}

// ParseFromSelection parses a rule starting from a selection
func (p *JsoupParser) ParseFromSelection(sel *goquery.Selection, rule string) []string {
	if rule == "" || sel == nil || sel.Length() == 0 {
		return nil
	}

	// Check for reverse flag
	reverse := false
	if strings.HasPrefix(rule, "-") {
		reverse = true
		rule = rule[1:]
	}

	// Split by @ to get segments
	segments := strings.Split(rule, "@")
	if len(segments) == 0 {
		return nil
	}

	// Process each segment except the last (which is content extraction)
	current := sel
	for i := 0; i < len(segments)-1; i++ {
		current = p.applySegment(current, segments[i])
		if current == nil || current.Length() == 0 {
			return nil
		}
	}

	// Last segment is the content extraction method
	contentType := segments[len(segments)-1]

	// Check if it's a complex segment (contains .) that's not a content type
	if strings.Contains(contentType, ".") && !isContentType(contentType) {
		// It's another selector segment, extract as text by default
		current = p.applySegment(current, contentType)
		contentType = "text"
	}

	results := p.extractContent(current, contentType)

	if reverse {
		// Reverse the results
		for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
			results[i], results[j] = results[j], results[i]
		}
	}

	return results
}

// Parse parses a rule on the document
func (p *JsoupParser) Parse(rule string) []string {
	return p.ParseFromSelection(p.doc.Selection, rule)
}

// applySegment applies a single selector segment to the current selection
func (p *JsoupParser) applySegment(sel *goquery.Selection, segment string) *goquery.Selection {
	segment = strings.TrimSpace(segment)
	if segment == "" {
		return sel
	}

	// Parse segment: type.name.position or type.name!exclude1:exclude2
	parts := strings.SplitN(segment, ".", 3)
	if len(parts) == 0 {
		return sel
	}

	selectorType := parts[0]
	var name string
	var position string
	var excludes []int

	if len(parts) > 1 {
		// Check for exclusion syntax (name!pos1:pos2)
		if idx := strings.Index(parts[1], "!"); idx != -1 {
			name = parts[1][:idx]
			excludeStr := parts[1][idx+1:]
			excludeParts := strings.Split(excludeStr, ":")
			for _, e := range excludeParts {
				if n, err := strconv.Atoi(e); err == nil {
					excludes = append(excludes, n)
				}
			}
		} else {
			name = parts[1]
		}
	}

	if len(parts) > 2 {
		position = parts[2]
	}

	// Apply the selector based on type
	var result *goquery.Selection
	switch selectorType {
	case "class":
		if name != "" {
			// Handle class names with spaces (multiple classes)
			className := strings.ReplaceAll(name, " ", ".")
			result = sel.Find("." + className)
		} else {
			result = sel
		}
	case "id":
		if name != "" {
			result = sel.Find("#" + name)
		} else {
			result = sel
		}
	case "tag":
		if name != "" {
			result = sel.Find(name)
		} else {
			result = sel
		}
	case "text":
		// Filter by text content
		if name != "" {
			result = sel.FilterFunction(func(i int, s *goquery.Selection) bool {
				return strings.Contains(s.Text(), name)
			})
		} else {
			result = sel
		}
	case "children":
		result = sel.Children()
	default:
		// Try as a CSS selector
		result = sel.Find(segment)
	}

	if result == nil || result.Length() == 0 {
		return nil
	}

	// Apply position filter
	if position != "" {
		result = p.applyPosition(result, position, excludes)
	} else if len(excludes) > 0 {
		result = p.applyExcludes(result, excludes)
	}

	return result
}

// applyPosition applies position filtering to a selection
func (p *JsoupParser) applyPosition(sel *goquery.Selection, position string, excludes []int) *goquery.Selection {
	pos, err := strconv.Atoi(position)
	if err != nil {
		return sel
	}

	length := sel.Length()
	if length == 0 {
		return nil
	}

	// Handle negative positions (from end)
	if pos < 0 {
		pos = length + pos
	}

	if pos < 0 || pos >= length {
		return nil
	}

	return sel.Eq(pos)
}

// applyExcludes removes elements at excluded positions
func (p *JsoupParser) applyExcludes(sel *goquery.Selection, excludes []int) *goquery.Selection {
	length := sel.Length()
	excludeMap := make(map[int]bool)

	for _, e := range excludes {
		pos := e
		if pos < 0 {
			pos = length + pos
		}
		if pos >= 0 && pos < length {
			excludeMap[pos] = true
		}
	}

	return sel.FilterFunction(func(i int, s *goquery.Selection) bool {
		return !excludeMap[i]
	})
}

// extractContent extracts content from selection based on content type
func (p *JsoupParser) extractContent(sel *goquery.Selection, contentType string) []string {
	var results []string

	sel.Each(func(i int, s *goquery.Selection) {
		var content string
		switch contentType {
		case "text":
			content = strings.TrimSpace(s.Text())
		case "textNodes":
			// Get only direct text nodes
			content = p.getTextNodes(s)
		case "ownText":
			// Get own text, excluding children's text
			content = p.getOwnText(s)
		case "href":
			if href, exists := s.Attr("href"); exists {
				content = p.resolveURL(href)
			}
		case "src":
			if src, exists := s.Attr("src"); exists {
				content = p.resolveURL(src)
			}
		case "html":
			content, _ = s.Html()
		case "all":
			content, _ = goquery.OuterHtml(s)
		default:
			// Check if it's an attribute
			if attr, exists := s.Attr(contentType); exists {
				content = attr
				// Resolve URLs for common URL attributes
				if contentType == "href" || contentType == "src" || strings.HasSuffix(contentType, "-src") {
					content = p.resolveURL(content)
				}
			} else {
				content = strings.TrimSpace(s.Text())
			}
		}

		if content != "" {
			results = append(results, content)
		}
	})

	return results
}

// getTextNodes gets text from direct text nodes only
func (p *JsoupParser) getTextNodes(s *goquery.Selection) string {
	var texts []string
	s.Contents().Each(func(i int, c *goquery.Selection) {
		if goquery.NodeName(c) == "#text" {
			if text := strings.TrimSpace(c.Text()); text != "" {
				texts = append(texts, text)
			}
		}
	})
	return strings.Join(texts, "\n")
}

// getOwnText gets own text excluding children
func (p *JsoupParser) getOwnText(s *goquery.Selection) string {
	clone := s.Clone()
	clone.Children().Remove()
	return strings.TrimSpace(clone.Text())
}

// resolveURL resolves a relative URL against the base URL
func (p *JsoupParser) resolveURL(href string) string {
	if href == "" {
		return ""
	}
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	if strings.HasPrefix(href, "//") {
		return "https:" + href
	}
	if strings.HasPrefix(href, "/") {
		// Extract base domain
		if p.baseURL != "" {
			if matches := baseURLPattern.FindStringSubmatch(p.baseURL); len(matches) > 1 {
				return matches[1] + href
			}
		}
		return href
	}
	// Relative URL
	if p.baseURL != "" {
		base := p.baseURL
		if idx := strings.LastIndex(base, "/"); idx != -1 {
			base = base[:idx+1]
		}
		return base + href
	}
	return href
}

// isContentType checks if a string is a content extraction type
func isContentType(s string) bool {
	switch s {
	case "text", "textNodes", "ownText", "href", "src", "html", "all":
		return true
	default:
		return false
	}
}

// GetElements returns all matched elements as selections for list processing
func (p *JsoupParser) GetElements(rule string) []*goquery.Selection {
	if rule == "" {
		return nil
	}

	// Check for reverse flag
	reverse := false
	if strings.HasPrefix(rule, "-") {
		reverse = true
		rule = rule[1:]
	}

	// Split by @ and process all but last as selectors
	segments := strings.Split(rule, "@")

	// Find how many segments are selectors vs content type
	selectorEnd := len(segments)
	if len(segments) > 0 {
		last := segments[len(segments)-1]
		if isContentType(last) || !strings.Contains(last, ".") {
			selectorEnd = len(segments) - 1
			if selectorEnd == 0 {
				selectorEnd = len(segments)
			}
		}
	}

	current := p.doc.Selection
	for i := 0; i < selectorEnd; i++ {
		current = p.applySegment(current, segments[i])
		if current == nil || current.Length() == 0 {
			return nil
		}
	}

	var results []*goquery.Selection
	current.Each(func(i int, s *goquery.Selection) {
		results = append(results, s)
	})

	if reverse {
		for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
			results[i], results[j] = results[j], results[i]
		}
	}

	return results
}
