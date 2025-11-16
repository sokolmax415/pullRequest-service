package types

import (
	"encoding/json"
	"net/http"
	"pullrequest-service/internal/entity"
)

type TeamRequestDTO struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDTO `json:"members"`
}

type TeamMemberDTO struct {
	UserID   string `json:"user_id"`
	UserName string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamResponseDTO struct {
	Team TeamRequestDTO `json:"team"`
}

func ParseTeamRequestDTO(r *http.Request) (*TeamRequestDTO, error) {
	var req TeamRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &req, nil
}

func FromEntityTeam(t *entity.Team) TeamRequestDTO {
	members := make([]TeamMemberDTO, len(t.Members))
	for i, m := range t.Members {
		members[i] = TeamMemberDTO{
			UserID:   m.UserID,
			UserName: m.UserName,
			IsActive: m.IsActive,
		}
	}

	return TeamRequestDTO{
		TeamName: t.TeamName,
		Members:  members,
	}
}
