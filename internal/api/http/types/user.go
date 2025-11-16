package types

import (
	"encoding/json"
	"net/http"
	"pullrequest-service/internal/entity"
)

type SetActiveRequestDTO struct {
	UserId   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type SetActiveDTO struct {
	UserID   string `json:"user_id"`
	UserName string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type SetActiveResponseDTO struct {
	User SetActiveDTO `json:"user"`
}

type PullRequestShortDTO struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

type UserDTO struct {
	UserID       string                `json:"user_id"`
	PullRequests []PullRequestShortDTO `json:"pull_requests"`
}

func ParseSetActiveRequest(r *http.Request) (*SetActiveRequestDTO, error) {
	var req SetActiveRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &req, nil
}

func FromEntityUser(user *entity.User) SetActiveDTO {
	return SetActiveDTO{
		UserID:   user.UserID,
		UserName: user.UserName,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}
}

func FromEntityPRShort(pr []entity.PullRequestShort) []PullRequestShortDTO {
	var res []PullRequestShortDTO

	for _, r := range pr {
		res = append(res, PullRequestShortDTO{
			PullRequestID:   r.PullRequestID,
			PullRequestName: r.PullRequestName,
			AuthorID:        r.AuthorID,
			Status:          string(r.Status),
		})

	}
	return res
}
