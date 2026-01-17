package legado

import (
	"testing"
)

func TestJSEngine_BasicEval(t *testing.T) {
	engine := NewJSEngine("https://example.com")

	tests := []struct {
		script   string
		expected string
	}{
		{"1 + 1", "2"},
		{"'hello' + ' ' + 'world'", "hello world"},
		{"Math.max(1, 2, 3)", "3"},
	}

	for _, tt := range tests {
		t.Run(tt.script, func(t *testing.T) {
			got, err := engine.EvalString(tt.script)
			if err != nil {
				t.Fatalf("EvalString error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("EvalString(%q) = %q, want %q", tt.script, got, tt.expected)
			}
		})
	}
}

func TestJSEngine_Base64(t *testing.T) {
	engine := NewJSEngine("")

	// Test base64Encode
	encoded, err := engine.EvalString("java.base64Encode('hello')")
	if err != nil {
		t.Fatalf("base64Encode error: %v", err)
	}
	if encoded != "aGVsbG8=" {
		t.Errorf("base64Encode = %q, want 'aGVsbG8='", encoded)
	}

	// Test base64Decode
	decoded, err := engine.EvalString("java.base64Decode('aGVsbG8=')")
	if err != nil {
		t.Fatalf("base64Decode error: %v", err)
	}
	if decoded != "hello" {
		t.Errorf("base64Decode = %q, want 'hello'", decoded)
	}
}

func TestJSEngine_MD5(t *testing.T) {
	engine := NewJSEngine("")

	// Test md5Encode
	md5, err := engine.EvalString("java.md5Encode('hello')")
	if err != nil {
		t.Fatalf("md5Encode error: %v", err)
	}
	expected := "5d41402abc4b2a76b9719d911017c592"
	if md5 != expected {
		t.Errorf("md5Encode = %q, want %q", md5, expected)
	}

	// Test md5Encode16
	md5_16, err := engine.EvalString("java.md5Encode16('hello')")
	if err != nil {
		t.Fatalf("md5Encode16 error: %v", err)
	}
	expected16 := "bc4b2a76b9719d91"
	if md5_16 != expected16 {
		t.Errorf("md5Encode16 = %q, want %q", md5_16, expected16)
	}
}

func TestJSEngine_PutGet(t *testing.T) {
	engine := NewJSEngine("")

	// Put a value
	_, err := engine.EvalString("java.put('key1', 'value1')")
	if err != nil {
		t.Fatalf("put error: %v", err)
	}

	// Get the value
	got, err := engine.EvalString("java.get('key1')")
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if got != "value1" {
		t.Errorf("get('key1') = %q, want 'value1'", got)
	}
}

func TestJSEngine_ResultVariable(t *testing.T) {
	engine := NewJSEngine("")
	engine.SetResult("test content")

	got, err := engine.EvalString("result")
	if err != nil {
		t.Fatalf("EvalString error: %v", err)
	}
	if got != "test content" {
		t.Errorf("result = %q, want 'test content'", got)
	}
}

func TestJSEngine_ProcessJSRule(t *testing.T) {
	engine := NewJSEngine("")

	tests := []struct {
		rule     string
		result   string
		expected string
	}{
		{"@js:result.toUpperCase()", "hello", "HELLO"},
		{"@js:result.replace(/\\s+/g, ' ')", "hello   world", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			got, err := engine.ProcessJSRule(tt.rule, tt.result)
			if err != nil {
				t.Fatalf("ProcessJSRule error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("ProcessJSRule(%q, %q) = %q, want %q",
					tt.rule, tt.result, got, tt.expected)
			}
		})
	}
}

func TestJSEngine_ProcessTemplate(t *testing.T) {
	engine := NewJSEngine("")
	engine.SetResult("test")

	vars := map[string]interface{}{
		"page": 2,
		"key":  "search",
	}

	template := "page={{page}}&key={{key}}"
	expected := "page=2&key=search"

	got, err := engine.ProcessTemplate(template, vars)
	if err != nil {
		t.Fatalf("ProcessTemplate error: %v", err)
	}
	if got != expected {
		t.Errorf("ProcessTemplate = %q, want %q", got, expected)
	}
}

func TestJSEngine_BuildURL(t *testing.T) {
	engine := NewJSEngine("")

	tests := []struct {
		template string
		key      string
		page     int
		expected string
	}{
		{"/search?q={{key}}&p={{page}}", "test", 1, "/search?q=test&p=1"},
		{"/list?offset={{(page-1)*20}}", "", 2, "/list?offset=20"},
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			got, err := engine.BuildURL(tt.template, tt.key, tt.page)
			if err != nil {
				t.Fatalf("BuildURL error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("BuildURL(%q, %q, %d) = %q, want %q",
					tt.template, tt.key, tt.page, got, tt.expected)
			}
		})
	}
}

func TestJSEngine_TimeFormat(t *testing.T) {
	engine := NewJSEngine("")

	// Test with a known timestamp (2024-01-15 12:30:00 UTC)
	got, err := engine.EvalString("java.timeFormat(1705321800)")
	if err != nil {
		t.Fatalf("timeFormat error: %v", err)
	}

	// Result depends on timezone, just check format
	if len(got) < 10 {
		t.Errorf("timeFormat result too short: %q", got)
	}
}
