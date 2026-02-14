package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

var defaultFields = []ent.Field{
	field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
	field.Time("created_at").Default(time.Now).Immutable(),
}

var varchar255 = map[string]string{
	dialect.Postgres: "VARCHAR(255)",
}

var varchar10 = map[string]string{
	dialect.Postgres: "VARCHAR(10)",
}

var varchar50 = map[string]string{
	dialect.Postgres: "VARCHAR(50)",
}
