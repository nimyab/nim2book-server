package translate_book

import (
	"github.com/nimyab/nim2book-back/internal/models"
	"github.com/nimyab/nim2book-back/pkg/parsers/epub_parser"
	"github.com/timsims/pamphlet"
)

type translateStruct struct {
	UserId    models.ID
	Chapters  []epub_parser.FormattedChapter
	CoverData []byte
	Book      *pamphlet.Book
	From      models.SupportedLang
	To        models.SupportedLang
}
