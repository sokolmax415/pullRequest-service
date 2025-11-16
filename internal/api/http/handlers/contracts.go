package handler

import (
	"context"
	"pullrequest-service/internal/entity"
)

type UserUsecase interface {
	SetActiveFlag(ctx context.Context, userId string, isActive bool) (*entity.User, error)
	GetPR(ctx context.Context, userId string) ([]entity.PullRequestShort, error)
}

type TeamUsecase interface {
	AddTeam(ctx context.Context, team *entity.Team) error
	GetTeam(ctx context.Context, teamName string) (*entity.Team, error)
}

type PRUsecase interface {
	MergePR(ctx context.Context, prId string) (*entity.PullRequest, error)
	CreatePR(ctx context.Context, prId string, prName string, authorId string) (*entity.PullRequest, error)
	ReAssign(ctx context.Context, prId, oldReviewerId string) (*entity.PullRequest, error)
}
