package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/aarzilli/golua/lua"
	"github.com/missdeer/golib/ic"
	"gitlab.com/ambrevar/golua/unicode"
)

func ConvertEncoding(L *lua.State) int {
	fromEncoding := L.CheckString(1)
	toEncoding := L.CheckString(2)
	fromStr := L.CheckString(3)
	toStr := ic.ConvertString(fromEncoding, toEncoding, fromStr)
	L.PushString(toStr)
	return 1
}

type extractExternalChapterListRequest struct {
	url            string
	rawPageContent []byte
}

type extractExternalChapterListResponse struct {
	title    string
	chapters []*NovelChapterInfo
}

type ExternalHandler struct {
	l                                            *lua.State
	quit                                         chan bool
	preprocessExternalChapterListURLRequestParam chan string
	preprocessExternalChapterListURLResponse     chan string
	extractExternalChapterListRequestParam       chan extractExternalChapterListRequest
	extractExternalChapterListResponse           chan extractExternalChapterListResponse
	extractExternalChapterContentRequestParam    chan []byte
	extractExternalChapterContentResponse        chan []byte
	canHandleExternalSiteRequestParam            chan string
	canHandleExternalSiteResponse                chan bool
}

func newExternalHandler() *ExternalHandler {
	return &ExternalHandler{
		l:    lua.NewState(),
		quit: make(chan bool),
		preprocessExternalChapterListURLRequestParam: make(chan string),
		preprocessExternalChapterListURLResponse:     make(chan string),
		extractExternalChapterListRequestParam:       make(chan extractExternalChapterListRequest),
		extractExternalChapterListResponse:           make(chan extractExternalChapterListResponse),
		extractExternalChapterContentRequestParam:    make(chan []byte),
		extractExternalChapterContentResponse:        make(chan []byte),
		canHandleExternalSiteRequestParam:            make(chan string),
		canHandleExternalSiteResponse:                make(chan bool),
	}
}

func (h *ExternalHandler) initLuaEnv() {
	h.l.OpenLibs()

	// add string.convert(from, to, str) method
	h.l.GetGlobal("string")
	h.l.PushGoFunction(ConvertEncoding)
	h.l.SetField(-2, "convert")
	h.l.Pop(1)

	unicode.GoLuaReplaceFuncs(h.l)

	// get current executable path
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	exePath = filepath.Dir(exePath)

	h.l.DoFile(exePath + `/lua/init.lua`)

	// traverse ./handler directory, find all .lua files and load them
	directory := exePath + "/handlers"
	files, err := os.ReadDir(directory)
	if err == nil {
		for _, file := range files {
			if filepath.Ext(file.Name()) == ".lua" {
				h.l.DoFile(filepath.Join(directory, file.Name()))
			}
		}
	}
}

func (h *ExternalHandler) destroyLuaEnv() {
	h.l.Close()
}

func (h *ExternalHandler) invokePreprocessExternalChapterListURL(u string) string {
	h.l.GetGlobal("PreprocessChapterListURL")
	h.l.PushString(u)
	h.l.Call(1, 1)
	if !h.l.IsString(-1) {
		return u
	}
	return h.l.ToString(-1)
}

func (h *ExternalHandler) invokeExtractExternalChapterList(u string, rawPageContent []byte) (title string, chapters []*NovelChapterInfo) {
	h.l.GetGlobal("ExtractChapterList")
	h.l.PushString(u)
	h.l.PushBytes(rawPageContent)
	h.l.Call(2, 2)
	if !h.l.IsString(-2) || !h.l.IsTable(-1) {
		return "", []*NovelChapterInfo{}
	}

	title = h.l.ToString(-2)
	h.l.PushNil()
	for h.l.Next(-2) != 0 {
		// 现在栈的情况：-1 => 值（表），-2 => 键，-3 => 章节表
		if h.l.IsTable(-1) {
			chapter := &NovelChapterInfo{}

			h.l.PushString("Index")
			h.l.GetTable(-2)
			if h.l.IsNumber(-1) {
				chapter.Index = int(h.l.ToInteger(-1))
			}
			h.l.Pop(1)

			h.l.PushString("Title")
			h.l.GetTable(-2)
			if h.l.IsString(-1) {
				chapter.Title = h.l.ToString(-1)
			}
			h.l.Pop(1)

			h.l.PushString("URL")
			h.l.GetTable(-2)
			if h.l.IsString(-1) {
				chapter.URL = h.l.ToString(-1)
			}
			h.l.Pop(1)

			chapters = append(chapters, chapter)
		}
		h.l.Pop(1)

	}

	return
}

