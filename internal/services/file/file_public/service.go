package file_public

import (
	"errors"
	"log/slog"
)

type S3 interface {
	Get(path string) ([]byte, error)
}

type Service struct {
	s3 S3
}

func New(s3 S3) *Service {
	return &Service{s3: s3}
}

func (s *Service) GetFile(input *Input) (Output, error) {
	const operation = "file.file_public.GetFile"

	chapterData, err := s.s3.Get(input.Path)
	if err != nil {
		slog.Error(err.Error(), slog.String("path", input.Path), slog.String("operation", operation))
		return nil, errors.New("failed to get file")
	}

	return chapterData, nil
}
