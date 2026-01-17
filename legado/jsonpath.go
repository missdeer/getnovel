package legado

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
)

// JSONPathParser implements JSONPath parsing ($. or @json: prefix)
type JSONPathParser struct {
	jsonData string
	baseURL  string
}

// NewJSONPathParser creates a new JSONPath parser
func NewJSONPathParser(data []byte, baseURL string) *JSONPathParser {
	return &JSONPathParser{
		jsonData: string(data),
		baseURL:  baseURL,
	}
}

// NewJSONPathParserFromString creates a parser from a JSON string
func NewJSONPathParserFromString(data string, baseURL string) *JSONPathParser {
	return &JSONPathParser{
		jsonData: data,
		baseURL:  baseURL,
	}
}

// Parse parses a JSONPath rule
func (p *JSONPathParser) Parse(rule string) []string {
	return p.ParseFromJSON(p.jsonData, rule)
}

// ParseFromJSON parses a JSONPath rule from JSON data
func (p *JSONPathParser) ParseFromJSON(jsonData string, rule string) []string {
	if jsonData == "" {
		return nil
	}

	// Remove prefix
	path := rule
	if strings.HasPrefix(rule, "@json:") {
		path = strings.TrimPrefix(rule, "@json:")
	}
	path = strings.TrimSpace(path)

	// Convert JSONPath syntax to gjson syntax
	// $. prefix is used in JSONPath, gjson doesn't need it
	if strings.HasPrefix(path, "$.") {
		path = strings.TrimPrefix(path, "$.")
	}
	// JSONPath uses [*] or .# for array iteration, gjson uses #
	path = strings.ReplaceAll(path, "[*]", ".#")

	if path == "" {
		return nil
	}

	// Handle regex replacement ##pattern##replacement
	var regexPattern, regexReplace string
	if idx := strings.Index(path, "##"); idx != -1 {
		regexPart := path[idx:]
		path = path[:idx]
		if matches := regexReplacePattern.FindStringSubmatch(regexPart); matches != nil {
			regexPattern = matches[1]
			if len(matches) > 2 {
				regexReplace = matches[2]
			}
		}
	}

	// Use gjson to parse
	result := gjson.Get(jsonData, path)
	if !result.Exists() {
		return nil
	}

	var values []string
	if result.IsArray() {
		result.ForEach(func(key, value gjson.Result) bool {
			s := p.resultToString(value)
			if s != "" {
				values = append(values, s)
			}
			return true
		})
	} else {
		s := p.resultToString(result)
		if s != "" {
			values = append(values, s)
		}
	}

	// Apply regex replacement
	if regexPattern != "" {
		re, err := regexp.Compile(regexPattern)
		if err == nil {
			for i, v := range values {
				values[i] = re.ReplaceAllString(v, regexReplace)
			}
		}
	}

	return values
}

// resultToString converts a gjson.Result to string
func (p *JSONPathParser) resultToString(r gjson.Result) string {
	switch r.Type {
	case gjson.String:
		return r.String()
	case gjson.Number:
		return r.Raw
	case gjson.True:
		return "true"
	case gjson.False:
		return "false"
	case gjson.Null:
		return ""
	case gjson.JSON:
		return r.Raw
	default:
		return r.String()
	}
}

// GetElements returns JSON objects/arrays for list processing
func (p *JSONPathParser) GetElements(rule string) []string {
	path := rule
	if strings.HasPrefix(rule, "@json:") {
		path = strings.TrimPrefix(rule, "@json:")
	}
	path = strings.TrimSpace(path)

	// Convert JSONPath syntax to gjson syntax
	if strings.HasPrefix(path, "$.") {
		path = strings.TrimPrefix(path, "$.")
	}
	path = strings.ReplaceAll(path, "[*]", ".#")

	// Remove any content extraction suffix
	if idx := strings.Index(path, "##"); idx != -1 {
		path = path[:idx]
	}

	result := gjson.Get(p.jsonData, path)
	if !result.Exists() {
		return nil
	}

	var elements []string
	if result.IsArray() {
		result.ForEach(func(key, value gjson.Result) bool {
			elements = append(elements, value.Raw)
			return true
		})
	} else if result.IsObject() {
		elements = append(elements, result.Raw)
	}

	return elements
}

// ParseValue extracts a single value from a JSON element
func (p *JSONPathParser) ParseValue(jsonElement string, path string) string {
	if jsonElement == "" || path == "" {
		return ""
	}

	// Remove prefix
	if strings.HasPrefix(path, "@json:") {
		path = strings.TrimPrefix(path, "@json:")
	}
	if strings.HasPrefix(path, "$.") {
		path = strings.TrimPrefix(path, "$.")
	}
	path = strings.TrimSpace(path)

	// Handle regex replacement
	var regexPattern, regexReplace string
	if idx := strings.Index(path, "##"); idx != -1 {
		regexPart := path[idx:]
		path = path[:idx]
		if matches := regexReplacePattern.FindStringSubmatch(regexPart); matches != nil {
			regexPattern = matches[1]
			if len(matches) > 2 {
				regexReplace = matches[2]
			}
		}
	}

	result := gjson.Get(jsonElement, path)
	if !result.Exists() {
		return ""
	}

	value := p.resultToString(result)

	if regexPattern != "" {
		if re, err := regexp.Compile(regexPattern); err == nil {
			value = re.ReplaceAllString(value, regexReplace)
		}
	}

	return value
}

// IsValidJSON checks if data is valid JSON
func IsValidJSON(data []byte) bool {
	return json.Valid(data)
}
