package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/dfordsoft/golib/fsutil"
	"github.com/gin-gonic/gin"
)

var (
	fontFiles   []string
	mutexMaking sync.Mutex
	books       = Books{}
	sha1ver     string // sha1 revision used to build the program
	buildTime   string // when the executable was built
)

// Books book collect
type Books struct {
	sync.Mutex
	items []*HistoryItem
}

func (books *Books) append(item *HistoryItem) {
	books.Lock()
	books.items = append(books.items, item)
	books.Unlock()
}

func (books *Books) clear() {
	books.Lock()
	books.items = []*HistoryItem{}
	books.Unlock()
}

func (books *Books) delete(name string) {
	books.Lock()
	for i, book := range books.items {
		if book.BookName == name {
			books.items = append(books.items[:i], books.items[i+1:]...)
			break
		}
	}
	books.Unlock()
}

// HistoryItem - 书籍记录，有4种状态，分别是有效，失败，进行，等待
type HistoryItem struct {
	TOCURL       string `template:"tocurl"`
	BookName     string `template:"name"`
	Status       string `template:"status"`
	DownloadLink string `template:"downloadLink"`
	DeleteLink   string `template:"deleteLink"`
}

func homePage(c *gin.Context) {
	c.HTML(http.StatusOK, "home.tmpl", gin.H{
		"title":     "GetNovel",
		"fontFiles": fontFiles,
		"items":     books.items,
	})
}

type GetNovelAuguments struct {
	TOCURL          string  `form:"tocurl"`
	Format          string  `form:"format"`
	PageType        string  `form:"pageType"`
	LeftMargin      int     `form:"leftMargin"`
	TopMargin       int     `form:"topMargin"`
	TitleFontSize   int     `form:"titleFontSize"`
	ContentFontSize int     `form:"contentFontSize"`
	LineSpacing     float32 `form:"lineSpacing"`
	PagesPerFile    int     `form:"pagesPerFile"`
	ChaptersPerFile int     `form:"chaptersPerFile"`
	FontFile        string  `form:"fontFile"`
	FromTitle       string  `form:"fromTitle"`
	ToTitle         string  `form:"toTitle"`
	FromChapter     int     `form:"fromChapter"`
	ToChapter       int     `form:"toChapter"`
}

