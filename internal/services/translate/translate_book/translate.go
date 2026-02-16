package translate_book

import (
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/pkg/parsers/epub_parser"
	"github.com/timsims/pamphlet"
)

type translateStruct struct {
	UserId    domain.ID
	Chapters  []epub_parser.FormattedChapter
	CoverData []byte
	Book      *pamphlet.Book
	From      domain.SupportedLang
	To        domain.SupportedLang
}
