package legado

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/missdeer/golib/httputil"
)

// JSEngine wraps goja runtime with Legado-compatible java.* methods
type JSEngine struct {
	vm       *goja.Runtime
	baseURL  string
	result   string
	cookie   string
	headers  map[string]string
	timeout  time.Duration
	jsLib    string // jsLib from book source
	analyzer *RuleAnalyzer
}

// NewJSEngine creates a new JavaScript engine
func NewJSEngine(baseURL string) *JSEngine {
	vm := goja.New()
	e := &JSEngine{
		vm:      vm,
		baseURL: baseURL,
		timeout: 30 * time.Second,
		headers: make(map[string]string),
	}
	e.registerMethods()
	return e
}

// SetResult sets the result variable (page content)
func (e *JSEngine) SetResult(result string) {
	e.result = result
	e.vm.Set("result", result)
}

// SetBaseURL sets the baseUrl variable
func (e *JSEngine) SetBaseURL(baseURL string) {
	e.baseURL = baseURL
	e.vm.Set("baseUrl", baseURL)
}

// SetVariables sets common variables
func (e *JSEngine) SetVariables(vars map[string]interface{}) {
	for k, v := range vars {
		e.vm.Set(k, v)
	}
}

// SetJsLib sets the jsLib code to be loaded
func (e *JSEngine) SetJsLib(jsLib string) {
	e.jsLib = jsLib
}

// SetAnalyzer sets the rule analyzer for java.getString etc
func (e *JSEngine) SetAnalyzer(analyzer *RuleAnalyzer) {
	e.analyzer = analyzer
}

// SetCookie sets the cookie for HTTP requests
func (e *JSEngine) SetCookie(cookie string) {
	e.cookie = cookie
}

// SetTimeout sets the HTTP request timeout
func (e *JSEngine) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