func (h *ExternalHandler) invokeExtractExternalChapterContent(rawPageContent []byte) (c []byte) {
	h.l.GetGlobal("ExtractChapterContent")
	h.l.PushBytes(rawPageContent)
	h.l.Call(1, 1)
	if !h.l.IsString(-1) {
		return []byte{}
	}
	return []byte(h.l.ToString(-1))
}

func (h *ExternalHandler) invokeCanHandleExternalSite(u string) bool {
	h.l.GetGlobal("CanHandle")
	h.l.PushString(u)
	h.l.Call(1, 1)
	if !h.l.IsBoolean(-1) {
		return false
	}
	return h.l.ToBoolean(-1)
}

func (h *ExternalHandler) invokeMethodLoop() {
	for {
		select {
		case <-h.quit:
			return

		case u := <-h.preprocessExternalChapterListURLRequestParam:
			h.preprocessExternalChapterListURLResponse <- h.invokePreprocessExternalChapterListURL(u)
			break

		case req := <-h.extractExternalChapterListRequestParam:
			title, chapters := h.invokeExtractExternalChapterList(req.url, req.rawPageContent)
			h.extractExternalChapterListResponse <- extractExternalChapterListResponse{title: title, chapters: chapters}
			break

		case rawPageContent := <-h.extractExternalChapterContentRequestParam:
			h.extractExternalChapterContentResponse <- h.invokeExtractExternalChapterContent(rawPageContent)
			break

		case u := <-h.canHandleExternalSiteRequestParam:
			h.canHandleExternalSiteResponse <- h.invokeCanHandleExternalSite(u)
			break
		}
	}
}

func (h *ExternalHandler) begin() {
	go func() {
		h.initLuaEnv()
		h.invokeMethodLoop()
		h.destroyLuaEnv()
	}()
}

func (h *ExternalHandler) end() {
	h.quit <- true
}

func (h *ExternalHandler) preprocessExternalChapterListURL(u string) string {
	h.preprocessExternalChapterListURLRequestParam <- u
	return <-h.preprocessExternalChapterListURLResponse
}

func (h *ExternalHandler) extractExternalChapterList(u string, rawPageContent []byte) (title string, chapters []*NovelChapterInfo) {
	h.extractExternalChapterListRequestParam <- extractExternalChapterListRequest{
		url:            u,
		rawPageContent: rawPageContent,
	}
	resp := <-h.extractExternalChapterListResponse
	return resp.title, resp.chapters
}

func (h *ExternalHandler) extractExternalChapterContent(rawPageContent []byte) (c []byte) {
	h.extractExternalChapterContentRequestParam <- rawPageContent
	return <-h.extractExternalChapterContentResponse
}

func (h *ExternalHandler) canHandleExternalSite(u string) bool {
	h.canHandleExternalSiteRequestParam <- u
	return <-h.canHandleExternalSiteResponse
}

func init() {
	handler := newExternalHandler()

	registerNovelSiteHandler(&NovelSiteHandler{
		Title:                    `外部脚本处理器`,
		Urls:                     []string{},
		CanHandle:                handler.canHandleExternalSite,
		PreprocessChapterListURL: handler.preprocessExternalChapterListURL,
		ExtractChapterList:       handler.extractExternalChapterList,
		ExtractChapterContent:    handler.extractExternalChapterContent,
		Begin:                    handler.begin,
		End:                      handler.end,
	})
}
