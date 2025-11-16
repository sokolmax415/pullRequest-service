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

	reviewers, err := u.prRep.GetReviewersIdByPR(ctx, prId)
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
			u.logger.Warn("PR exists", "pull_request_id", prId)
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

		teamName, err := u.teamRep.GetTeamNameByUserId(ctx, authorId)
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

		if err := u.prRep.CreatePR(ctx, prShort); err != nil {
			u.logger.Error("failed to create PR", "pull_request_id", prId, "pull_request_name", prName, "author_id", authorId, "error", err)
			return entity.ErrInternalError
		}

		for _, id := range reviewers {
			if err := u.prRep.AddReviewerForPR(ctx, prId, id); err != nil {
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

func (u *PRUsecase) ReAssign(ctx context.Context, prId, oldReviewerId string) (*entity.PullRequest, error) {
	u.logger.Info("start reassigning reviewer", "pull_request_id", prId, "old_reviewer_id", oldReviewerId)

	if prId == "" || oldReviewerId == "" {
		u.logger.Warn("invalid data: empty fields", "pull_request_id", prId, "old_reviewer_id", oldReviewerId)
		return nil, entity.ErrInvalidRequest
	}

	resultPR := &entity.PullRequest{}

	operation := func(ctx context.Context) error {
		_, err := u.prRep.IsPRExist(ctx, prId)
		if err != nil {
			if errors.Is(err, entity.ErrNotFound) {
				u.logger.Warn("PR not found", "pull_request_id", prId, "error", err)
				return err
			}
			u.logger.Error("failed to check PR existence", "pull_request_id", prId, "error", err)
			return entity.ErrInternalError
		}

		_, err = u.userRep.IsUserExist(ctx, oldReviewerId)
		if err != nil {
			if errors.Is(err, entity.ErrNotFound) {
				u.logger.Warn("user not found", "old_reviewer_id", oldReviewerId, "error", err)
				return err
			}

			u.logger.Error("failed to check user existence", "old_reviewer_id", oldReviewerId, "error", err)

			return entity.ErrInternalError
		}

		_, err = u.prRep.IsReviewerForPR(ctx, prId, oldReviewerId)
		if err != nil {
			if errors.Is(err, entity.ErrNotAssigned) {
				u.logger.Warn("Reviwer not assigned for PR", "old_reviewer_id", oldReviewerId, "error", err)
				return err
			}

			u.logger.Error("failed to check reviewer for PR", "pull_request_id", prId, "old_reviewer_id", oldReviewerId, "error", err)

			return entity.ErrInternalError
		}

		open, err := u.prRep.IsPROpen(ctx, prId)

		if err != nil {
			u.logger.Error("failed to check PR status", "pull_request_id", prId, "old_reviewer_id", oldReviewerId, "error", err)
			return entity.ErrInternalError
		}

		if !open {
			u.logger.Warn("failed to assign reviewer for MERGED PR", "pull_request_id", prId)
			return entity.ErrPRMerged
		}

		teamName, err := u.teamRep.GetTeamNameByUserId(ctx, oldReviewerId)
		if err != nil {
			if errors.Is(err, entity.ErrNotFound) {
				u.logger.Warn("team not found", "user_id", oldReviewerId, "error", err)
				return err
			}
			u.logger.Error("failed to get team", "user_id", oldReviewerId, "error", err)
			return entity.ErrInternalError
		}

		pr, err := u.prRep.GetPRById(ctx, prId)
		if err != nil {
			u.logger.Error("failed to get PR", "pull_request_id", prId, "error", err)
			return entity.ErrInternalError
		}

		activeUsers, err := u.userRep.GetActiveUsersByTeam(ctx, *teamName)

		if err != nil {
			u.logger.Error("failed to get active users", "team_name", teamName, "error", err)
			return entity.ErrInternalError
		}

		candidates := make([]string, 0, len(activeUsers))

		for _, memberId := range activeUsers {
			if memberId != oldReviewerId && memberId != pr.AuthorID {
				candidates = append(candidates, memberId)
			}
		}

		rand.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})

		if len(candidates) == 0 {
			u.logger.Warn("no available candidates for PR reviewers", "pull_request_id", prId)
			return entity.ErrNoCandidate
		}

		newReviewerId := candidates[0]

		if err = u.prRep.DeleteReviewer(ctx, prId, oldReviewerId); err != nil {
			u.logger.Error("failed to delete reviewer for PR", "pull_request_id", prId, "old_reviewer_id", oldReviewerId, "error", err)
			return entity.ErrInternalError
		}

		if err := u.prRep.AddReviewerForPR(ctx, prId, newReviewerId); err != nil {
			u.logger.Error("failed to add reviewer to PR", "pull_request_id", prId, "reviewer_id", newReviewerId, "error", err)
			return entity.ErrInternalError
		}

		reviewers, err := u.prRep.GetReviewersIdByPR(ctx, prId)
		if err != nil {
			u.logger.Error("failed to get reviewers", "pull_request_id", prId, "error", err)
			return entity.ErrInternalError
		}

		resultPR = &entity.PullRequest{
			PullRequestID:     pr.PullRequestID,
			PullRequestName:   pr.PullRequestName,
			AuthorID:          pr.AuthorID,
			Status:            pr.Status,
			AssignedReviewers: reviewers,
			CreatedAt:         pr.CreatedAt,
			MergedAt:          pr.MergedAt,
		}
		return nil
	}

	err := withRetry(ctx, func(ctx context.Context) error {
		return u.txMgr.WithTx(ctx, operation)
	}, 3)

	if err != nil {
		return nil, err
	}

	u.logger.Info("reassigning reviewer finished successfully", "pull_request_id", prId, "old_reviewer_id", oldReviewerId)

	return resultPR, nil

}
