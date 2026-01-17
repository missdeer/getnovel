package config

import (
	"path/filepath"
	"runtime"
)

// Options for all command line options, long name must match field name
type Options struct {
	InsecureSkipVerify         bool    `short:"V" long:"insecureSkipVerify" description:"if true, TLS accepts any certificate"`
	ListenAndServe             string  `short:"s" long:"listenAndServe" description:"set http listen and serve address, example: :8080"`
	Format                     string  `short:"f" long:"format" description:"set generated file format, candidate values: mobi, epub, pdf, html, txt"`
	List                       bool    `short:"l" long:"list" description:"list supported novel websites"`
	LeftMargin                 float64 `long:"leftMargin" description:"set left margin for PDF format"`
	TopMargin                  float64 `long:"topMargin" description:"set top margin for PDF format"`
	PageWidth                  float64 `long:"pageWidth" description:"set page width for PDF format(unit: mm)"`
	PageHeight                 float64 `long:"pageHeight" description:"set page height for PDF format(unit: mm)"`
	PageType                   string  `short:"p" long:"pageType" description:"set page type for PDF format, add suffix to output file name"`
	TitleFontSize              int     `long:"titleFontSize" description:"set title font point size for PDF format"`
	ContentFontSize            int     `long:"contentFontSize" description:"set content font point size for PDF format"`
	LineSpacing                float64 `long:"lineSpacing" description:"set line spacing rate for PDF format"`
	PagesPerFile               int     `long:"pagesPerFile" description:"split the big single PDF file to several smaller PDF files, how many pages should be included in a file, 0 means don't split"`
	ChaptersPerFile            int     `long:"chaptersPerFile" description:"split the big single PDF file to several smaller PDF files, how many chapters should be included in a file, 0 means don't split"`
	FontFile                   string  `long:"fontFile" description:"set TTF font file path"`
	H1FontFamily               string  `long:"h1FontFamily" description:"set H1 font family for mobi/epub/html format"`
	H1FontSize                 string  `long:"h1FontSize" description:"set H1 font size for mobi/epub/html format"`
	H2FontFamily               string  `long:"h2FontFamily" description:"set H2 font family for mobi/epub/html format"`
	H2FontSize                 string  `long:"h2FontSize" description:"set H2 font size for mobi/epub/html format"`
	BodyFontFamily             string  `long:"bodyFontFamily" description:"set body font family for mobi/epub/html format"`
	BodyFontSize               string  `long:"bodyFontSize" description:"set body font size for mobi/epub/html format"`
	ParaFontFamily             string  `long:"paraFontFamily" description:"set paragraph font family for mobi/epub/html format"`
	ParaFontSize               string  `long:"paraFontSize" description:"set paragraph font size for mobi/epub/html format"`
	ParaLineHeight             string  `long:"paraLineHeight" description:"set paragraph line height for mobi/epub/html format"`
	RetryCount                 int     `short:"r" long:"retryCount" description:"download retry count"`
	Timeout                    int     `short:"t" long:"timeout" description:"download timeout seconds"`
	ParallelCount              int64   `long:"parallelCount" description:"parallel count for downloading"`
	ConfigFile                 string  `short:"c" long:"configFile" description:"read configurations from local file"`
	OutputFile                 string  `short:"o" long:"outputFile" description:"output file path"`
	FromChapter                int     `long:"fromChapter" description:"from chapter"`
	FromTitle                  string  `long:"fromTitle" description:"from title"`
	ToChapter                  int     `long:"toChapter" description:"to chapter"`
	ToTitle                    string  `long:"toTitle" description:"to title"`
	Author                     string  `short:"a" long:"author" description:"author"`
	WaitInterval               int     `long:"waitInterval" description:"wait interval seconds between each download"`
}

var (
	Opts Options
)

func init() {
	Opts = Options{
		InsecureSkipVerify:         false,
		Format:                     "epub",
		List:                       false,
		LeftMargin:                 10,
		TopMargin:                  10,
		PageHeight:                 841.89,
		PageWidth:                  595.28,
		TitleFontSize:              24,
		ContentFontSize:            18,
		H1FontFamily:               "CustomFont",
		H2FontFamily:               "CustomFont",
		BodyFontFamily:             "CustomFont",
		ParaFontFamily:             "CustomFont",
		H1FontSize:                 "4em",
		H2FontSize:                 "1.2em",
		BodyFontSize:               "1.2em",
		ParaFontSize:               "1.0em",
		ParaLineHeight:             "1.0em",
		LineSpacing:                1.2,
		PagesPerFile:               0,
		ChaptersPerFile:            0,
		FontFile:                   filepath.Join("fonts", "CustomFont.ttf"),
		RetryCount:                 3,
		Timeout:                    60,
		ParallelCount:              int64(runtime.NumCPU()) * 2, // get cpu logical core number
		Author:                     "GetNovel用户",
		WaitInterval:               0,
	}
}
