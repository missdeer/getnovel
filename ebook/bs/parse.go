package bs

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/jaytaylor/html2text"
	"github.com/missdeer/getnovel/legado"
	"golang.org/x/net/html"
)

// useLegadoParser controls whether to use legado parser for advanced rules
var useLegadoParser = true

// SetUseLegadoParser sets whether to use legado parser
func SetUseLegadoParser(use bool) {
	useLegadoParser = use
}

// isAdvancedRule checks if a rule uses advanced syntax that requires legado parser
func isAdvancedRule(rule string) bool {
	// Check for CSS selector prefix
	if strings.HasPrefix(rule, "@css:") {
		return true
	}
	// Check for JSONPath
	if strings.HasPrefix(rule, "$.") || strings.HasPrefix(rule, "@json:") {
		return true
	}
	// Check for XPath
	if strings.HasPrefix(rule, "//") || strings.HasPrefix(rule, "@XPath:") {
		return true
	}
	// Check for JavaScript
	if strings.HasPrefix(rule, "@js:") || strings.Contains(rule, "<js>") {
		return true
	}
	// Check for combinators
	if strings.Contains(rule, "&&") || strings.Contains(rule, "||") || strings.Contains(rule, "%%") {
		return true
	}
	// Check for templates
	if strings.Contains(rule, "{{") && strings.Contains(rule, "}}") {
		return true
	}
	return false
}

// ParseRulesWithLegado uses legado.RuleAnalyzer for parsing
func ParseRulesWithLegado(doc interface{}, rule string, baseURL string) (*goquery.Selection, string) {
	if rule == "" {
		return nil, ""
	}

	var htmlContent []byte
	var document *goquery.Document

	switch d := doc.(type) {
	case *goquery.Document:
		document = d
		h, _ := d.Html()
		htmlContent = []byte(h)
	case *goquery.Selection:
		document = nil
		h, _ := goquery.OuterHtml(d)
		htmlContent = []byte(h)
	default:
		return nil, ""
	}

	analyzer := legado.NewRuleAnalyzer(htmlContent, baseURL)

	// Get elements for selection-based rules
	elements := analyzer.GetElements(rule)
	if len(elements) > 0 {
		// Wrap elements into a single selection if we have a document
		if document != nil && len(elements) > 0 {
			// Create a combined selection from all elements
			var nodes []*html.Node
			for _, elem := range elements {
				for _, n := range elem.Nodes {
					nodes = append(nodes, n)
				}
			}
			if len(nodes) > 0 {
				return document.FindNodes(nodes...), ""
			}
		}
		return elements[0], ""
	}

	// Get string result
	result := analyzer.GetString(rule)
	return nil, result
}

func ParseRules(doc interface{}, rule string) (*goquery.Selection, string) {
	return ParseRulesWithBaseURL(doc, rule, "")
}

