package usecase

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"pullrequest-service/internal/entity"
)

type PRUsecase struct {
	prRep   PRRepository
	userRep UserRepository
	teamRep TeamRepository
	txMgr   TxManager
	logger  *slog.Logger
}

func NewPRUsecase(prRep PRRepository, userRep UserRepository, teamRep TeamRepository, txMgr TxManager, logger *slog.Logger) *PRUsecase {
	return &PRUsecase{prRep: prRep, userRep: userRep, teamRep: teamRep, txMgr: txMgr, logger: logger}
}

func (u *PRUsecase) MergePR(ctx context.Context, prId string) (*entity.PullRequest, error) {
	u.logger.Info("start merging PR", "pull_request_id", prId)

	if prId == "" {
		u.logger.Warn("invalid pull_request_id: empty", "pull_request_id", prId)
		return nil, entity.ErrInvalidRequest
	}

	open, err := u.prRep.IsPROpen(ctx, prId)

	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			u.logger.Warn("PR not found", "pull_request_id", prId, "error", err)
			return nil, err
		}

		u.logger.Error("failed to check PR status", "pull_request_id", prId, "error", err)

		return nil, entity.ErrInternalError
	}

	if open {
		err = u.prRep.MergePR(ctx, prId)
		if err != nil {
			u.logger.Error("failed to merge PR", "pull_request_id", prId, "error", err)
			return nil, entity.ErrInternalError
		}
	} else {
		u.logger.Info("PR is not OPEN, skipping merge", "pull_request_id", prId)
	}

	pr, err := u.prRep.GetPRById(ctx, prId)
	if err != nil {
		u.logger.Error("failed to get PR", "pull_request_id", prId, "error", err)
		return nil, entity.ErrInternalError
	}

	reviewers, err := u.prRep.GetReviewersIdForPR(ctx, prId)
	if err != nil {
		u.logger.Error("failed to get reviewers for PR", "pull_request_id", prId, "error", err)
		return nil, entity.ErrInternalError
	}

	pr.AssignedReviewers = reviewers

	u.logger.Info("successfully merged PR", "pull_request_id", prId)
	return pr, nil

}

func (u *PRUsecase) CreatePR(ctx context.Context, prId string, prName string, authorId string) (*entity.PullRequest, error) {
	u.logger.Info("start creating PR", "pull_request_id", prId, "pull_request_name", prName, "author_id", authorId)

	if prId == "" || prName == "" || authorId == "" {
		u.logger.Warn("invalid data: empty fields", "pull_request_id", prId, "pull_request_name", prName, "author_id", authorId)
		return nil, entity.ErrInvalidRequest
	}

	prShort := &entity.PullRequestShort{PullRequestID: prId, PullRequestName: prName, AuthorID: authorId, Status: entity.OPEN}
	createdPR := entity.PullRequest{PullRequestID: prId, PullRequestName: prName, AuthorID: authorId, Status: entity.OPEN}

	operation := func(ctx context.Context) error {
		exist, err := u.prRep.IsPRExist(ctx, prId)
		if err != nil {
			if !errors.Is(err, entity.ErrNotFound) {
				u.logger.Error("failed to check PR existence", "pull_request_id", prId, "error", err)
				return entity.ErrInternalError
			}
		}

		if exist {
			return entity.ErrPRExists
		}

		_, err = u.userRep.IsUserExist(ctx, authorId)
		if err != nil {
			if errors.Is(err, entity.ErrNotFound) {
				u.logger.Warn("author not found", "author_id", authorId, "error", err)
				return err
			}
			u.logger.Error("failed to check user existence", "author_id", authorId, "error", err)
			return entity.ErrInternalError
		}

		teamName, err := u.teamRep.GetTeamForUserId(ctx, authorId)
		if err != nil {
			if errors.Is(err, entity.ErrNotFound) {
				u.logger.Warn("team not found", "user_id", authorId, "error", err)
				return err
			}
			u.logger.Error("failed to get team", "user_id", authorId, "error", err)
			return entity.ErrInternalError
		}

		activeUsers, err := u.userRep.GetActiveUsersByTeam(ctx, *teamName)

		if err != nil {
			u.logger.Error("failed to get active users", "team_name", teamName, "error", err)
			return entity.ErrInternalError
		}

		candidates := make([]string, 0, len(activeUsers))

		for _, memberId := range activeUsers {
			if memberId != authorId {
				candidates = append(candidates, memberId)
			}
		}

		rand.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})

		reviewers := candidates
		if len(reviewers) > 2 {
			reviewers = reviewers[:2]
		}

		err = u.prRep.CreatePR(ctx, prShort)
		if err != nil {
			u.logger.Error("failed to create PR", "pull_request_id", prId, "pull_request_name", prName, "author_id", authorId, "error", err)
			return entity.ErrInternalError
		}

		for _, id := range reviewers {
			err := u.prRep.AddReviewerForPR(ctx, prId, id)
			if err != nil {
				u.logger.Error("failed to add reviewer to PR", "pull_request_id", prId, "reviewer_id", id, "error", err)
				return entity.ErrInternalError
			}
		}

		createdPR.AssignedReviewers = reviewers
		return nil

	}

	err := withRetry(ctx, func(ctx context.Context) error {
		return u.txMgr.WithTx(ctx, operation)
	}, 3)

	if err != nil {
		return nil, err
	}

	u.logger.Info("PR created successfully", "pull_request_id", prId, "pull_request_name", prName, "author_id", authorId)

	return &createdPR, nil

}
