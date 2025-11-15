package usecase

import (
	"context"
	"pullrequest-service/internal/entity"
)

type TxManager interface {
	WithTx(context.Context, func(context.Context) error) error
}

type TeamRepository interface {
	CreateNewTeam(ctx context.Context, teamName string) error
	GetTeamForUserId(ctx context.Context, userId string) (*string, error)
	GetTeam(ctx context.Context, teamName string) (*entity.Team, error)
}

type UserRepository interface {
	AddUserToTeam(ctx context.Context, user *entity.User) error
}

type PRRepository interface {
}
