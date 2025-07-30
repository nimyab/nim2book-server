package align

type Input struct {
	SourceText string `json:"st" validate:"required"`
	TargetText string `json:"tt" validate:"required"`
}

type Alignments struct {
	SourceWord    string `json:"sw"`
	TargetWord    string `json:"tw"`
	SourceIndexes [2]int `json:"si"`
	TargetIndexes [2]int `json:"ti"`
}

type Output struct {
	SourceText string       `json:"st"`
	TargetText string       `json:"tt"`
	Alignments []Alignments `json:"a"`
}
