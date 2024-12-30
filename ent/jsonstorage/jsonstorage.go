// Code generated by ent, DO NOT EDIT.

package jsonstorage

import (
	"entgo.io/ent/dialect/sql"
)

const (
	// Label holds the string label denoting the jsonstorage type in the database.
	Label = "json_storage"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldData holds the string denoting the data field in the database.
	FieldData = "data"
	// FieldExpiresAt holds the string denoting the expires_at field in the database.
	FieldExpiresAt = "expires_at"
	// FieldCreatorKey holds the string denoting the creator_key field in the database.
	FieldCreatorKey = "creator_key"
	// Table holds the table name of the jsonstorage in the database.
	Table = "json_storages"
)

// Columns holds all SQL columns for jsonstorage fields.
var Columns = []string{
	FieldID,
	FieldData,
	FieldExpiresAt,
	FieldCreatorKey,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

var (
	// DefaultCreatorKey holds the default value on creation for the "creator_key" field.
	DefaultCreatorKey string
)

// OrderOption defines the ordering options for the JsonStorage queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByData orders the results by the data field.
func ByData(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldData, opts...).ToFunc()
}

// ByExpiresAt orders the results by the expires_at field.
func ByExpiresAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldExpiresAt, opts...).ToFunc()
}

// ByCreatorKey orders the results by the creator_key field.
func ByCreatorKey(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatorKey, opts...).ToFunc()
}