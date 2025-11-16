package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"pullrequest-service/internal/entity"
)

type TxManager struct {
	db *sql.DB
}

func NewTxManager(db *sql.DB) *TxManager {
	return &TxManager{db: db}
}

type txKey struct{}

func (m *TxManager) WithTx(ctx context.Context, fn func(context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	txCtx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(txCtx); err != nil {
		return fmt.Errorf("transaction function: %w", err)
	}

	if err := tx.Commit(); err != nil {
		if isSerializationFailure(err) {
			return entity.ErrSerializationFailure
		}
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
