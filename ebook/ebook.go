package ebook

// IBook interface for variant ebook generators
type IBook interface {
	Info()
	Begin()
	End()
	SetTitle(string)
	AppendContent(string, string, string)
	SetMargins(float64, float64)
	SetPageType(string)
	SetPageSize(float64, float64)
	SetFontSize(int, int)
	SetLineSpacing(float64)
	SetFontFile(string)
	PagesPerFile(int)
	ChaptersPerFile(int)
	Output(string)
}

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