// registerMethods registers the java.* methods
func (e *JSEngine) registerMethods() {
	java := e.vm.NewObject()

	// java.ajax(url) - fetch URL content
	java.Set("ajax", func(call goja.FunctionCall) goja.Value {
		urlStr := call.Argument(0).String()
		content, err := e.fetchURL(urlStr)
		if err != nil {
			return goja.Null()
		}
		return e.vm.ToValue(content)
	})

	// java.base64Encode(str) / java.base64Encode(str, flags)
	java.Set("base64Encode", func(call goja.FunctionCall) goja.Value {
		str := call.Argument(0).String()
		encoded := base64.StdEncoding.EncodeToString([]byte(str))
		return e.vm.ToValue(encoded)
	})

	// java.base64Decode(str)
	java.Set("base64Decode", func(call goja.FunctionCall) goja.Value {
		str := call.Argument(0).String()
		decoded, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			// Try URL-safe base64
			decoded, err = base64.URLEncoding.DecodeString(str)
			if err != nil {
				return e.vm.ToValue("")
			}
		}
		return e.vm.ToValue(string(decoded))
	})

	// java.md5Encode(str) - 32-bit MD5
	java.Set("md5Encode", func(call goja.FunctionCall) goja.Value {
		str := call.Argument(0).String()
		hash := md5.Sum([]byte(str))
		return e.vm.ToValue(hex.EncodeToString(hash[:]))
	})

	// java.md5Encode16(str) - 16-bit MD5
	java.Set("md5Encode16", func(call goja.FunctionCall) goja.Value {
		str := call.Argument(0).String()
		hash := md5.Sum([]byte(str))
		return e.vm.ToValue(hex.EncodeToString(hash[4:12]))
	})

	// java.timeFormat(timestamp) - format timestamp to yyyy/MM/dd HH:mm
	java.Set("timeFormat", func(call goja.FunctionCall) goja.Value {
		ts := call.Argument(0).ToInteger()
		t := time.Unix(ts, 0)
		return e.vm.ToValue(t.Format("2006/01/02 15:04"))
	})

	// java.getString(rule) - parse rule from result
	java.Set("getString", func(call goja.FunctionCall) goja.Value {
		rule := call.Argument(0).String()
		var content string
		if len(call.Arguments) > 1 {
			content = call.Argument(1).String()
		} else {
			content = e.result
		}

		if e.analyzer != nil {
			results := e.analyzer.ParseRule([]byte(content), rule)
			if len(results) > 0 {
				return e.vm.ToValue(strings.Join(results, "\n"))
			}
		}
		return e.vm.ToValue("")
	})

	// java.getStringList(rule, isUrl) - parse rule and return list
	java.Set("getStringList", func(call goja.FunctionCall) goja.Value {
		rule := call.Argument(0).String()

		if e.analyzer != nil {
			results := e.analyzer.ParseRule([]byte(e.result), rule)
			arr := e.vm.NewArray()
			for i, r := range results {
				arr.Set(fmt.Sprintf("%d", i), r)
			}
			return arr
		}
		return e.vm.NewArray()
	})

	// java.getElements(rule) - get elements matching rule
	java.Set("getElements", func(call goja.FunctionCall) goja.Value {
		// Returns elements for iteration - implementation depends on context
		return e.vm.NewArray()
	})

	// java.put(key, value) - store variable
	vars := make(map[string]interface{})
	java.Set("put", func(call goja.FunctionCall) goja.Value {
		key := call.Argument(0).String()
		value := call.Argument(1).Export()
		vars[key] = value
		return goja.Undefined()
	})

	// java.get(key) - retrieve variable
	java.Set("get", func(call goja.FunctionCall) goja.Value {
		key := call.Argument(0).String()
		if v, ok := vars[key]; ok {
			return e.vm.ToValue(v)
		}
		return goja.Null()
	})

	// java.log(msg) - log message
	java.Set("log", func(call goja.FunctionCall) goja.Value {
		// Silent in production, can be hooked for debugging
		return goja.Undefined()
	})

	// java.toast(msg) / java.longToast(msg) - show message (no-op in CLI)
	java.Set("toast", func(call goja.FunctionCall) goja.Value {
		return goja.Undefined()
	})
	java.Set("longToast", func(call goja.FunctionCall) goja.Value {
		return goja.Undefined()
	})

	// java.setContent(content, baseUrl) - set content for parsing
	java.Set("setContent", func(call goja.FunctionCall) goja.Value {
		content := call.Argument(0).String()
		if len(call.Arguments) > 1 {
			e.baseURL = call.Argument(1).String()
		}
		e.result = content
		e.vm.Set("result", content)
		return goja.Undefined()
	})

	// java.getWebViewUA() - return a mobile user agent
	java.Set("getWebViewUA", func(call goja.FunctionCall) goja.Value {
		ua := "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Mobile Safari/537.36"
		return e.vm.ToValue(ua)
	})

	// java.t2s(str) - Traditional to Simplified Chinese conversion (placeholder)
	java.Set("t2s", func(call goja.FunctionCall) goja.Value {
		// TODO: implement actual conversion
		return call.Argument(0)
	})

	// java.s2t(str) - Simplified to Traditional Chinese conversion (placeholder)
	java.Set("s2t", func(call goja.FunctionCall) goja.Value {
		// TODO: implement actual conversion
		return call.Argument(0)
	})

	e.vm.Set("java", java)

	// Also set common global variables
	e.vm.Set("baseUrl", e.baseURL)
	e.vm.Set("result", e.result)

	// Cookie helper object
	cookie := e.vm.NewObject()
	cookie.Set("getKey", func(call goja.FunctionCall) goja.Value {
		// Placeholder for cookie management
		return e.vm.ToValue("")
	})
	e.vm.Set("cookie", cookie)

	// Source object placeholder
	source := e.vm.NewObject()
	source.Set("getKey", func(call goja.FunctionCall) goja.Value {
		return e.vm.ToValue("")
	})
	source.Set("getLoginInfoMap", func(call goja.FunctionCall) goja.Value {
		return e.vm.NewObject()
	})
	e.vm.Set("source", source)
}

// isPrivateIP checks if an IP address is in a private/reserved range
func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	// Check for loopback
	if ip.IsLoopback() {
		return true
	}
	// Check for private ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16", // link-local
		"127.0.0.0/8",    // loopback
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 unique local
		"fe80::/10",      // IPv6 link-local
	}
	for _, cidr := range privateRanges {
		_, block, err := net.ParseCIDR(cidr)
		if err == nil && block.Contains(ip) {
			return true
		}
	}
	return false
}

// validateURL validates a URL to prevent SSRF attacks
func validateURL(urlStr string) error {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Only allow http and https protocols
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("only http and https protocols are allowed")
	}

	// Resolve the hostname to check for private IPs
	host := parsed.Hostname()
	ips, err := net.LookupIP(host)
	if err != nil {
		// Allow if DNS lookup fails (might be valid external host)
		return nil
	}

	for _, ip := range ips {
		if isPrivateIP(ip) {
			return fmt.Errorf("access to private IP addresses is not allowed")
		}
	}

	return nil
}

