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
	GetTeamNameByUserId(ctx context.Context, userId string) (*string, error)
	GetTeamByName(ctx context.Context, teamName string) (*entity.Team, error)
}

type UserRepository interface {
	AddUserToTeam(ctx context.Context, user *entity.User) error
	IsUserExist(ctx context.Context, userId string) (bool, error)
	SetActive(ctx context.Context, userId string, isActive bool) error
	GetActiveUsersByTeam(ctx context.Context, teamName string) ([]string, error)
	IsUserActive(ctx context.Context, userId string) (bool, error)
	GetUserById(ctx context.Context, userId string) (*entity.User, error)
}

type PRRepository interface {
	GetAllPRForReviewer(ctx context.Context, userId string) ([]entity.PullRequestShort, error)
	CreatePR(ctx context.Context, pr *entity.PullRequestShort) error
	AddReviewerForPR(ctx context.Context, prId string, userId string) error
	MergePR(ctx context.Context, prId string) error
	IsReviewerForPR(ctx context.Context, prId string, userId string) (bool, error)
	IsPROpen(ctx context.Context, prId string) (bool, error)
	DeleteReviewer(ctx context.Context, prId string, userId string) error
	GetPRById(ctx context.Context, prId string) (*entity.PullRequest, error)
	GetReviewersIdByPR(ctx context.Context, prId string) ([]string, error)
	IsPRExist(ctx context.Context, prId string) (bool, error)
}
