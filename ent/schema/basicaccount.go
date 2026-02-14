package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// BasicAccount holds the schema definition for the BasicAccount entity.
type BasicAccount struct {
	ent.Schema
}

// Fields of the BasicAccount.
func (BasicAccount) Fields() []ent.Field {
	return append([]ent.Field{
		field.String("email").Unique().SchemaType(varchar255),
		field.String("password_hash").Sensitive().SchemaType(varchar255),
		field.Bool("is_verified").Default(false),
		field.String("verify_link").Unique().SchemaType(varchar255),
	}, defaultFields...)
}

// Edges of the BasicAccount.
func (BasicAccount) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("basic_account").Unique(),
	}
}
