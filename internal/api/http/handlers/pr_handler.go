package handler

import (
	"log/slog"
	"net/http"
	"pullrequest-service/internal/api/http/types"
)

type PRHandler struct {
	prUsecase PRUsecase
	logger    *slog.Logger
}

func NewPRHandler(prUsecase PRUsecase, logger *slog.Logger) *PRHandler {
	return &PRHandler{prUsecase: prUsecase, logger: logger}
}

func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	req, err := types.ParseCreatePrRequest(r)
	if err != nil {
		types.HandleError(w, err)
		return
	}

	pr, err := h.prUsecase.CreatePR(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		types.HandleError(w, err)
		return
	}

	resp := types.CreatePrResponse{
		PR: types.FromEntityPR(pr),
	}

	types.WriteJSON(w, http.StatusCreated, resp)
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	req, err := types.ParseMergeRequest(r)
	if err != nil {
		types.HandleError(w, err)
		return
	}

	pr, err := h.prUsecase.MergePR(r.Context(), req.PullRequestID)
	if err != nil {
		types.HandleError(w, err)
		return
	}

	resp := types.CreatePrResponse{
		PR: types.FromEntityPR(pr),
	}

	types.WriteJSON(w, http.StatusCreated, resp)
}

func (h *PRHandler) ReAssign(w http.ResponseWriter, r *http.Request) {
	req, err := types.ParseReAssignRequest(r)
	if err != nil {
		types.HandleError(w, err)
		return
	}

	pr, err := h.prUsecase.ReAssign(r.Context(), req.PullRequestID, req.OldReviewerId)
	if err != nil {
		types.HandleError(w, err)
		return
	}

	resp := types.ReAssignResponse{
		PR:          types.FromEntityPR(pr),
		OldReviewer: req.OldReviewerId,
	}

	types.WriteJSON(w, http.StatusCreated, resp)
}
