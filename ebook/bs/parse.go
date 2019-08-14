package bs

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"jaytaylor.com/html2text"
)

// ParseRules parse rules
func ParseRules(doc interface{}, rule string) (*goquery.Selection, string) {
	// log.Debugf("parsing rules:%s\n", rule)
	var sel *goquery.Selection
	var result string
	var exclude = make([]string, 0)

	if strings.HasPrefix(rule, "@JSon") {
		log.Println("json result. not implemented.")
		return nil, ""
	}

	if strings.HasPrefix(rule, "@css") {
		log.Println("jsoup selector.not implemented.")
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
			switch ruleStr {
			case "text":
				result = sel.Text()
			case "html":
				result, _ = sel.Html()
			case "textNodes":
				result, _ = sel.Html()
				text, err := html2text.FromString(result, html2text.Options{PrettyTables: false})
				if err == nil {
					result = text
				}
			case "src", "href":
				result, _ = sel.Attr(ruleStr)
			default:
				sel = sel.Find(ruleStr)
			}
			if result != "" {
				return nil, strings.TrimSpace(result)
				// return nil, result
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
			if index < 0 {
				index += sel.Length()
				// fmt.Printf("index = %d\n", index)
			}
			nodes = append(nodes, sel.Nodes[index])
		}
		sel.Nodes = RemoveNodes(sel.Nodes, nodes)
		// fmt.Printf("total %d after removed.\n", sel.Length())
	}
	return sel, ""
}

// ParseRule return selector,length of rules, index of selector
func ParseRule(rule string) (string, int, int) {
	ruleList := strings.Split(rule, ".")
	var index int
	sel := ""
	if len(ruleList) == 1 {
		return ruleList[0], 1, index
	}
	switch ruleList[0] {
	case "class":
		sel = fmt.Sprintf(".%s", ruleList[1])
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
