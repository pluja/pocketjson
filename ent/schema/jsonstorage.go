package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// JsonStorage holds the schema definition for the JsonStorage entity.
type JsonStorage struct {
	ent.Schema
}

// Fields of the JsonStorage.
func (JsonStorage) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.Text("data"),
		field.Time("expires_at"),
		field.String("creator_key").
			Default("guest"),
	}
}
