package postgres

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type Execer interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func isUniqueViolation(err error) bool {
	if err, ok := err.(*pq.Error); ok {
		return err.Code == "23505"
	}
	return false
}

func isSerializationFailure(err error) bool {
	if err, ok := err.(*pq.Error); ok {
		return err.Code == "40001"
	}
	return false
}

func executerFromContext(ctx context.Context, db *sql.DB) Execer {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok && tx != nil {
		return tx
	}
	return db
}
