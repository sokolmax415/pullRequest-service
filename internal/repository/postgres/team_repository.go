package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"pullrequest-service/internal/entity"

	"github.com/Masterminds/squirrel"
)

type PostgresTeamRepository struct {
	db *sql.DB
	sq squirrel.StatementBuilderType
}

func NewPostgresTeamRepository(db *sql.DB) *PostgresTeamRepository {
	return &PostgresTeamRepository{db: db, sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}

}

func (r *PostgresTeamRepository) CreateNewTeam(ctx context.Context, teamName string) error {
	query, args, err := r.sq.Insert("teams").Columns("team_name").Values(teamName).ToSql()
	if err != nil {
		return fmt.Errorf("build insert team query: %w", err)
	}

	exec := executerFromContext(ctx, r.db)

	_, err = exec.ExecContext(ctx, query, args...)
	if err != nil {
		if isUniqueViolation(err) {
			return entity.ErrTeamExists
		}
		return fmt.Errorf("exec insert team: %w", err)
	}

	return nil
}

func (r *PostgresTeamRepository) GetTeamNameByUserId(ctx context.Context, userId string) (*string, error) {
	var currentTeam string
	query, args, err := r.sq.Select("team_name").From("users").Where(squirrel.Eq{"user_id": userId}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select team_name query: %w", err)
	}

	exec := executerFromContext(ctx, r.db)

	err = exec.QueryRowContext(ctx, query, args...).Scan(&currentTeam)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team: %w", entity.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to select team_name from users: %w", err)
	}

	return &currentTeam, nil
}

func (r *PostgresTeamRepository) GetTeamByName(ctx context.Context, teamName string) (*entity.Team, error) {
	query, args, err := r.sq.Select("user_id", "username", "is_active").From("users").Where(squirrel.Eq{"team_name": teamName}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build get team members query")
	}

	exec := executerFromContext(ctx, r.db)

	members := make([]entity.TeamMember, 0)

	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("exec get team member: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var member entity.TeamMember
		if err := rows.Scan(&member.UserID, &member.UserName, &member.IsActive); err != nil {
			return nil, fmt.Errorf("scan user row: %w", err)
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	if len(members) == 0 {
		return nil, fmt.Errorf("members: %w", entity.ErrNotFound)
	}

	team := &entity.Team{TeamName: teamName, Members: members}
	return team, nil

}
