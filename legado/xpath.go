package legado

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// XPathParser implements XPath parsing (// or @XPath: prefix)
type XPathParser struct {
	doc     *html.Node
	baseURL string
}

// NewXPathParser creates a new XPath parser from HTML content
func NewXPathParser(content []byte, baseURL string) (*XPathParser, error) {
	doc, err := htmlquery.Parse(strings.NewReader(string(content)))
	if err != nil {
		return nil, err
	}
	return &XPathParser{doc: doc, baseURL: baseURL}, nil
}

// NewXPathParserFromNode creates a parser from an existing node
func NewXPathParserFromNode(node *html.Node, baseURL string) *XPathParser {
	return &XPathParser{doc: node, baseURL: baseURL}
}

// Parse parses an XPath rule
func (p *XPathParser) Parse(rule string) []string {
	return p.ParseFromNode(p.doc, rule)
}

// ParseFromNode parses an XPath rule from a given node
func (p *XPathParser) ParseFromNode(node *html.Node, rule string) []string {
	if node == nil {
		return nil
	}

	// Remove prefix
	xpath := rule
	if strings.HasPrefix(rule, "@XPath:") {
		xpath = strings.TrimPrefix(rule, "@XPath:")
	}
	xpath = strings.TrimSpace(xpath)

	if xpath == "" {
		return nil
	}

	// Handle regex replacement ##pattern##replacement
	var regexPattern, regexReplace string
	if idx := strings.Index(xpath, "##"); idx != -1 {
		regexPart := xpath[idx:]
		xpath = xpath[:idx]
		if matches := regexReplacePattern.FindStringSubmatch(regexPart); matches != nil {
			regexPattern = matches[1]
			if len(matches) > 2 {
				regexReplace = matches[2]
			}
		}
	}

	// Execute XPath query
	nodes, err := htmlquery.QueryAll(node, xpath)
	if err != nil || len(nodes) == 0 {
		return nil
	}

	var results []string
	for _, n := range nodes {
		content := p.extractContent(n)
		if content != "" {
			// Apply regex replacement
			if regexPattern != "" {
				if re, err := regexp.Compile(regexPattern); err == nil {
					content = re.ReplaceAllString(content, regexReplace)
				}
			}
			results = append(results, content)
		}
	}

	return results
}

// extractContent extracts content from an HTML node
func (p *XPathParser) extractContent(node *html.Node) string {
	if node == nil {
		return ""
	}

	switch node.Type {
	case html.TextNode:
		return strings.TrimSpace(node.Data)
	case html.ElementNode:
		return strings.TrimSpace(htmlquery.InnerText(node))
	case html.DocumentNode:
		return strings.TrimSpace(htmlquery.InnerText(node))
	default:
		// For attribute nodes, the value is in node.Data for namespace
		// but htmlquery handles @attr in xpath differently
		if node.Type == html.TextNode {
			return node.Data
		}
		return strings.TrimSpace(htmlquery.InnerText(node))
	}
}

// GetElements returns matched nodes for list processing
func (p *XPathParser) GetElements(rule string) []*html.Node {
	return p.GetElementsFromNode(p.doc, rule)
}

// GetElementsFromNode returns matched nodes from a given node
func (p *XPathParser) GetElementsFromNode(node *html.Node, rule string) []*html.Node {
	if node == nil {
		return nil
	}

	xpath := rule
	if strings.HasPrefix(rule, "@XPath:") {
		xpath = strings.TrimPrefix(rule, "@XPath:")
	}
	xpath = strings.TrimSpace(xpath)

	// Remove regex replacement for element selection
	if idx := strings.Index(xpath, "##"); idx != -1 {
		xpath = xpath[:idx]
	}

	nodes, err := htmlquery.QueryAll(node, xpath)
	if err != nil {
		return nil
	}

	return nodes
}

// GetAttribute gets an attribute value from a node
func (p *XPathParser) GetAttribute(node *html.Node, attrName string) string {
	if node == nil {
		return ""
	}
	return htmlquery.SelectAttr(node, attrName)
}

// resolveURL resolves a relative URL
func (p *XPathParser) resolveURL(href string) string {
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

// ConvertToGoquery converts an html.Node to a goquery.Selection
// This is useful for switching between parsers
func ConvertToGoquery(node *html.Node) *goquery.Selection {
	return goquery.NewDocumentFromNode(node).Selection
}
