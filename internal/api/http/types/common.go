package types

import (
	"encoding/json"
	"errors"
	"net/http"
	"pullrequest-service/internal/entity"
)

type ErrorResponse struct {
	Err ErrorDTO `json:"error"`
}

type ErrorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func HandleError(w http.ResponseWriter, err error) {
	resp := ErrorResponse{
		Err: ErrorDTO{
			Code:    entity.CodeInternal,
			Message: entity.ErrInternalError.Error(),
		},
	}

	status := http.StatusInternalServerError

	switch {
	case errors.Is(err, entity.ErrPRExists):
		status = http.StatusConflict
		resp.Err.Code = entity.CodePRExists
		resp.Err.Message = entity.ErrPRExists.Error()

	case errors.Is(err, entity.ErrNotFound):
		status = http.StatusNotFound
		resp.Err.Code = entity.CodeNotFound
		resp.Err.Message = entity.ErrNotFound.Error()

	case errors.Is(err, entity.ErrPRMerged):
		status = http.StatusConflict
		resp.Err.Code = entity.CodePRMerged
		resp.Err.Message = entity.ErrPRMerged.Error()

	case errors.Is(err, entity.ErrTeamExists):
		status = http.StatusConflict
		resp.Err.Code = entity.CodeTeamExists
		resp.Err.Message = entity.ErrTeamExists.Error()

	case errors.Is(err, entity.ErrUserInAnotherTeam):
		status = http.StatusConflict
		resp.Err.Code = entity.CodeUserInAnotherTeam
		resp.Err.Message = entity.ErrUserInAnotherTeam.Error()

	case errors.Is(err, entity.ErrNoCandidate):
		status = http.StatusConflict
		resp.Err.Code = entity.CodeNoCandidate
		resp.Err.Message = entity.ErrNoCandidate.Error()

	case errors.Is(err, entity.ErrNotAssigned):
		status = http.StatusBadRequest
		resp.Err.Code = entity.CodeNotAssigned
		resp.Err.Message = entity.ErrNotAssigned.Error()

	}
	WriteJSON(w, status, resp)
}
