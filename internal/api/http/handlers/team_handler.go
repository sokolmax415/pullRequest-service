package handler

import (
	"net/http"
	"pullrequest-service/internal/api/http/types"
	"pullrequest-service/internal/entity"
)

type TeamHandler struct {
	teamUsecase TeamUsecase
}

func NewTeamHandler(teamUsecase TeamUsecase) *TeamHandler {
	return &TeamHandler{teamUsecase: teamUsecase}
}

func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	req, err := types.ParseTeamRequestDTO(r)
	if err != nil {
		types.HandleError(w, err)
		return
	}

	team := &entity.Team{
		TeamName: req.TeamName,
		Members:  make([]entity.TeamMember, len(req.Members)),
	}

	for i, m := range req.Members {
		team.Members[i] = entity.TeamMember{
			UserID:   m.UserID,
			UserName: m.UserName,
			IsActive: m.IsActive,
		}
	}

	err = h.teamUsecase.AddTeam(r.Context(), team)
	if err != nil {
		types.HandleError(w, err)
		return
	}

	resp := types.TeamResponseDTO{Team: *req}

	types.WriteJSON(w, http.StatusCreated, resp)
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")

	team, err := h.teamUsecase.GetTeam(r.Context(), teamName)
	if err != nil {
		types.HandleError(w, err)
		return
	}

	resp := types.FromEntityTeam(team)

	types.WriteJSON(w, http.StatusCreated, resp)
}
