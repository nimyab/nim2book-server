package file_public

type Input struct {
	Path string `query:"path" validate:"required"`
}

type Output []byte
