package translate_book

import (
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/pkg/parsers/epub_parser"
)

type translateStruct struct {
	UserId     domain.Id
	ParsedBook *epub_parser.Book
	From       domain.SupportedLang
	To         domain.SupportedLang
}
