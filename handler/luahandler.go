package handler

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aarzilli/golua/lua"
	"github.com/missdeer/getnovel/config"
	lw "github.com/missdeer/getnovel/luawrapper"
	"github.com/missdeer/golib/httputil"
	"gitlab.com/ambrevar/golua/unicode"
)

type extractExternalChapterListRequest struct {
	url            string
	rawPageContent []byte
}

type extractExternalChapterListResponse struct {
	title    string
	chapters []*config.NovelChapterInfo
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

var (
	md5sumMap = make(map[string]string)
)

func readMD5SumMap() {
	md5sum, err := httputil.GetBytes("https://cdn.jsdelivr.net/gh/missdeer/getnovel@master/handlers/md5sum.txt",
		http.Header{"User-Agent": []string{"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0"}},
		60*time.Second,
		3)
	if err != nil {
		log.Println(err)
	} else {
		// 按行解析md5sum, 格式为 md5sum 文件名
		lines := strings.Split(string(md5sum), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			parts := strings.Split(line, " ")
			if len(parts) != 2 {
				continue
			}

			md5sumMap[parts[1]] = parts[0]
		}
	}
}

// 本地lua文件与md5sumMap中的记录进行md5校验
// 如果md5值不同，则从 https://cdn.jsdelivr.net/gh/missdeer/getnovel@master/handlers/ 下载新的lua文件
func checkLuaFile(localDirPath string, fileName string) error {
	f, err := os.Open(filepath.Join(localDirPath, fileName))
	if err != nil {
		return err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	sum := h.Sum(nil)
	localMd5sum := hex.EncodeToString(sum)
	if strings.ToLower(localMd5sum) != strings.ToLower(md5sumMap[fileName]) {
		content, err := httputil.GetBytes("https://cdn.jsdelivr.net/gh/missdeer/getnovel@master/handlers/"+fileName,
			http.Header{"User-Agent": []string{"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0"}},
			60*time.Second,
			3)
		if err != nil {
			return err
		}
		// save to local file
		f, err := os.OpenFile(filepath.Join(localDirPath, fileName), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := f.Write(content); err != nil {
			return err
		}
	}

	return nil
}

func (h *ExternalHandler) initLuaEnv() {
	h.l.OpenLibs()

	lw.RegisterLuaAPIs(h.l)

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
	if err != nil {
		log.Println(err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.ToLower(filepath.Ext(file.Name())) != ".lua" {
			continue
		}

		if config.Opts.AutoUpdateExternalHandlers {
			if err := checkLuaFile(directory, file.Name()); err != nil {
				log.Println(err)
			}
		}

		h.l.DoFile(filepath.Join(directory, file.Name()))
	}
}

func (h *ExternalHandler) destroyLuaEnv() {
	h.l.Close()
}

func (h *ExternalHandler) invokePreprocessExternalChapterListURL(u string) string {
	h.l.GetGlobal("PreprocessChapterListURL")
	h.l.PushString(u)
	h.l.Call(1, 1)
	defer h.l.Pop(1)
	if !h.l.IsString(-1) {
		return u
	}
	return h.l.ToString(-1)
}

func (h *ExternalHandler) invokeExtractExternalChapterList(u string, rawPageContent []byte) (title string, chapters []*config.NovelChapterInfo) {
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
			chapter := &config.NovelChapterInfo{}

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
	h.l.Pop(2)
	return
}

func (h *ExternalHandler) invokeExtractExternalChapterContent(rawPageContent []byte) (c []byte) {
	h.l.GetGlobal("ExtractChapterContent")
	h.l.PushBytes(rawPageContent)
	h.l.Call(1, 1)
	defer h.l.Pop(1)
	if !h.l.IsString(-1) {
		return
	}
	return []byte(h.l.ToString(-1))
}

func (h *ExternalHandler) invokeCanHandleExternalSite(u string) bool {
	h.l.GetGlobal("CanHandle")
	h.l.PushString(u)
	h.l.Call(1, 1)
	defer h.l.Pop(1)
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
	if len(md5sumMap) == 0 {
		readMD5SumMap()
	}
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

func (h *ExternalHandler) extractExternalChapterList(u string, rawPageContent []byte) (title string, chapters []*config.NovelChapterInfo) {
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

	registerNovelSiteHandler(&config.NovelSiteHandler{
		Sites: []config.NovelSite{
			{
				Title: `外部脚本处理器`,
				Urls:  []string{},
			},
		},
		CanHandle:                handler.canHandleExternalSite,
		PreprocessChapterListURL: handler.preprocessExternalChapterListURL,
		ExtractChapterList:       handler.extractExternalChapterList,
		ExtractChapterContent:    handler.extractExternalChapterContent,
		Begin:                    handler.begin,
		End:                      handler.end,
	})
}