func makeEbook(c *gin.Context) {
	var gnargs GetNovelAuguments
	if err := c.ShouldBindJSON(&gnargs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("%v", err),
		})
		return
	}
	getnovel, _ := exec.LookPath(`getnovel`)

	if b, e := fsutil.FileExists(getnovel); e != nil || !b {
		if dir, err := filepath.Abs(filepath.Dir(os.Args[0])); err == nil {
			getnovel = filepath.Join(dir, `getnovel`)
		}
	}

	if b, e := fsutil.FileExists(getnovel); e != nil || !b {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "getnovel not found",
		})
		return
	}

	args := []string{
		"-f",
		gnargs.Format,
		"-p",
		gnargs.PageType,
		fmt.Sprintf("--fontFile=%s", gnargs.FontFile),
	}

	if gnargs.Format == "pdf" {
		args = append(args,
			[]string{
				fmt.Sprintf("--leftMargin=%d", gnargs.LeftMargin),
				fmt.Sprintf("--topMargin=%d", gnargs.TopMargin),
				fmt.Sprintf("--titleFontSize=%d", gnargs.TitleFontSize),
				fmt.Sprintf("--contentFontSize=%d", gnargs.ContentFontSize),
				fmt.Sprintf("--lineSpacing=%f", gnargs.LineSpacing),
				fmt.Sprintf("--pagesPerFile=%d", gnargs.PagesPerFile),
				fmt.Sprintf("--chaptersPerFile=%d", gnargs.ChaptersPerFile),
			}...)
	}
	if gnargs.FromTitle != "" {
		args = append(args, fmt.Sprintf("--fromTitle=%s", gnargs.FromTitle))
	}
	if gnargs.ToTitle != "" {
		args = append(args, fmt.Sprintf("--toTitle=%s", gnargs.ToTitle))
	}
	if gnargs.FromChapter != 0 {
		args = append(args, fmt.Sprintf("--fromChapter=%d", gnargs.FromChapter))
	}
	if gnargs.ToChapter != 0 {
		args = append(args, fmt.Sprintf("--toChapter=%d", gnargs.ToChapter))
	}
	args = append(args, gnargs.TOCURL)

	cmd := exec.Command(getnovel, args...)

	kindlegenName := `kindlegen`
	if runtime.GOOS == "windows" {
		kindlegenName = `kindlegen.exe`
	}
	kindlegenPath, _ := exec.LookPath(kindlegenName)
	if b, e := fsutil.FileExists(kindlegenPath); e != nil || !b {
		if dir, err := filepath.Abs(filepath.Dir(os.Args[0])); err == nil {
			kindlegenPath = filepath.Join(dir, kindlegenName)
		}
	}
	if !filepath.IsAbs(kindlegenPath) {
		kindlegenPath, _ = filepath.Abs(kindlegenPath)
	}
	cmd.Env = append(os.Environ(),
		"KINDLEGEN_PATH="+kindlegenPath, // ignored
	)
	go func() {
		item := &HistoryItem{
			TOCURL:   gnargs.TOCURL,
			BookName: gnargs.TOCURL,
			Status:   "等待制作",
		}
		books.append(item)

		mutexMaking.Lock()
		item.Status = "制作中"
		if err := cmd.Run(); err != nil {
			item.Status = "制作失败"
		} else {
			item.Status = "有效"
		}
		mutexMaking.Unlock()

		books.clear()
		scanEbooks()
	}()

	c.JSON(http.StatusOK, gin.H{})
}

func downloadBook(c *gin.Context) {
	path := c.Param("path")
	if path == "pdf" {
		path = ""
	}
	name := c.Param("name")
	c.File(filepath.Join(path, name))
}

func deleteBook(c *gin.Context) {
	path := c.Param("path")
	if path == "pdf" {
		path = ""
	}
	name := c.Param("name")
	os.Remove(filepath.Join(path, name))
	books.delete(name)
	c.Redirect(http.StatusFound, "/")
}

func scanEbooks() {
	matches, err := filepath.Glob("*.pdf")
	if err == nil {
		for _, v := range matches {
			books.append(&HistoryItem{
				BookName:     v,
				Status:       "有效",
				DownloadLink: "/download/pdf/" + v,
				DeleteLink:   "/delete/pdf/" + v,
			})
		}
	}

	matches, err = filepath.Glob("**/*.mobi")
	if err == nil {
		for _, v := range matches {
			books.append(&HistoryItem{
				BookName:     filepath.Base(v),
				Status:       "有效",
				DownloadLink: "/download/" + strings.Replace(v, "\\", "/", -1),
				DeleteLink:   "/delete/" + strings.Replace(v, "\\", "/", -1),
			})
		}
	}
	matches, err = filepath.Glob("**/*.epub")
	if err == nil {
		for _, v := range matches {
			books.append(&HistoryItem{
				BookName:     filepath.Base(v),
				Status:       "有效",
				DownloadLink: "/download/" + strings.Replace(v, "\\", "/", -1),
				DeleteLink:   "/delete/" + strings.Replace(v, "\\", "/", -1),
			})
		}
	}
}

func main() {
	addr := ":8089"
	if bind, ok := os.LookupEnv("BIND"); ok {
		addr = bind
	}
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/", homePage)
	r.GET("/download/:path/:name", downloadBook)
	r.GET("/delete/:path/:name", deleteBook)
	r.POST("/makeebook", makeEbook)
	// glob font files
	matches, err := filepath.Glob("fonts/*.ttf")
	if err != nil {
		panic(err)
	}
	for _, v := range matches {
		fontFiles = append(fontFiles, filepath.Base(v))
	}
	scanEbooks()
	r.Run(addr)
}
