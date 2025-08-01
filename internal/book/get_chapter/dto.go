package get_chapter

type Input struct {
	Path string `param:"path" validate:"required"`
}

type Output []byte
