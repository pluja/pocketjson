// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"pocketjson/ent/jsonstorage"
	"pocketjson/ent/predicate"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// JsonStorageDelete is the builder for deleting a JsonStorage entity.
type JsonStorageDelete struct {
	config
	hooks    []Hook
	mutation *JsonStorageMutation
}

// Where appends a list predicates to the JsonStorageDelete builder.
func (jsd *JsonStorageDelete) Where(ps ...predicate.JsonStorage) *JsonStorageDelete {
	jsd.mutation.Where(ps...)
	return jsd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (jsd *JsonStorageDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, jsd.sqlExec, jsd.mutation, jsd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (jsd *JsonStorageDelete) ExecX(ctx context.Context) int {
	n, err := jsd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (jsd *JsonStorageDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(jsonstorage.Table, sqlgraph.NewFieldSpec(jsonstorage.FieldID, field.TypeString))
	if ps := jsd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, jsd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	jsd.mutation.done = true
	return affected, err
}

// JsonStorageDeleteOne is the builder for deleting a single JsonStorage entity.
type JsonStorageDeleteOne struct {
	jsd *JsonStorageDelete
}

// Where appends a list predicates to the JsonStorageDelete builder.
func (jsdo *JsonStorageDeleteOne) Where(ps ...predicate.JsonStorage) *JsonStorageDeleteOne {
	jsdo.jsd.mutation.Where(ps...)
	return jsdo
}

// Exec executes the deletion query.
func (jsdo *JsonStorageDeleteOne) Exec(ctx context.Context) error {
	n, err := jsdo.jsd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{jsonstorage.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (jsdo *JsonStorageDeleteOne) ExecX(ctx context.Context) {
	if err := jsdo.Exec(ctx); err != nil {
		panic(err)
	}
}