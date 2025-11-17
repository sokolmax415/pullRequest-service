package usecase

import (
	"context"
	"errors"
	"log/slog"
	"pullrequest-service/internal/entity"
)

type UserUsecase struct {
	userRep UserRepository
	prRep   PRRepository
	logger  *slog.Logger
}

func NewUserUsecase(userRep UserRepository, prRep PRRepository, logger *slog.Logger) *UserUsecase {
	return &UserUsecase{userRep: userRep, prRep: prRep, logger: logger}
}

func (u *UserUsecase) SetActiveFlag(ctx context.Context, userId string, isActive bool) (*entity.User, error) {
	u.logger.Info("start setting activity flag for user", "user_id", userId, "is_active", isActive)

	if userId == "" {
		u.logger.Warn("invalid user id: empty", "user_id", userId)
		return nil, entity.ErrInvalidRequest
	}

	_, err := u.userRep.IsUserExist(ctx, userId)

	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			u.logger.Warn("user not found", "user_id", userId)
			return nil, err
		}
		u.logger.Error("error checking user existence", "user_id", userId, "error", err)
		return nil, entity.ErrInternalError
	}

	err = u.userRep.SetActive(ctx, userId, isActive)
	if err != nil {
		u.logger.Error("failed to set user activity flag", "user_id", userId, "is_active", isActive, "error", err)
		return nil, entity.ErrInternalError
	}

	user, err := u.userRep.GetUserById(ctx, userId)

	if err != nil {
		u.logger.Error("failed to get user", "user_id", userId, "error", err)
		return nil, entity.ErrInternalError
	}

	u.logger.Info("successfully set activity flag for user", "user_id", userId, "is_active", isActive)

	return user, nil
}

func (u *UserUsecase) GetPR(ctx context.Context, userId string) ([]entity.PullRequestShort, error) {
	u.logger.Info("start get PR list for user", "user_id", userId)

	if userId == "" {
		u.logger.Warn("invalid user id: empty", "user_id", userId)
		return nil, entity.ErrInvalidRequest
	}

	_, err := u.userRep.IsUserExist(ctx, userId)

	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			u.logger.Warn("user not found", "user_id", userId)
			return nil, err
		}
		u.logger.Error("error checking user existence", "user_id", userId, "error", err)
		return nil, entity.ErrInternalError
	}

	prList, err := u.prRep.GetAllPRForReviewer(ctx, userId)

	if err != nil {
		u.logger.Error("failed to get PR list for user", "user_id", userId, "error", err)
	}

	u.logger.Info("successfully got PR list for user", "user_id", userId, "pr_count", len(prList))

	return prList, nil
}
