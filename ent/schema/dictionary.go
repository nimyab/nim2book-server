package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Dictionary holds the schema definition for the Dictionary entity.
type Dictionary struct {
	ent.Schema
}

// Fields of the Dictionary.
func (Dictionary) Fields() []ent.Field {
	return append([]ent.Field{
		field.String("text").SchemaType(varchar255),
		field.String("part_of_speech").SchemaType(varchar255),
		field.String("transcription").SchemaType(varchar255),
		field.String("from_lang_code").SchemaType(varchar10),
		field.String("to_lang_code").SchemaType(varchar10),
		field.Strings("translations"),
	}, defaultFields...)
}

// Edges of the Dictionary.
func (Dictionary) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("dictionary_examples", DictionaryExample.Type),
	}
}

func (Dictionary) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("text", "part_of_speech").Unique(),
	}
}
