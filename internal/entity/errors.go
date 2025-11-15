package entity

import "errors"

var (
	ErrNotFound       = errors.New("resource not found")
	ErrInvalidRequest = errors.New("invalid request")

	ErrTeamExists        = errors.New("team_name already exists")
	ErrUserInAnotherTeam = errors.New("user already in another")

	ErrPRExists    = errors.New("PR id already exists")
	ErrPRMerged    = errors.New("cannot reassign on merged PR")
	ErrNotAssigned = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate = errors.New("no active replacement candidate in team")

	ErrSerializationFailure = errors.New("serialization failure")
	ErrInternalError        = errors.New("internal error")
)
