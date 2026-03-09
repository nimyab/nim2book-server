package get_chapter

type Input struct {
	Path string `query:"path" validate:"required"`
}

type Output []byte
