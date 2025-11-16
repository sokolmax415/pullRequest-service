package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"pullrequest-service/internal/entity"

	"github.com/Masterminds/squirrel"
)

type PostgresPRRepository struct {
	db *sql.DB
	sq squirrel.StatementBuilderType
}

func NewPostgresPRRepository(db *sql.DB) *PostgresPRRepository {
	return &PostgresPRRepository{db: db, sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}

}

func (r *PostgresPRRepository) GetAllPRForReviewer(ctx context.Context, userId string) ([]entity.PullRequestShort, error) {
	query, args, err := r.sq.Select("pr.pull_request_id", "pr.pull_request_name", "pr.author_id", "pr.status").From("pull_requests pr").
		Join("pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id").Where(squirrel.Eq{"prr.user_id": userId}).ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build get PR author reviewer query: %w", err)
	}

	exec := executerFromContext(ctx, r.db)

	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("exec PR for reviewer: %w", err)
	}
	defer rows.Close()

	prList := make([]entity.PullRequestShort, 0)
	for rows.Next() {
		var pr entity.PullRequestShort

		if err = rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}
		prList = append(prList, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return prList, nil
}

func (r *PostgresPRRepository) CreatePR(ctx context.Context, pr *entity.PullRequestShort) error {
	query, args, err := r.sq.Insert("pull_requests").Columns("pull_request_id", "pull_request_name", "author_id", "status").
		Values(pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status).ToSql()

	if err != nil {
		return fmt.Errorf("failed to build insert PR: %w", err)
	}

	exec := executerFromContext(ctx, r.db)

	_, err = exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec insert PR: %w", err)
	}

	return nil
}

func (r *PostgresPRRepository) AddReviewerForPR(ctx context.Context, prId string, userId string) error {
	query, args, err := r.sq.Insert("pr_reviewers").Columns("pull_request_id", "user_id").
		Values(prId, userId).ToSql()

	if err != nil {
		return fmt.Errorf("failed to build insert pr_reviewer: %w", err)
	}

	exec := executerFromContext(ctx, r.db)

	_, err = exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec insert pr_reviewer: %w", err)
	}

	return nil
}

func (r *PostgresPRRepository) MergePR(ctx context.Context, prId string) error {
	query, args, err := r.sq.Update("pull_requests").Set("status", entity.MERGED).Set("merged_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"pull_request_id": prId}).ToSql()

	if err != nil {
		return fmt.Errorf("failed to build update PR status: %w", err)
	}

	exec := executerFromContext(ctx, r.db)

	_, err = exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec update PR status: %w", err)
	}

	return nil

}

func (r *PostgresPRRepository) IsReviewerForPR(ctx context.Context, prId string, userId string) (bool, error) {
	query, args, err := r.sq.Select("1").From("pull_requests pr").
		Join("pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id").
		Where(squirrel.Eq{"pr.pull_request_id": prId, "prr.user_id": userId}).ToSql()

	if err != nil {
		return false, fmt.Errorf("failed to build select reviewer exists query: %w", err)
	}

	exec := executerFromContext(ctx, r.db)

	var dummy string
	if err := exec.QueryRowContext(ctx, query, args...).Scan(&dummy); err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("user not found for PR: %w", entity.ErrNotAssigned)
		}
		return false, fmt.Errorf("exec select reviewer exists query: %w", err)
	}
	return true, nil

}

func (r *PostgresPRRepository) IsPROpen(ctx context.Context, prId string) (bool, error) {
	query, args, err := r.sq.Select("status").From("pull_requests").Where(squirrel.Eq{"pull_request_id": prId}).ToSql()

	if err != nil {
		return false, fmt.Errorf("failed to build select status for PR: %w", err)
	}

	exec := executerFromContext(ctx, r.db)

	var status string
	if err := exec.QueryRowContext(ctx, query, args...).Scan(&status); err != nil {
		if err == sql.ErrNoRows {
			return false, entity.ErrNotFound
		}
		return false, fmt.Errorf("failed to scan status row: %w", err)
	}

	return status == string(entity.OPEN), nil
}

func (r *PostgresPRRepository) DeleteReviewer(ctx context.Context, prId string, userId string) error {
	query, args, err := r.sq.Delete("pr_reviewers").Where(squirrel.Eq{"pull_request_id": prId, "user_id": userId}).ToSql()

	if err != nil {
		return fmt.Errorf("failed to build delete reviewer: %w", err)
	}

	exec := executerFromContext(ctx, r.db)

	_, err = exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to exec delete reviewer: %w", err)
	}

	return nil
}

func (r PostgresPRRepository) GetPRById(ctx context.Context, prId string) (*entity.PullRequest, error) {
	query, args, err := r.sq.Select("pull_request_id", "pull_request_name", "author_id", "status", "created_at", "merged_at").
		From("pull_requests").Where(squirrel.Eq{"pull_request_id": prId}).ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select PR: %w", err)
	}

	exec := executerFromContext(ctx, r.db)

	var PR entity.PullRequest
	if err := exec.QueryRowContext(ctx, query, args...).Scan(&PR.PullRequestID, &PR.PullRequestName, &PR.AuthorID, &PR.Status, &PR.CreatedAt, &PR.MergedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("PR not found: %w", entity.ErrNotFound)
		}
		return nil, fmt.Errorf("failed exec select PR: %w", err)
	}

	return &PR, nil
}

func (r *PostgresPRRepository) GetReviewersIdByPR(ctx context.Context, prId string) ([]string, error) {
	query, args, err := r.sq.Select("prr.user_id").From("pull_requests pr").
		Join("pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id").
		Where(squirrel.Eq{"pr.pull_request_id": prId}).ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select reviewers for PR: %w", err)
	}

	exec := executerFromContext(ctx, r.db)

	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to exec select reviewers for PR: %w", err)
	}
	defer rows.Close()

	reviewers := make([]string, 0)
	for rows.Next() {
		var reviewer string

		if err := rows.Scan(&reviewer); err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}
		reviewers = append(reviewers, reviewer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return reviewers, nil
}

func (r *PostgresPRRepository) IsPRExist(ctx context.Context, prId string) (bool, error) {
	query, args, err := r.sq.Select("1").From("pull_requests").Where(squirrel.Eq{"pull_request_id": prId}).ToSql()

	if err != nil {
		return false, fmt.Errorf("failed to build query: %w", err)
	}

	exec := executerFromContext(ctx, r.db)

	var dummy int
	err = exec.QueryRowContext(ctx, query, args...).Scan(&dummy)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("PR not found: %w", entity.ErrNotFound)
		}
		return false, fmt.Errorf("exec select user exists query: %w", err)
	}

	return true, nil
}