// fetchURL fetches content from a URL
func (e *JSEngine) fetchURL(urlStr string) (string, error) {
	// Parse URL options (url, {options})
	urlStr = strings.TrimSpace(urlStr)

	var requestURL string
	var method = "GET"
	var body string
	var charset = "UTF-8"

	// Check for comma-separated options
	if idx := strings.Index(urlStr, ",{"); idx != -1 {
		requestURL = strings.TrimSpace(urlStr[:idx])
		// Options are in JSON format after the comma
		// For now, just use the URL
	} else {
		requestURL = urlStr
	}

	// Validate URL to prevent SSRF attacks
	if err := validateURL(requestURL); err != nil {
		return "", err
	}

	// Build headers
	headers := http.Header{
		"User-Agent": []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36"},
	}

	if e.cookie != "" {
		headers.Set("Cookie", e.cookie)
	}

	for k, v := range e.headers {
		headers.Set(k, v)
	}

	var content []byte
	var err error

	if method == "POST" {
		content, err = postBytes(requestURL, headers, []byte(body), e.timeout, 3)
	} else {
		content, err = httputil.GetBytes(requestURL, headers, e.timeout, 3)
	}

	if err != nil {
		return "", err
	}

	// Handle charset conversion if needed
	if strings.ToUpper(charset) != "UTF-8" && strings.ToUpper(charset) != "UTF8" {
		// TODO: implement charset conversion
	}

	return string(content), nil
}

// Eval executes JavaScript code and returns the result
func (e *JSEngine) Eval(script string) (interface{}, error) {
	// Load jsLib first if available
	if e.jsLib != "" {
		_, err := e.vm.RunString(e.jsLib)
		if err != nil {
			// Log but don't fail - jsLib might have app-specific code
		}
	}

	// Execute the script
	val, err := e.vm.RunString(script)
	if err != nil {
		return nil, err
	}

	return val.Export(), nil
}

// EvalString executes JavaScript and returns string result
func (e *JSEngine) EvalString(script string) (string, error) {
	result, err := e.Eval(script)
	if err != nil {
		return "", err
	}

	if result == nil {
		return "", nil
	}

	return fmt.Sprintf("%v", result), nil
}

// ProcessJSRule processes a rule that may contain JavaScript
// Handles @js: prefix and <js>...</js> blocks
func (e *JSEngine) ProcessJSRule(rule string, result string) (string, error) {
	e.SetResult(result)

	// Handle @js: prefix
	if strings.HasPrefix(rule, "@js:") {
		script := strings.TrimPrefix(rule, "@js:")
		return e.EvalString(script)
	}

	// Handle <js>...</js> blocks
	if strings.Contains(rule, "<js>") {
		before, js, after, hasJS := ExtractJSBlock(rule)
		if hasJS {
			// The 'before' part should be processed first to get result
			// Then JS processes that result
			// Then 'after' is appended
			jsResult, err := e.EvalString(js)
			if err != nil {
				return "", err
			}
			return before + jsResult + after, nil
		}
	}

	return rule, nil
}

// ProcessTemplate processes {{...}} templates in a string
func (e *JSEngine) ProcessTemplate(template string, vars map[string]interface{}) (string, error) {
	// Set variables
	for k, v := range vars {
		e.vm.Set(k, v)
	}

	// Replace {{...}} with evaluated results
	var lastErr error

	result := templatePattern.ReplaceAllStringFunc(template, func(match string) string {
		expr := match[2 : len(match)-2] // Remove {{ and }}

		// Check if it's a rule reference (starts with @@ or rule prefix)
		if strings.HasPrefix(expr, "@@") {
			// It's a JSOUP rule reference
			if e.analyzer != nil {
				results := e.analyzer.ParseRule([]byte(e.result), expr[2:])
				if len(results) > 0 {
					return results[0]
				}
			}
			return ""
		}

		// Otherwise evaluate as JavaScript
		val, err := e.EvalString(expr)
		if err != nil {
			lastErr = err
			return ""
		}
		return val
	})

	return result, lastErr
}

// BuildURL builds a URL from template with variables
func (e *JSEngine) BuildURL(template string, key string, page int) (string, error) {
	e.vm.Set("key", key)
	e.vm.Set("page", page)

	// URL encode key for use in URLs
	e.vm.Set("keyEncoded", url.QueryEscape(key))

	return e.ProcessTemplate(template, nil)
}
