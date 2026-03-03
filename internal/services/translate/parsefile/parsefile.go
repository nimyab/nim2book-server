package parsefile

import (
	"errors"
	"io"
	"log/slog"
	"mime/multipart"

	"github.com/nimyab/nim2book-back/pkg/parsers/epub_parser"
)

var (
	ErrOpenFile  = errors.New("file open error")
	ErrReadFile  = errors.New("file read error")
	ErrParseFile = errors.New("file parse error")
)

func ParseFile(file *multipart.FileHeader) (*epub_parser.ParsedData, error) {
	const operation = "translate.ChapterTranslator.ParseFile"
	logger := slog.With(slog.String("operation", operation), slog.String("fileName", file.Filename))

	openedFile, err := file.Open()
	if err != nil {
		logger.Error("failed to open file", slog.String("err", err.Error()))
		return nil, ErrOpenFile
	}
	defer func() {
		if err := openedFile.Close(); err != nil {
			logger.Error("failed to close file", slog.String("err", err.Error()))
		}
	}()

	dataFromFile, err := io.ReadAll(openedFile)
	if err != nil {
		logger.Error("failed to read file", slog.String("err", err.Error()))
		return nil, ErrReadFile
	}

	parsedData, err := epub_parser.Parse(dataFromFile)
	if err != nil {
		logger.Error("failed to parse epub file", slog.String("err", err.Error()))
		return nil, ErrParseFile
	}

	logger.Info(
		"file parsed successfully",
		slog.Int("chaptersCount", len(parsedData.FormattedChapter)),
		slog.String("Author", parsedData.Book.Author),
		slog.String("Title", parsedData.Book.Title),
	)

	return parsedData, nil
}
