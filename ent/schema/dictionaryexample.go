package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// DictionaryExample holds the schema definition for the DictionaryExample entity.
type DictionaryExample struct {
	ent.Schema
}

// Fields of the DictionaryExample.
func (DictionaryExample) Fields() []ent.Field {
	return append([]ent.Field{
		field.String("text"),
		field.String("translation"),
		field.Int("target_position_start").Nillable().NonNegative(),
		field.Int("target_position_end").Nillable().NonNegative(),
	}, defaultFields...)
}

// Edges of the DictionaryExample.
func (DictionaryExample) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("dictionary", Dictionary.Type).Ref("dictionary_examples").Unique(),
	}
}
