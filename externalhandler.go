package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/aarzilli/golua/lua"
	"github.com/missdeer/golib/ic"
	"gitlab.com/ambrevar/golua/unicode"
)

type ExternalHandler struct {
	l *lua.State
}

func ConvertEncoding(L *lua.State) int {
	fromEncoding := L.CheckString(1)
	toEncoding := L.CheckString(2)
	fromStr := L.CheckString(3)
	toStr := ic.ConvertString(fromEncoding, toEncoding, fromStr)
	L.PushString(toStr)
	return 1
}

func newExternalHandler() *ExternalHandler {
	h := &ExternalHandler{}
	h.l = lua.NewState()
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

	return h
}

func (h *ExternalHandler) preprocessExternalChapterListURL(u string) string {
	h.l.GetGlobal("PreprocessChapterListURL")
	h.l.PushString(u)
	h.l.Call(1, 1)
	if !h.l.IsString(-1) {
		return u
	}
	return h.l.ToString(-1)
}

func (h *ExternalHandler) extractExternalChapterList(u string, rawPageContent []byte) (title string, chapters []*NovelChapterInfo) {
	h.l.GetGlobal("ExtractChapterList")
	h.l.PushString(u)
	h.l.PushBytes(rawPageContent)
	h.l.Call(2, 2)
	if !h.l.IsString(-2) || !h.l.IsTable(-1) {
		return
	}

	title = h.l.ToString(-2)
	h.l.PushNil()
	for h.l.Next(-2) != 0 {
		// 现在栈的情况：-1 => 值（表），-2 => 键，-3 => 章节表
		if h.l.IsTable(-1) {
			chapter := &NovelChapterInfo{}
			h.l.PushString("Index")
			h.l.GetTable(-2)
			if !h.l.IsNumber(-1) {
				continue
			}
			chapter.Index = int(h.l.ToInteger(-1))
			h.l.Pop(1)

			h.l.PushString("Title")
			h.l.GetTable(-2)
			if !h.l.IsString(-1) {
				continue
			}
			chapter.Title = h.l.ToString(-1)
			h.l.Pop(1)

			h.l.PushString("URL")
			h.l.GetTable(-2)
			if !h.l.IsString(-1) {
				continue
			}
			chapter.URL = h.l.ToString(-1)
			h.l.Pop(1)

			chapters = append(chapters, chapter)
		}
		h.l.Pop(1)
	}

	return
}

func (h *ExternalHandler) extractExternalChapterContent(rawPageContent []byte) (c []byte) {
	h.l.GetGlobal("ExtractChapterContent")
	h.l.PushBytes(rawPageContent)
	h.l.Call(1, 1)
	if !h.l.IsString(-1) {
		return
	}
	return []byte(h.l.ToString(-1))
}

func (h *ExternalHandler) canHandleExternalSite(u string) bool {
	h.l.GetGlobal("FindHandler")
	h.l.PushString(u)
	h.l.Call(1, 1)
	if !h.l.IsNumber(-1) {
		return false
	}
	return h.l.ToBoolean(-1)
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
	})
}
