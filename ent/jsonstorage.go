// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"pocketjson/ent/jsonstorage"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
)

// JsonStorage is the model entity for the JsonStorage schema.
type JsonStorage struct {
	config `json:"-"`
	// ID of the ent.
	ID string `json:"id,omitempty"`
	// Data holds the value of the "data" field.
	Data string `json:"data,omitempty"`
	// ExpiresAt holds the value of the "expires_at" field.
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	// CreatorKey holds the value of the "creator_key" field.
	CreatorKey   string `json:"creator_key,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*JsonStorage) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case jsonstorage.FieldID, jsonstorage.FieldData, jsonstorage.FieldCreatorKey:
			values[i] = new(sql.NullString)
		case jsonstorage.FieldExpiresAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the JsonStorage fields.
func (js *JsonStorage) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case jsonstorage.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				js.ID = value.String
			}
		case jsonstorage.FieldData:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field data", values[i])
			} else if value.Valid {
				js.Data = value.String
			}
		case jsonstorage.FieldExpiresAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field expires_at", values[i])
			} else if value.Valid {
				js.ExpiresAt = value.Time
			}
		case jsonstorage.FieldCreatorKey:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field creator_key", values[i])
			} else if value.Valid {
				js.CreatorKey = value.String
			}
		default:
			js.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the JsonStorage.
// This includes values selected through modifiers, order, etc.
func (js *JsonStorage) Value(name string) (ent.Value, error) {
	return js.selectValues.Get(name)
}

// Update returns a builder for updating this JsonStorage.
// Note that you need to call JsonStorage.Unwrap() before calling this method if this JsonStorage
// was returned from a transaction, and the transaction was committed or rolled back.
func (js *JsonStorage) Update() *JsonStorageUpdateOne {
	return NewJsonStorageClient(js.config).UpdateOne(js)
}

// Unwrap unwraps the JsonStorage entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (js *JsonStorage) Unwrap() *JsonStorage {
	_tx, ok := js.config.driver.(*txDriver)
	if !ok {
		panic("ent: JsonStorage is not a transactional entity")
	}
	js.config.driver = _tx.drv
	return js
}

// String implements the fmt.Stringer.
func (js *JsonStorage) String() string {
	var builder strings.Builder
	builder.WriteString("JsonStorage(")
	builder.WriteString(fmt.Sprintf("id=%v, ", js.ID))
	builder.WriteString("data=")
	builder.WriteString(js.Data)
	builder.WriteString(", ")
	builder.WriteString("expires_at=")
	builder.WriteString(js.ExpiresAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("creator_key=")
	builder.WriteString(js.CreatorKey)
	builder.WriteByte(')')
	return builder.String()
}

// JsonStorages is a parsable slice of JsonStorage.
type JsonStorages []*JsonStorage