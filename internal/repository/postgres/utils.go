package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type Execer interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func executerFromContext(ctx context.Context, db *sql.DB) (exec Execer) {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok && tx != nil {
		exec = tx
	} else {
		exec = db
	}

	return
}
