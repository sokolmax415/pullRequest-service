package handler

import (
	"net/http"
	"pullrequest-service/internal/api/http/types"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	userUsecase UserUsecase
}

func NewUserHandler(userUsecase UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase}
}

func (h *UserHandler) SetActive(w http.ResponseWriter, r *http.Request) {
	req, err := types.ParseSetActiveRequest(r)
	if err != nil {
		types.HandleError(w, err)
		return
	}

	user, err := h.userUsecase.SetActiveFlag(r.Context(), req.UserId, req.IsActive)
	if err != nil {
		types.HandleError(w, err)
		return
	}

	resp := types.SetActiveResponseDTO{
		User: types.FromEntityUser(user),
	}

	types.WriteJSON(w, http.StatusCreated, resp)
}

func (h *UserHandler) GetPR(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "user_id")

	pr, err := h.userUsecase.GetPR(r.Context(), userId)
	if err != nil {
		types.HandleError(w, err)
		return
	}

	resp := types.UserDTO{
		UserID:       userId,
		PullRequests: types.FromEntityPRShort(pr),
	}

	types.WriteJSON(w, http.StatusCreated, resp)
}
