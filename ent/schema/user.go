package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return append([]ent.Field{
		field.Bool("is_vip").Default(false),
		field.Bool("is_admin").Default(false),
		field.JSON("metadata", map[string]any{}).Default(map[string]any{}),
	}, defaultFields...)
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("google_account", GoogleAccount.Type).Unique(),
		edge.To("basic_account", BasicAccount.Type).Unique(),
		edge.To("personal_books", PersonalBook.Type),
		edge.To("fcm_tokens", FcmToken.Type),
	}
}
