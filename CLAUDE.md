# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GetNovel is a CLI tool for downloading novels from various Chinese novel websites and converting them to multiple ebook formats (epub, pdf, mobi, html).

## Build Commands

### Building GetNovel
```bash
CGO_ENABLED=0 go build -ldflags="-s -w"
```

### Running Tests
```bash
go test ./...
```

## Architecture

### Core Components

- **main.go**: Entry point, handles CLI arguments via `go-flags`, orchestrates download and ebook generation
- **config/**: Configuration types (`Options` struct) and command-line option definitions
- **handler/**: Website handlers that extract chapter lists and content from novel sites
  - Built-in handlers: `piaotian.go`, `69shuba.go`, `uukanshu.go`, `7mao.go`
- **ebook/**: Ebook generators implementing `IBook` interface
  - `epub.go`, `pdf.go`, `kindlegenmobi.go`, `html.go`
- **ebook/bs/**: Book source support for "阅读" app format (v2/v3 JSON book sources)
- **legado/**: Full Legado book source rule parser implementation
  - `types.go`: Core type definitions (BookSource, rules, search results)
  - `rule.go`: Rule parsing utilities, type detection, combinators
  - `jsoup.go`: JSOUP Default syntax parser (class.name@selector)
  - `css.go`: CSS selector parser (@css: prefix)
  - `jsonpath.go`: JSONPath parser ($. or @json: prefix)
  - `xpath.go`: XPath parser (// or @XPath: prefix)
  - `regex.go`: Regex pattern parser for content extraction/replacement
  - `jsengine.go`: JavaScript engine (goja) with java.* method bindings
  - `analyzer.go`: Unified rule analyzer that auto-detects rule types
  - `executor.go`: Book source executor for search, book info, chapters, content

### Handler Registration Pattern

Handlers are registered in `init()` functions using `registerNovelSiteHandler()`. Each handler implements:
- `CanHandle(url)`: Check if handler supports the URL
- `ExtractChapterList(url, rawContent)`: Parse chapter list from index page
- `ExtractChapterContent(url, rawContent)`: Extract chapter text from chapter page
- Optional: `PreprocessChapterListURL`, `PreprocessContentLink`, `Begin`, `End`

## Key Dependencies

- `github.com/PuerkitoBio/goquery`: HTML parsing
- `github.com/bmaupin/go-epub`: EPUB generation
- `github.com/signintech/gopdf`: PDF generation
- `github.com/jessevdk/go-flags`: CLI argument parsing
- `github.com/dop251/goja`: JavaScript engine for Legado rule execution
- `github.com/tidwall/gjson`: Fast JSON parsing for JSONPath rules
- `github.com/antchfx/htmlquery`: XPath queries on HTML documents

## Environment Variables

- `HTTP_PROXY`/`HTTPS_PROXY`: HTTP proxy with scheme (e.g., `http://127.0.0.1:7890`)
- `SOCKS5_PROXY`: SOCKS5 proxy without scheme (e.g., `127.0.0.1:7891`)
- `KINDLEGEN_PATH`: Path to kindlegen executable for mobi generation

## Legado Book Source Support

The `legado/` package implements comprehensive support for "阅读" (Legado) app book source rules:

### Supported Rule Types
- **JSOUP Default**: `class.name.index@selector@content` (e.g., `class.author@text`)
- **CSS Selector**: `@css:.selector@content` (e.g., `@css:.book-title@text`)
- **JSONPath**: `$.path` or `@json:$.path` (e.g., `$.data.books.#.name`)
- **XPath**: `//path` or `@XPath://path` (e.g., `//div[@class='title']`)
- **Regex**: `:pattern` for AllInOne matching
- **JavaScript**: `@js:code` or `<js>code</js>` blocks

### Rule Combinators
- `&&`: Merge all results from multiple rules
- `||`: Use first non-empty result (fallback)
- `%%`: Format/template combinator

### JavaScript Engine (java.* methods)
- `java.ajax(url)`: Fetch URL content
- `java.base64Encode/Decode`: Base64 encoding
- `java.md5Encode`: MD5 hashing
- `java.getString(rule)`: Parse rule from result
- `java.put/get`: Variable storage
- `java.timeFormat`: Timestamp formatting

### Not Supported
- WebView-dependent rules (requires browser environment)
- Some advanced app-specific features
