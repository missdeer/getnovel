package legado

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// CSSParser implements CSS selector parsing (@css: prefix)
type CSSParser struct {
	doc     *goquery.Document
	baseURL string
}

// NewCSSParser creates a new CSS parser
func NewCSSParser(doc *goquery.Document, baseURL string) *CSSParser {
	return &CSSParser{doc: doc, baseURL: baseURL}
}

// Parse parses a CSS rule (must start with @css:)
func (p *CSSParser) Parse(rule string) []string {
	return p.ParseFromSelection(p.doc.Selection, rule)
}

// ParseFromSelection parses a CSS rule from a given selection
func (p *CSSParser) ParseFromSelection(sel *goquery.Selection, rule string) []string {
	if sel == nil || sel.Length() == 0 {
		return nil
	}

	// Remove @css: prefix
	rule = strings.TrimPrefix(rule, "@css:")
	rule = strings.TrimSpace(rule)

	if rule == "" {
		return nil
	}

	// Check for content extraction suffix @attr
	var contentType string
	if idx := strings.LastIndex(rule, "@"); idx != -1 {
		contentType = rule[idx+1:]
		rule = rule[:idx]
	} else {
		contentType = "text"
	}

	// Handle regex replacement ##pattern##replacement
	var regexPattern, regexReplace string
	if idx := strings.Index(contentType, "##"); idx != -1 {
		regexPart := contentType[idx:]
		contentType = contentType[:idx]
		if matches := regexReplacePattern.FindStringSubmatch(regexPart); matches != nil {
			regexPattern = matches[1]
			if len(matches) > 2 {
				regexReplace = matches[2]
			}
		}
	}

	// Find elements
	elements := sel.Find(rule)
	if elements.Length() == 0 {
		return nil
	}

	// Extract content
	var results []string
	elements.Each(func(i int, s *goquery.Selection) {
		content := p.extractContent(s, contentType)
		if content != "" {
			// Apply regex replacement if specified
			if regexPattern != "" {
				if re, err := regexp.Compile(regexPattern); err == nil {
					content = re.ReplaceAllString(content, regexReplace)
				}
			}
			results = append(results, content)
		}
	})

	return results
}

// extractContent extracts content based on content type
func (p *CSSParser) extractContent(s *goquery.Selection, contentType string) string {
	switch contentType {
	case "text":
		return strings.TrimSpace(s.Text())
	case "textNodes":
		return p.getTextNodes(s)
	case "ownText":
		return p.getOwnText(s)
	case "html":
		html, _ := s.Html()
		return html
	case "all":
		html, _ := goquery.OuterHtml(s)
		return html
	case "href":
		if href, exists := s.Attr("href"); exists {
			return p.resolveURL(href)
		}
		return ""
	case "src":
		if src, exists := s.Attr("src"); exists {
			return p.resolveURL(src)
		}
		return ""
	default:
		// Try as attribute
		if attr, exists := s.Attr(contentType); exists {
			if contentType == "href" || contentType == "src" || strings.HasSuffix(contentType, "-src") {
				return p.resolveURL(attr)
			}
			return attr
		}
		return strings.TrimSpace(s.Text())
	}
}

// getTextNodes gets text from direct text nodes only
func (p *CSSParser) getTextNodes(s *goquery.Selection) string {
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
func (p *CSSParser) getOwnText(s *goquery.Selection) string {
	clone := s.Clone()
	clone.Children().Remove()
	return strings.TrimSpace(clone.Text())
}

// resolveURL resolves a relative URL against the base URL
func (p *CSSParser) resolveURL(href string) string {
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
		if p.baseURL != "" {
			if matches := baseURLPattern.FindStringSubmatch(p.baseURL); len(matches) > 1 {
				return matches[1] + href
			}
		}
		return href
	}
	if p.baseURL != "" {
		base := p.baseURL
		if idx := strings.LastIndex(base, "/"); idx != -1 {
			base = base[:idx+1]
		}
		return base + href
	}
	return href
}

// GetElements returns matched elements for list processing
func (p *CSSParser) GetElements(rule string) []*goquery.Selection {
	return p.GetElementsFromSelection(p.doc.Selection, rule)
}

// GetElementsFromSelection returns matched elements from a selection
func (p *CSSParser) GetElementsFromSelection(sel *goquery.Selection, rule string) []*goquery.Selection {
	if sel == nil || sel.Length() == 0 {
		return nil
	}

	rule = strings.TrimPrefix(rule, "@css:")
	rule = strings.TrimSpace(rule)

	// Remove content extraction suffix for element selection
	if idx := strings.LastIndex(rule, "@"); idx != -1 {
		suffix := rule[idx+1:]
		// Only remove if it looks like a content type, not a CSS pseudo-selector
		if isContentType(suffix) || strings.Contains(suffix, "##") {
			rule = rule[:idx]
		}
	}

	elements := sel.Find(rule)
	var results []*goquery.Selection
	elements.Each(func(i int, s *goquery.Selection) {
		results = append(results, s)
	})
	return results
}
