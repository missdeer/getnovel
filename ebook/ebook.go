package ebook

// IBook interface for variant ebook generators
type IBook interface {
	Info()
	Begin()
	End()
	SetAuthor(string)
	SetTitle(string)
	AppendContent(string, string, string)
	SetMargins(float64, float64)
	SetPageType(string)
	SetPageSize(float64, float64)
	SetPDFFontSize(int, int)
	SetHTMLH1Font(string, string)
	SetHTMLH2Font(string, string)
	SetHTMLBodyFont(string, string)
	SetHTMLParaFont(string, string, string)
	SetLineSpacing(float64)
	SetFontFile(string)
	PagesPerFile(int)
	ChaptersPerFile(int)
	Output(string)
}

const (
	creator = `GetNovel，仅限个人研究学习，对其造成的所有后果，软件/库作者不承担任何责任`
)

// NewBook create an instance and return as an interface
func NewBook(bookType string) IBook {
	switch bookType {
	case "pdf":
		return &pdfBook{}
	case "mobi":
		return &kindlegenMobiBook{}
	case "epub":
		return &epubBook{}
	case "html":
		return &singleHTMLBook{}
	default:
		return nil
	}
}
