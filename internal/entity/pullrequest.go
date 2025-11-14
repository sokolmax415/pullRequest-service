package entity

import "time"

type Status string

const (
	MERGED Status = "MERGED"
	OPEN   Status = "OPEN"
)

type PullRequest struct {
	PullRequestID     string
	PullRequestName   string
	AuthorID          string
	Status            Status
	AssignedReviewers []string
	CreatedAt         *time.Time
	MergedAt          *time.Time
}

type PullRequestShort struct {
	PullRequestID   string
	PullRequestName string
	AuthorID        string
	Status          Status
}
