package models

// Dictionary represents word translations/definitions
type Dictionary struct {
	BaseModel

	Text    string `gorm:"column:text;type:text;not null;index:idx_text_lang" json:"text"`
	Lang    string `gorm:"column:lang;type:varchar(10);not null;index:idx_text_lang,idx_lang" json:"lang"`
	Content []byte `gorm:"column:content;type:jsonb;not null" json:"content"`
}

func (Dictionary) TableName() string {
	return "dictionary"
}
