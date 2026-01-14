package models

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ID is a custom type for UUID primary keys
type ID = uuid.UUID

// StringArray is a wrapper for pq.StringArray for PostgreSQL string arrays
type StringArray []string

// Value implements driver.Valuer interface for database writes
func (a StringArray) Value() (driver.Value, error) {
	return pq.StringArray(a).Value()
}

// Scan implements sql.Scanner interface for database reads
func (a *StringArray) Scan(src interface{}) error {
	return (*pq.StringArray)(a).Scan(src)
}

// BaseModel contains common fields for all models with UUID primary key
type BaseModel struct {
	ID ID `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
}

// JSONB custom type for PostgreSQL JSONB
type JSONB map[string]interface{}

// Value implements driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return nil
	}

	return json.Unmarshal(data, j)
}

// GormDataType returns the data type for GORM
func (JSONB) GormDataType() string {
	return "jsonb"
}
