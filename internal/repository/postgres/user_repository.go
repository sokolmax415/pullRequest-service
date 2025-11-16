package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"pullrequest-service/internal/entity"

	"github.com/Masterminds/squirrel"
)

type PostgresUserRepository struct {
	db *sql.DB
	sq squirrel.StatementBuilderType
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db, sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}

}

func (r *PostgresUserRepository) AddUserToTeam(ctx context.Context, user *entity.User) error {
	query, args, err := r.sq.Insert("users").Columns("user_id", "username", "is_active", "team_name").Values(user.UserID, user.UserName, user.IsActive, user.TeamName).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert user query")
	}

	exec := executerFromContext(ctx, r.db)
	_, err = exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec insert user: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) IsUserExist(ctx context.Context, userId string) (bool, error) {
	query, args, err := r.sq.Select("1").From("users").Where(squirrel.Eq{"user_id": userId}).ToSql()

	if err != nil {
		return false, fmt.Errorf("failed to build insert user query: %w", err)
	}

	exec := executerFromContext(ctx, r.db)

	var dummy int
	err = exec.QueryRowContext(ctx, query, args...).Scan(&dummy)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("exec select user exists query: %w", err)
	}

	return true, nil

}

func (r *PostgresUserRepository) SetActive(ctx context.Context, userId string, isActive string) error {
	query, args, err := r.sq.Update("users").Set("is_active", isActive).Where(squirrel.Eq{"user_id": userId}).ToSql()

	if err != nil {
		return fmt.Errorf("failed to build set user activity: %w", err)
	}

	exec := executerFromContext(ctx, r.db)

	_, err = exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec update user activity: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) GetActiveUsersByTeam(ctx context.Context, teamName string) ([]entity.User, error) {
	query, args, err := r.sq.Select("user_id", "username", "is_active", "team_name").From("users").
		Where(squirrel.Eq{"is_active": true, "team_name": teamName}).ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build get active users by team_name: %w", err)
	}

	exec := executerFromContext(ctx, r.db)

	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("exec select active users by team_name: %w", err)
	}
	defer rows.Close()

	usersList := make([]entity.User, 0)

	for rows.Next() {
		var activeUser entity.User
		if err := rows.Scan(&activeUser.UserID, &activeUser.UserName, &activeUser.IsActive, &activeUser.TeamName); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		usersList = append(usersList, activeUser)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return usersList, nil

}

func (r *PostgresUserRepository) IsUserActive(ctx context.Context, userId string) (bool, error) {
	query, args, err := r.sq.Select("is_active").From("users").Where(squirrel.Eq{"user_id": userId}).ToSql()

	if err != nil {
		return false, fmt.Errorf("failed to build get is_active for user: %w", err)
	}

	exec := executerFromContext(ctx, r.db)

	var isActive bool
	if err := exec.QueryRowContext(ctx, query, args...).Scan(&isActive); err != nil {
		if err == sql.ErrNoRows {
			return false, entity.ErrNotFound
		}

		return false, fmt.Errorf("failed to scan is_active row: %w", err)
	}

	return isActive, nil
}
