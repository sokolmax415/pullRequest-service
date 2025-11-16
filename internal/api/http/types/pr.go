package types

import (
	"encoding/json"
	"net/http"
	"pullrequest-service/internal/entity"
	"time"
)

type CreatePrRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type PrDTO struct {
	PullRequestID     string        `json:"pull_request_id"`
	PullRequestName   string        `json:"pull_request_name"`
	AuthorID          string        `json:"author_id"`
	Status            entity.Status `json:"status"`
	AssignedReviewers []string      `json:"assigned_reviewers"`
	CreatedAt         *time.Time    `json:"created_at,omitempty"`
	MergedAt          *time.Time    `json:"merged_at,omitempty"`
}

type CreatePrResponse struct {
	PR PrDTO `json:"pr"`
}

type MergeRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type ReAssignRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldReviewerId string `json:"old_reviewer_id"`
}

type ReAssignResponse struct {
	PR          PrDTO  `json:"pr"`
	OldReviewer string `json:"replaced_by"`
}

func ParseCreatePrRequest(r *http.Request) (*CreatePrRequest, error) {
	var req CreatePrRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &req, nil
}

func ParseMergeRequest(r *http.Request) (*MergeRequest, error) {
	var req MergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &req, nil
}

func ParseReAssignRequest(r *http.Request) (*ReAssignRequest, error) {
	var req ReAssignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &req, nil
}

func FromEntityPR(pr *entity.PullRequest) PrDTO {
	return PrDTO{
		PullRequestID:     pr.PullRequestID,
		PullRequestName:   pr.PullRequestName,
		AuthorID:          pr.AuthorID,
		Status:            pr.Status,
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}