// ParseRulesWithBaseURL parses rules with a base URL for URL resolution
func ParseRulesWithBaseURL(doc interface{}, rule string, baseURL string) (*goquery.Selection, string) {
	// Use legado parser for advanced rules or when enabled
	if useLegadoParser && isAdvancedRule(rule) {
		return ParseRulesWithLegado(doc, rule, baseURL)
	}

	// log.Debugf("parsing rules:%s\n", rule)
	var sel *goquery.Selection
	var result string
	var tmpRule = make([]string, 0)
	var exclude = make([]string, 0)

	if strings.HasPrefix(rule, "@JSon") {
		// log.Error("json result. not implemented.")
		return nil, ""
	}

	if strings.HasPrefix(rule, "@css") {
		// log.Error("jsoup selector.not implemented.")
		return nil, ""
	}
	rules := strings.Split(rule, "@")
	for i, r := range rules {
		ruleStr, length, index := ParseRule(r)

		// 需要排除的情况
		if strings.Contains(ruleStr, "!") {
			splitedStr := strings.Split(ruleStr, "!")
			ruleStr = splitedStr[0]
			exclude = strings.Split(splitedStr[1], ":")
		}
		switch i {
		case 0:
			document, ok := doc.(*goquery.Document)
			if ok {
				sel = document.Find(ruleStr)
			} else {
				sel, _ = doc.(*goquery.Selection)

				sel = sel.Find(ruleStr)
			}

			if length == 3 {
				sel = sel.Eq(index)
			}

		case len(rules) - 1:
			if strings.Contains(ruleStr, "#") {
				tmpRule = strings.Split(ruleStr, "#")
				ruleStr = tmpRule[0]
			}
			switch ruleStr {
			case "text":
				var s []string
				for _, n := range sel.Nodes {
					s = append(s, Nodetext(n))
				}
				result = strings.Join(s, "　　\n")
			case "html":
				result, _ = sel.Html()
			case "textNodes":
				result, _ = sel.Html()
				// log.DebugF("length of sel:%d\n length of children:%d\n", len(sel.Nodes), len(sel.Children().Nodes))
				text, err := html2text.FromString(result, html2text.Options{PrettyTables: false})

				if err == nil {
					s := strings.Split(text, "\n\n")
					for i, v := range s {
						s[i] = fmt.Sprintf("　　%s", strings.TrimSpace(v))
					}
					result = strings.Join(s, "\n")
				}

			case "src", "href":
				result, _ = sel.Attr(ruleStr)
			case "a":
				break
			default:
				// tHtml, _ := sel.Html()
				// log.Debugf("ruleStr is %s.\tsel is: %s.\n", ruleStr, tHtml)
				sel = sel.Find(ruleStr)
				// tHtml, _ = sel.Html()
				// log.Debugf("ruleStr is %s.\tsel is: %s.\n", ruleStr, tHtml)
			}
			if result != "" {
				if len(tmpRule) >= 2 {
					result = strings.Replace(result, tmpRule[1], "", 0)
				}
				return nil, strings.TrimSpace(result)
			}
		default:

			sel = sel.Find(ruleStr)
			if length == 3 {
				sel = sel.Eq(index)
			}
		}
	}
	if len(exclude) != 0 && sel.Length() != 0 {
		// fmt.Printf("total %d. %d needs to be removed.\n", sel.Length(), len(exclude))
		var nodes = make([]*html.Node, 0)
		for _, i := range exclude {
			index, err := strconv.Atoi(i)
			if err != nil {
				fmt.Printf("convert string to int error:%s\n", err.Error())
			}
			if index < 0 { // !是排除,有些位置不符合需要排除用!,后面的序号用:隔开0是第1个,负数为倒数序号,-1最后一个,-2倒数第2个,依次
				index += sel.Length()
			}
			if index < len(sel.Nodes) { // 有时候规则写的不是很准确，排除的节点序号超过实际可用的节点数，会引发越界异常
				nodes = append(nodes, sel.Nodes[index])
			}
		}
		sel.Nodes = RemoveNodes(sel.Nodes, nodes)
	}
	return sel, ""
}

// return selector,length of rules, index of selector
func ParseRule(rule string) (string, int, int) {
	ruleList := strings.Split(rule, ".")
	var index int
	sel := ""
	if len(ruleList) == 1 {
		return ruleList[0], 1, index
	}
	switch ruleList[0] {
	case "class":
		if strings.Contains(ruleList[1], " ") { // 多个class name的情况
			var s = ""
			for _, v := range strings.Split(ruleList[1], " ") {
				v = strings.TrimSpace(v)
				if v != "" {
					s = fmt.Sprintf("%s.%s", s, v)
				}
			}
			sel = s
		} else {
			sel = fmt.Sprintf(".%s", ruleList[1])
		}

	case "tag":
		sel = ruleList[1]
	case "id":
		sel = fmt.Sprintf("#%s", ruleList[1])
	}
	if len(ruleList) == 3 {
		index, _ = strconv.Atoi(ruleList[2])
	}
	return sel, len(ruleList), index
}

func RemoveNodes(srcNodes, removeNodes []*html.Node) []*html.Node {
	var nodes = make([]*html.Node, 0)
	for _, n := range srcNodes {
		found := false
		for _, rn := range removeNodes {
			if reflect.DeepEqual(n, rn) {
				found = true
			}
		}
		if !found {
			nodes = append(nodes, n)
		}
	}
	return nodes
}

func Nodetext(node *html.Node) string {
	if node.Type == html.TextNode {
		// Keep newlines and spaces, like jQuery
		return node.Data
	} else if node.FirstChild != nil {
		var buf bytes.Buffer
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			buf.WriteString(Nodetext(c))
		}
		return buf.String()
	}

	return ""
}
