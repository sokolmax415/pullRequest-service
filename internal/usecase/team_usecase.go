package usecase

import (
	"context"
	"errors"
	"pullrequest-service/internal/entity"
	"time"

	"github.com/gookit/slog"
)

type TeamUsecase struct {
	teamRep TeamRepository
	userRep UserRepository
	txMgr   TxManager
	logger  *slog.Logger
}

func NewTeamUsecase(teamRep TeamRepository, userRep UserRepository, txMgr TxManager, logger *slog.Logger) *TeamUsecase {
	return &TeamUsecase{teamRep: teamRep, userRep: userRep, txMgr: txMgr, logger: logger}
}

func (u *TeamUsecase) AddTeam(ctx context.Context, team *entity.Team) error {
	u.logger.Info("adding team", "team_name", team.TeamName, "members_count", len(team.Members))

	if err := team.Validate(); err != nil {
		u.logger.Info("team validation failed", "team_name", team.TeamName, "error", err.Error())
		return err
	}

	operation := func(ctx context.Context) error {
		if err := u.teamRep.CreateNewTeam(ctx, team.TeamName); err != nil {
			if errors.Is(err, entity.ErrTeamExists) {
				u.logger.Warn("team already exsists", "team_name", team.TeamName, "error", err.Error())
				return err
			}

			u.logger.Error("failed to insert team", "team_name", team.TeamName, "error", err.Error())
			return entity.ErrInternalError
		}

		for _, member := range team.Members {
			teamForMember, err := u.teamRep.GetTeamForUserId(ctx, member.UserID)

			if err != nil {
				u.logger.Error("failed to get user's team", "user_id", member.UserID, "error", err.Error())
				return entity.ErrInternalError
			}

			if teamForMember != nil {
				u.logger.Warn("user is already in another team", "user_id", member.UserID)
				return entity.ErrUserInAnotherTeam
			}

			user := &entity.User{UserID: member.UserID, UserName: member.UserName, IsActive: member.IsActive, TeamName: team.TeamName}
			if err := u.userRep.AddUserToTeam(ctx, user); err != nil {
				u.logger.Error("failed to insert user", "user_id", member.UserID, "error", err.Error())
				return entity.ErrInternalError
			}
		}

		return nil
	}
	err := withRetry(ctx, func(txContext context.Context) error {
		return u.txMgr.WithTx(txContext, operation)
	}, 3)

	if err == nil {
		u.logger.Info("team created successfully", "team_name", team.TeamName)
	}

	return err
}

func (u *TeamUsecase) GetTeam(ctx context.Context, teamName string) (*entity.Team, error) {
	u.logger.Info("getting team", "team_name", teamName)

	team, err := u.teamRep.GetTeam(ctx, teamName)

	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			u.logger.Warn("team not found", "team_name", teamName, "error", err.Error())
			return nil, err
		}

		u.logger.Error("failed to get team", "team_name", teamName, "error", err.Error())
		return nil, entity.ErrInternalError
	}

	u.logger.Info("team is got", "team_name", teamName)

	return team, nil

}

func withRetry(ctx context.Context, fun func(context.Context) error, retryCount int) error {
	if fun == nil {
		return errors.New("fun operation is nil")
	}
	var err error
	for i := 0; i < int(retryCount); i++ {
		err = fun(ctx)
		if !errors.Is(err, entity.ErrSerializationFailure) {
			return err
		}

		time.Sleep(1 * time.Millisecond)
	}

	return err
}
