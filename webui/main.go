package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dfordsoft/golib/fsutil"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
)

var (
	fontFiles   []string
	mutexBooks  sync.Mutex
	mutexMaking sync.Mutex
	books       []*HistoryItem
)

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
		"items":     books,
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
	kindlegen, _ := exec.LookPath(`kindlegen`)
	if b, e := fsutil.FileExists(kindlegen); e != nil || !b {
		if dir, err := filepath.Abs(filepath.Dir(os.Args[0])); err == nil {
			kindlegen = filepath.Join(dir, `kindlegen`)
		}
	}
	cmd.Env = append(os.Environ(),
		"KINDLEGEN_PATH="+kindlegen, // ignored
	)
	go func() {
		item := &HistoryItem{
			TOCURL:   gnargs.TOCURL,
			BookName: gnargs.TOCURL,
			Status:   "等待制作",
		}
		mutexBooks.Lock()
		books = append(books, item)
		mutexBooks.Unlock()

		mutexMaking.Lock()
		// monitor current directory
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Println(err)
		}

		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Println(err)
		}

		go func() {
			err := watcher.Add(dir)
			if err != nil {
				log.Println(err)
				return
			}
			defer watcher.Close()
			for {
				select {
				case event := <-watcher.Events:
					if event.Op&fsnotify.Write == fsnotify.Write {
						if strings.ToLower(filepath.Ext(event.Name)) == ".pdf" {
							item.BookName = filepath.Base(event.Name)
							item.DownloadLink = "/download/pdf/" + filepath.Base(event.Name)
							item.DeleteLink = "/delete/pdf/" + filepath.Base(event.Name)
							return
						}
						if strings.ToLower(filepath.Ext(event.Name)) == ".mobi" ||
							strings.ToLower(filepath.Ext(event.Name)) == ".epub" {
							item.BookName = filepath.Base(event.Name)
							item.DownloadLink = "/download/" + filepath.Base(path.Dir(event.Name)) + "/" + filepath.Base(event.Name)
							item.DeleteLink = "/delete/" + filepath.Base(path.Dir(event.Name)) + "/" + filepath.Base(event.Name)
							return
						}
						if b, e := fsutil.IsDir(event.Name); e == nil && b {
							err = watcher.Add(event.Name)
							if err != nil {
								log.Println(err)
							}
							watcher.Remove(dir)
						}
					}
				case err := <-watcher.Errors:
					if err != nil {
						log.Println("error:", err)
					}
				}
			}
		}()

		item.Status = "制作中"
		err = cmd.Run()
		mutexMaking.Unlock()

		if err != nil {
			item.Status = "制作失败"
		} else {
			item.Status = "有效"
		}
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
	mutexBooks.Lock()
	for i, book := range books {
		if book.BookName == strings.Replace(filepath.Join(path, name), "\\", "/", -1) {
			books = append(books[:i], books[i+1:]...)
			break
		}
	}
	mutexBooks.Unlock()
	c.Redirect(http.StatusFound, "/")
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

	matches, err = filepath.Glob("*.pdf")
	if err == nil {
		for _, v := range matches {
			books = append(books, &HistoryItem{
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
			books = append(books, &HistoryItem{
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
			books = append(books, &HistoryItem{
				BookName:     filepath.Base(v),
				Status:       "有效",
				DownloadLink: "/download/" + strings.Replace(v, "\\", "/", -1),
				DeleteLink:   "/delete/" + strings.Replace(v, "\\", "/", -1),
			})
		}
	}
	r.Run(addr)
}
